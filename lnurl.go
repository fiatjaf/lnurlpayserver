package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/fiatjaf/go-lnurl"
	"github.com/gorilla/mux"
	"github.com/kr/pretty"
)

func parseURLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		shopId := vars["shop"]
		tplId := vars["tpl"]

		var t Template
		err = pg.Get(&t, `
          SELECT `+TEMPLATEFIELDS+` FROM template
          WHERE shop = $1 AND id = $2
        `, shopId, tplId)
		if err != nil {
			json.NewEncoder(w).Encode(lnurl.ErrorResponse("'" + tplId + "' not found on '" + shopId + "'."))
			return
		}

		params, err := t.ParseURL(r.URL)
		if err != nil {
			json.NewEncoder(w).Encode(lnurl.ErrorResponse("Failed to parse URL: " + err.Error()))
			return
		}

		r = r.WithContext(
			context.WithValue(
				context.WithValue(r.Context(),
					"template", &t,
				),
				"params", params,
			),
		)

		next.ServeHTTP(w, r)
	})
}

func lnurlPayParams(w http.ResponseWriter, r *http.Request) {
	t := r.Context().Value("template").(*Template)
	params := r.Context().Value("params").(map[string]string)

	log.Debug().Str("tpl", t.Id).Str("shop", t.Shop).Interface("params", params).
		Msg("lnurl-pay 1st call")

	min, max, err := t.GetPrices(params)
	if err != nil {
		json.NewEncoder(w).Encode(lnurl.ErrorResponse("Failed to calculate price: " + err.Error()))
		return
	}

	json.NewEncoder(w).Encode(lnurl.LNURLPayResponse1{
		Tag:             "payRequest",
		Callback:        strings.Replace(t.MakeURL(params), "/p/", "/v/", 1),
		EncodedMetadata: t.EncodedMetadata(params),
		MinSendable:     min,
		MaxSendable:     max,
	})
}

func lnurlPayValues(w http.ResponseWriter, r *http.Request) {
	t := r.Context().Value("template").(*Template)
	params := r.Context().Value("params").(map[string]string)
	amountStr := r.URL.Query().Get("amount")

	log.Debug().Str("tpl", t.Id).Str("shop", t.Shop).Interface("params", params).
		Str("amount", amountStr).
		Msg("lnurl-pay 2nd call")

	amount, _ := strconv.ParseInt(amountStr, 10, 64)
	invoice, err := t.MakeInvoice(amount, params)
	if err != nil {
		json.NewEncoder(w).Encode(lnurl.ErrorResponse("Failed to generate invoice: " + err.Error()))
		return
	}

	var shop Shop
	err = pg.Get(&shop, `
      SELECT `+SHOPFIELDS+` FROM shop
      WHERE id = $1
    `, t.Shop)
	if err != nil {
		json.NewEncoder(w).Encode(lnurl.ErrorResponse("Couldn't get shop: " + err.Error()))
		return
	}

	sa, err := shop.MakeSuccessAction(params, invoice.Preimage)
	if err != nil {
		json.NewEncoder(w).Encode(lnurl.ErrorResponse("SuccessAction error: " + err.Error()))
		return
	}

	pretty.Log(invoice.Bolt11)
	pretty.Log(sa)

	r.Header.Set("X-Invoice-Id", invoice.Hash)
	json.NewEncoder(w).Encode(lnurl.LNURLPayResponse2{
		Routes:        make([][]lnurl.RouteInfo, 0),
		PR:            invoice.Bolt11,
		SuccessAction: sa,
	})
}
