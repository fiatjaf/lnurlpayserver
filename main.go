package main

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
)

type Settings struct {
	Host        string `envconfig:"HOST" default:"0.0.0.0"`
	Port        string `envconfig:"PORT" required:"true"`
	ServiceURL  string `envconfig:"SERVICE_URL" required:"true"`
	PostgresURL string `envconfig:"DATABASE_URL" required:"true"`
	Secret      string `envconfig:"SECRET" required:"true"`
}

var err error
var s Settings
var pg *sqlx.DB
var log = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stderr})

func main() {
	err = envconfig.Process("", &s)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't process envconfig.")
	}

	// postgres connection
	pg, err = sqlx.Connect("postgres", s.PostgresURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to postgres")
	}

	// run check/cleanup tasks on start
	// and then every 30 minutes
	go func() {
		for {
			checkOldInvoices()
			cleanupInvoices()
			time.Sleep(30 * time.Minute)
		}
	}()

	// files
	indexhtml := MustAsset("public/index.html")

	// routers
	basemux := mux.NewRouter()

	staticmux := mux.NewRouter()
	staticmux.PathPrefix("/").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			asset, err := Asset(filepath.Join("public", r.URL.Path[1:]))
			mimetype := "text/html"
			if err != nil {
				asset = indexhtml
			} else {
				mimetype = mime.TypeByExtension(filepath.Ext(r.URL.Path))
			}

			w.Header().Add("Content-Type", mimetype)
			w.Write(asset)
		},
	)

	lnurlmux := mux.NewRouter()
	lnurlmux.Use(parseURLMiddleware)
	lnurlmux.PathPrefix("/lnurl/p/{shop}/{tpl}/").Methods("GET").HandlerFunc(lnurlPayParams)
	lnurlmux.PathPrefix("/lnurl/v/{shop}/{tpl}/").Methods("GET").HandlerFunc(lnurlPayValues)

	apimux := mux.NewRouter()
	apimux.Use(allJSONMiddleware)
	apimux.Use(authMiddleware)
	apimux.Path("/api/shop/{shop}").Methods("GET").HandlerFunc(getShop)
	apimux.Path("/api/shop/{shop}").Methods("PUT").HandlerFunc(setShop)
	apimux.Path("/api/shop/{shop}/templates").Methods("GET").HandlerFunc(listTemplates)
	apimux.Path("/api/shop/{shop}/template/{tpl}").Methods("PUT").HandlerFunc(setTemplate)
	apimux.Path("/api/shop/{shop}/template/{tpl}").Methods("DELETE").HandlerFunc(deleteTemplate)
	apimux.Path("/api/shop/{shop}/template/{tpl}").Methods("GET").HandlerFunc(getTemplate)
	apimux.Path("/api/shop/{shop}/template/{tpl}/lnurl").Methods("GET").HandlerFunc(getLNURL)
	apimux.Path("/api/shop/{shop}/invoices").Methods("GET").HandlerFunc(listInvoices)
	apimux.Path("/api/shop/{shop}/invoice/{hash}").Methods("GET").HandlerFunc(getInvoice)

	basemux.PathPrefix("/api/").Handler(apimux)
	basemux.PathPrefix("/lnurl/").Handler(lnurlmux)
	basemux.PathPrefix("/").Handler(staticmux)

	handler := cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept"},
		AllowCredentials: true,
	}).Handler(basemux)

	srv := &http.Server{
		Handler:      handler,
		Addr:         s.Host + ":" + s.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Debug().Str("addr", srv.Addr).Msg("listening")
	srv.ListenAndServe()
}
