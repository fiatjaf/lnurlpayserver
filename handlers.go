package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fiatjaf/go-lnurl"
	"github.com/gorilla/mux"
)

type Response struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// middleware
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shopId := mux.Vars(r)["shop"]
		if _, key, ok := r.BasicAuth(); ok {
			var shop Shop
			err := pg.Get(&shop, `
              SELECT `+SHOPFIELDS+`
              FROM shop
              WHERE id = $1 AND key = $2
            `, shopId, key)
			if err == nil {
				r = r.WithContext(
					context.WithValue(
						r.Context(),
						"shop", shop,
					),
				)
				next.ServeHTTP(w, r)
				return
			}
		} else if r.Method == "PUT" && len(strings.Split(r.URL.Path, "/")) == 2 {
			// creating a shop
			next.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(401)
		json.NewEncoder(w).Encode(Response{
			false, "can't get shop " + shopId + " with given key."})
		return
	})
}

func allJSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
		return
	})
}

// handlers
func getShop(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(r.Context().Value("shop").(*Shop))
}

func setShop(w http.ResponseWriter, r *http.Request) {
	shopId := mux.Vars(r)["shop"]

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	// parse shop data
	var shop Shop
	err = json.Unmarshal(body, &shop)
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	// parse backend data
	var backend Backend
	err = json.Unmarshal(body, &backend)
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	txn, err := pg.Beginx()
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}
	defer txn.Rollback()

	existingShop, shopExists := r.Context().Value("shop").(*Shop)
	_, _, keyProvided := r.BasicAuth()
	shopDataProvided := len(shop.Verification) != 0
	backendDataProvided := backend.Kind != ""
	backendMatchesShop := false

	if !shopDataProvided {
		// this means we will probably only update the backend in an existing shop
		// we do this variable meddling now so things get normalized next
		shop = *existingShop
	}

	if backendDataProvided {
		err = backend.GetId()
		if err != nil {
			json.NewEncoder(w).Encode(
				Response{false, "failed to get node id: " + err.Error()})
			return
		}

		// always store the backend
		_, err = txn.Exec(`
          INSERT INTO backend (id, kind, connection)
          VALUES ($1, $2, $3)
          ON CONFLICT (id) DO UPDATE SET
            kind = $2,
            connection = $3
        `, backend.Id, backend.Kind, backend.Connection)
		if err != nil {
			log.Error().Err(err).Interface("backend", backend).
				Msg("invalid backend upsert")
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(Response{false, "missing data or key or backend"})
			return
		}

		if shopExists {
			backendMatchesShop = existingShop.Backend == backend.Id
		}

		// will always update backend (or leave unchanged)
		shop.Backend = backend.Id
	}

	shop.Id = shopId

	if shopExists && (keyProvided || backendMatchesShop) {
		// ok to update
	} else if !shopExists && shopDataProvided && backendDataProvided {
		// ok to create
	} else {
		// invalid operation
		log.Warn().
			Interface("shop", shop).
			Interface("backend", backend).
			Msg("invalid shop set action")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(Response{false, "missing data or key or backend"})
		return
	}

	// insert or update while also updating backend
	_, err = txn.Exec(`
      INSERT INTO shop
        (id, backend, message, verification, webhook, telegram)
      VALUES ($1, $2, $3, $4, $5, $6)
      ON CONFLICT (id) DO UPDATE SET
        backend = $2,
        message = $3,
        verification= $4,
        webhook = $5,
        telegram = $6
    `, shop.Id,
		shop.Backend,
		sql.NullString{String: shop.Message, Valid: shop.Message != ""},
		shop.Verification,
		sql.NullString{String: shop.Webhook, Valid: shop.Webhook != ""},
		sql.NullInt64{Int64: shop.Telegram, Valid: shop.Telegram != 0},
	)
	if err != nil {
		log.Error().Err(err).Interface("shop", shop).Msg("failed to upsert shop")
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	// key will be created automatically if shop is new
	if shop.Key != "" {
		err = txn.Get(&shop.Key, `SELECT key FROM shop WHERE id = $1`, shop.Id)
		if err != nil {
			log.Error().Err(err).Str("shop", shop.Id).Msg("got no key for shop")
			json.NewEncoder(w).Encode(Response{false, err.Error()})
			return
		}
	}

	err = txn.Commit()
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	// return the key
	json.NewEncoder(w).Encode(shop.Key)
}

