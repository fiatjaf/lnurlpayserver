package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hoisie/mustache"
)

type Template struct {
	Id          string               `db:"id" json:"id"`
	Shop        string               `db:"shop" json:"shop"`
	PathParams  DelimitedStringArray `db:"path_params" json:"path_params"`
	QueryParams DelimitedStringArray `db:"query_params" json:"query_params"`
	Description string               `db:"description" json:"description"`
	Image       string               `db:"image" json:"image,omitempty"`
	Currency    string               `db:"currency" json:"currency"`
	MinPrice    string               `db:"min_price" json:"min_price"`
	MaxPrice    string               `db:"max_price" json:"max_price"`
}

var TEMPLATEFIELDS = `id, shop, array_to_string(path_params, '|') AS path_params, array_to_string(query_params, '|') AS query_params, description, coalesce(image, '') AS image, currency, min_price, max_price`

func (t *Template) MakeURL(params map[string]string) string {
	path := "/lnurl/p/" + t.Shop + "/" + t.Id + "/"

	// add path params
	ppath := make([]string, len(t.PathParams))
	for i, key := range t.PathParams {
		value, _ := params[key]
		ppath[i] = fmt.Sprint(value)
	}
	if len(ppath) > 0 {
		path += strings.Join(ppath, "/")
	}

	// add querystring params
	qs := url.Values{}
	for _, key := range t.QueryParams {
		if value, ok := params[key]; ok {
			qs.Set(key, fmt.Sprint(value))
		}
	}

	// add hmac
	mac := hmac.New(sha256.New, []byte(s.Secret))
	mac.Write([]byte(path[8:]))
	qs.Set("hmac", hex.EncodeToString(mac.Sum(nil)))

	return s.ServiceURL + path + "?" + qs.Encode()
}

func (t *Template) ParseURL(u *url.URL) (params map[string]string, err error) {
	if !strings.HasPrefix(u.Path, "/") {
		u.Path = "/" + u.Path
	}

	qs := u.Query()
	spl := strings.Split(u.Path, "/")
	if len(spl) < 5 {
		err = fmt.Errorf("invalid path: %s", u.Path)
		return
	}

	// get params from URL path
	params = make(map[string]string)
	for i, paramName := range t.PathParams {
		value := spl[5+i]
		params[paramName] = value
	}

	// verify path hmac
	code, _ := hex.DecodeString(qs.Get("hmac"))
	mac := hmac.New(sha256.New, []byte(s.Secret))
	mac.Write([]byte(u.Path[8:]))
	smac := mac.Sum(nil)
	if !hmac.Equal(code, smac) {
		err = errors.New("Invalid lnurl: HMAC doesn't match.")
		return
	}
	qs.Del("hmac")

	// get params from querystring
	for _, paramName := range t.QueryParams {
		if values, ok := qs[paramName]; ok {
			params[paramName] = values[0]
		}
	}

	return
}

func (t Template) MakeInvoice(
	amount int64,
	params map[string]string,
) (invoice *Invoice, err error) {
	// validate amount
	min, max, err := t.GetPrices(params)
	if err != nil {
		return nil, fmt.Errorf("error getting prices: %w", err)
	}
	if amount > max || amount < min {
		return nil, fmt.Errorf("Invalid amount: %d", amount)
	}

	// get metadata as string
	encodedMetadata := t.EncodedMetadata(params)

	// generate invoice and save invoice object
	inv, err := NewInvoice(t.Id, t.Shop, amount, params, encodedMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to make invoice: %w", err)
	}

	return inv, nil
}

func (t *Template) GetPrices(params map[string]string) (min int64, max int64, err error) {
	names, values := paramsToJQVars(params)

	// calculate raw prices
	fmin, err1 := runJQPrice(t.MinPrice, names, values)
	fmax, err2 := runJQPrice(t.MaxPrice, names, values)
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("min: %w, max: %w", err1, err2)
	}

	// convert to satoshis
	var satoshis float64 = 1
	if t.Currency != "sat" {
		satoshis, err = getSatoshisPer(t.Currency)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get %s price: %w", t.Currency, err)
		}
	}

	return int64(fmin * satoshis * 1000), int64(fmax * satoshis * 1000), nil
}

func (t *Template) EncodedMetadata(params map[string]string) string {
	kv := make([][]string, 1, 2)

	description := mustache.Render(t.Description, params)
	kv[0] = []string{"text/plain", description}

	if t.Image != "" {
		// should be in format 'data:image/png;base64,...' (or jpeg)
		spl := strings.Split(t.Image[5:], ",")
		mime := spl[0]
		content := spl[1]
		kv = append(kv, []string{mime, content})
	}

	j, _ := json.Marshal(kv)
	return string(j)
}
