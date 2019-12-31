package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func getSatoshisPer(currency string) (int64, error) {
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
			return int64(float64(100000000) / float64(price.(int64))), nil
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
	price, err := strconv.ParseInt(sprice, 10, 64)
	if err != nil {
		return 0, err
	}

	fiatPrices.MSet(map[string]interface{}{
		currency + ":price": price,
		currency + ":time":  now.Unix(),
	})

	return int64(float64(100000000) / float64(price)), nil
}

func paramsToInterface(params map[string]string) map[string]interface{} {
	res := make(map[string]interface{})
	for k, str := range params {
		var v interface{}

		err := json.Unmarshal([]byte(str), &v)
		if err != nil {
			v = str
		}

		res[k] = v
	}
	return res
}