func listTemplates(w http.ResponseWriter, r *http.Request) {
	shopId := mux.Vars(r)["shop"]

	var templates []Template
	err := pg.Select(&templates, `
      SELECT `+TEMPLATEFIELDS+`
      FROM template
      WHERE id = $1
    `, shopId)
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	json.NewEncoder(w).Encode(templates)
	return
}

func setTemplate(w http.ResponseWriter, r *http.Request) {
	shop := r.Context().Value("shop").(*Shop)
	tplId := mux.Vars(r)["tpl"]

	var t Template
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}
	t.Id = tplId
	t.Shop = shop.Id

	_, err = pg.Exec(`
          INSERT INTO template
            (id, shop, path_params, query_params, description, image,
             currency, min_price, max_price)
          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9
          ON CONFLICT DO UPDATE SET
            path_params = $3, query_params = $4,
            description = $5, image = $6,
            currency = $7, $min_price = $8, max_price = $9
        `, t.Id, t.Shop,
		t.PathParams, t.QueryParams,
		t.Description, sql.NullString{String: t.Image, Valid: t.Image != ""},
		t.Currency, t.MinPrice, t.MaxPrice,
	)
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	json.NewEncoder(w).Encode(Response{Ok: true})
	return
}

func deleteTemplate(w http.ResponseWriter, r *http.Request) {
	shop := r.Context().Value("shop").(*Shop)
	tplId := mux.Vars(r)["tpl"]

	_, err := pg.Exec(`
      DELETE FROM template WHERE id = $1 AND shop = $2
    `, tplId, shop)
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	json.NewEncoder(w).Encode(Response{Ok: true})
	return
}

func getTemplate(w http.ResponseWriter, r *http.Request) {
	shop := r.Context().Value("shop").(*Shop)
	tplId := mux.Vars(r)["tpl"]

	var template Template
	_, err := pg.Exec(`
      SELECT `+TEMPLATEFIELDS+` FROM template WHERE id = $1 AND shop = $2
    `, tplId, shop)
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	json.NewEncoder(w).Encode(template)
	return
}

func getLNURL(w http.ResponseWriter, r *http.Request) {
	shop := r.Context().Value("shop").(*Shop)
	tplId := mux.Vars(r)["tpl"]

	var template Template
	_, err := pg.Exec(`
      SELECT `+TEMPLATEFIELDS+` FROM template WHERE id = $1 AND shop = $2
    `, tplId, shop)
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	params := make(map[string]string)
	for k, v := range r.URL.Query() {
		params[k] = v[0]
	}

	url := template.MakeURL(params)
	lnurlEncoded, err := lnurl.LNURLEncode(url)
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	json.NewEncoder(w).Encode(lnurlEncoded)
}

func listInvoices(w http.ResponseWriter, r *http.Request) {
	shop := r.Context().Value("shop").(*Shop)

	var invoices []Invoice
	err := pg.Select(`
      SELECT `+INVOICEFIELDS+`
      FROM invoice
      WHERE invoice.shop = $1
    `, shop.Id)
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	json.NewEncoder(w).Encode(invoices)
}

func getInvoice(w http.ResponseWriter, r *http.Request) {
	shop := r.Context().Value("shop").(*Shop)
	hash := mux.Vars(r)["hash"]

	var invoice Invoice
	err := pg.Select(`
      SELECT `+INVOICEFIELDS+`
      FROM invoice
      WHERE invoice.hash = $1
        AND invoice.shop = $2
    `, hash, shop.Id)
	if err != nil {
		json.NewEncoder(w).Encode(Response{false, err.Error()})
		return
	}

	json.NewEncoder(w).Encode(invoice)
}
