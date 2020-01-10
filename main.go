package main

import (
	"net/http"
	"os"
	"time"

	assetfs "github.com/elazarl/go-bindata-assetfs"
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
	assets := &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "/public/"}
	indexhtml := MustAsset("public/index.html")

	// routers
	basemux := mux.NewRouter()
	basemux.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(assets)))
	basemux.Path("/").Methods("GET").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/html")
			w.Write(indexhtml)
		},
	)

	lnurlmux := mux.NewRouter()
	lnurlmux.Use(parseURLMiddleware)
	lnurlmux.PathPrefix("/lnurl/p/{shop}/{tpl}/").Methods("GET").HandlerFunc(lnurlPayParams)
	lnurlmux.PathPrefix("/lnurl/v/{shop}/{tpl}/").Methods("GET").HandlerFunc(lnurlPayValues)

	apimux := mux.NewRouter()
	apimux.Use(allJSONMiddleware)
	apimux.Use(authMiddleware)
	apimux.Path("/shop/{shop}").Methods("GET").HandlerFunc(getShop)
	apimux.Path("/shop/{shop}").Methods("PUT").HandlerFunc(setShop)
	apimux.Path("/shop/{shop}/templates").Methods("GET").HandlerFunc(listTemplates)
	apimux.Path("/shop/{shop}/template/{tpl}").Methods("PUT").HandlerFunc(setTemplate)
	apimux.Path("/shop/{shop}/template/{tpl}").Methods("DELETE").HandlerFunc(deleteTemplate)
	apimux.Path("/shop/{shop}/template/{tpl}").Methods("GET").HandlerFunc(getTemplate)
	apimux.Path("/shop/{shop}/template/{tpl}/lnurl").Methods("GET").HandlerFunc(getLNURL)
	apimux.Path("/shop/{shop}/invoices").Methods("GET").HandlerFunc(listInvoices)
	apimux.Path("/shop/{shop}/invoice/{hash}").Methods("GET").HandlerFunc(getInvoice)

	basemux.PathPrefix("/shop/").Handler(apimux)
	basemux.PathPrefix("/lnurl/").Handler(lnurlmux)

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
