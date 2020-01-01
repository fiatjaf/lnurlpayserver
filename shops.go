package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/fiatjaf/go-lnurl"
	"github.com/hoisie/mustache"
	"github.com/jmoiron/sqlx/types"
	"github.com/tidwall/gjson"
)

type Shop struct {
	Id           string         `db:"id" json:"id"`
	Backend      string         `db:"backend" json:"backend"`
	Key          string         `db:"key" json:"key"`
	Message      string         `db:"message" json:"message,omitempty"`
	Verification types.JSONText `db:"verification" json:"verification"`
	Webhook      string         `db:"webhook" json:"webhook"`
}

var SHOPFIELDS = `id, backend, key, coalesce(message, '') AS message, verification, coalesce(webhook, '') AS webhook`

func (shop *Shop) MakeSuccessAction(
	params map[string]string,
	invoice string,
) (sa *lnurl.SuccessAction, err error) {
	key, err := hex.DecodeString(invoice)
	if err != nil {
		return
	}
	message := ""
	if shop.Message != "" {
		message = mustache.Render(shop.Message, params)
	}

	v := gjson.ParseBytes(shop.Verification)

	switch v.Get("kind").String() {
	case "none":
		if message != "" {
			return lnurl.Action(message, ""), nil
		} else {
			return nil, nil
		}
	case "sequential":
		// count all invoices paid today
		var seq int
		err = pg.Get(&seq, `
          SELECT count(*) FROM invoice
          WHERE invoice.shop = $1
            AND payment > current_date
        `, shop.Id)
		seq += int(v.Get("init").Int())

		code := fmt.Sprintf("%d", seq)
		if v.Get("words").Exists() {
			words := v.Get("words").Array()
			code = words[seq%len(words)].String()
		}

		return lnurl.AESAction(message, key, code)
	case "hmac":
		// produce a secret code valid for this shop every x minutes
		currentRange := time.Now().Unix() / (60 * v.Get("interval").Int())
		h := hmac.New(sha256.New, []byte(v.Get("key").String()))
		h.Write([]byte(strconv.FormatInt(currentRange, 10)))
		code := base64.StdEncoding.EncodeToString(h.Sum(nil))[:6]
		return lnurl.AESAction(message, key, code)
	default:
		return nil, errors.New("invalid success action type")
	}
}
