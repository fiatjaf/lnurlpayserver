package main

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/itchyny/gojq"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/tidwall/gjson"
)

type DelimitedStringArray []string

func (ss *DelimitedStringArray) Scan(src interface{}) error {
	if v, ok := src.(string); ok {
		*ss = strings.Split(v, "|")
		if len(*ss) == 1 && []string(*ss)[0] == "" {
			*ss = make(DelimitedStringArray, 0)
		}
		return nil
	} else {
		return errors.New("not a |-delimited string array")
	}
}

func (ss DelimitedStringArray) Value() (driver.Value, error) {
	return strings.Join(ss, "|"), nil
}

var fiatPrices = cmap.New()

func getSatoshisPer(currency string) (float64, error) {
	now := time.Now()

	// first check cache
	if fiatPrices.Has(currency + ":price") {
		if since, _ := fiatPrices.Get(currency + ":time"); since.(int64) < now.Add(-time.Minute*15).Unix() {
			// delete old
			fiatPrices.Remove(currency + ":price")
			fiatPrices.Remove(currency + ":time")
		} else {
			// use this
			price, _ := fiatPrices.Get(currency + ":price")
			return float64(100000000) / price.(float64), nil
		}
	}

	// otherwise proceed to fetch prices
	cur := strings.ToUpper(currency)

	resp, err := http.Get("https://api.kraken.com/0/public/Ticker?pair=XBT" + cur)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	sprice := gjson.ParseBytes(b).Get("result.XXBTZ" + cur + ".c.0").String()
	price, err := strconv.ParseFloat(sprice, 64)
	if err != nil {
		return 0, err
	}

	fiatPrices.MSet(map[string]interface{}{
		currency + ":price": price,
		currency + ":time":  now.Unix(),
	})

	return float64(100000000) / price, nil
}

func paramsToJQVars(params map[string]string) (names []string, values []interface{}) {
	for k, str := range params {
		var v interface{}

		err := json.Unmarshal([]byte(str), &v)
		if err != nil {
			v = str
		}

		names = append(names, k)
		values = append(values, v)
	}
	return
}

func runJQPrice(
	code string,
	names []string,
	values []interface{},
) (res float64, err error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0, nil
	}

	query, err := gojq.Parse(code)
	if err != nil {
		return
	}

	program, err := gojq.Compile(query, names...)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	iter := program.RunWithContext(ctx, nil, values...)
	v, ok := iter.Next()
	if !ok {
		return 0, errors.New("nothing returned")
	}
	if err, ok := v.(error); ok {
		return 0, err
	}

	if price, ok := v.(float64); !ok {
		return 0, errors.New("result is not a number")
	} else {
		return price, nil
	}
}
