package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	lightning "github.com/fiatjaf/lightningd-gjson-rpc"
	decodepay "github.com/fiatjaf/ln-decodepay"
	"github.com/jmoiron/sqlx/types"
	"github.com/lucsky/cuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type Backend struct {
	Id         string         `db:"id" json:"id"`
	Kind       string         `db:"kind" json:"kind"`
	Connection types.JSONText `db:"connection" json:"connection"`
}

func (b Backend) Conn() gjson.Result {
	return gjson.ParseBytes(b.Connection)
}

func (b *Backend) GetId() error {
	useless := make([]byte, 32)
	rand.Read(useless)
	nothing := sha256.Sum256([]byte{0})
	bolt11, err := b.MakeInvoice(1000, nothing, useless, 600)
	if err != nil {
		return err
	}

	inv, err := decodepay.Decodepay(bolt11)
	if err != nil {
		return err
	}

	b.Id = inv.Payee
	return nil
}

func (b Backend) MakeInvoice(msatoshi int64, h [32]byte, preimage []byte, expiry int) (bolt11 string, err error) {
	defer func(prevTransport http.RoundTripper) {
		http.DefaultClient.Transport = prevTransport
	}(http.DefaultClient.Transport)

	conn := b.Conn()

	if conn.Get("cert").Exists() {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(conn.Get("cert").String()))

		http.DefaultClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: caCertPool},
		}
	} else {
		http.DefaultClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	hexh := hex.EncodeToString(h[:])
	b64h := base64.StdEncoding.EncodeToString(h[:])

	switch b.Kind {
	case "spark":
		spark := &lightning.Client{
			SparkURL:    conn.Get("endpoint").String(),
			SparkToken:  conn.Get("key").String(),
			CallTimeout: time.Second * 3,
		}
		inv, err := spark.CallNamed("lnurlinvoice",
			"msatoshi", msatoshi,
			"label", "lnurlpayserver/"+cuid.Slug(),
			"description_hash", hexh,
			"expiry", expiry,
			"preimage", hex.EncodeToString(preimage),
		)
		if err != nil {
			return "", fmt.Errorf("lnurlinvoice call failed: %w", err)
		}
		return inv.Get("bolt11").String(), nil

	case "lnd":
		body, _ := sjson.Set("{}", "description_hash", b64h)
		body, _ = sjson.Set(body, "value", msatoshi/1000)
		body, _ = sjson.Set(body, "preimage", base64.StdEncoding.EncodeToString(preimage))
		body, _ = sjson.Set(body, "expiry", strconv.Itoa(expiry))

		req, err := http.NewRequest("POST",
			conn.Get("endpoint").String()+"/v1/invoices",
			bytes.NewBufferString(body),
		)
		if err != nil {
			return "", err
		}
		req.Header.Set("Grpc-Metadata-macaroon", conn.Get("macaroon").String())
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		if resp.StatusCode >= 300 {
			return "", errors.New("call to lnd failed")
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		return gjson.ParseBytes(b).Get("payment_request").String(), nil
	}

	return "", errors.New("unsupported lightning server kind: " + b.Kind)
}

func (backend Backend) waitInvoicePaid(hash string) bool {
	stop := make(chan bool, 1)
	go func() {
		<-time.After(30 * time.Minute)
		stop <- true
	}()

	select {
	case <-stop:
		return false
	case <-backend.waitInvoice(hash):
		return true
	}
}

func (backend *Backend) waitInvoice(hash string) <-chan bool {
	res := make(chan bool, 1)

	conn := backend.Conn()

	logger := log.With().Str("hash", hash).Str("backend", backend.Kind).
		Str("conn", conn.String()).Logger()

	go func() {
		defer func(prevTransport http.RoundTripper) {
			http.DefaultClient.Transport = prevTransport
		}(http.DefaultClient.Transport)

		if conn.Get("cert").Exists() {
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM([]byte(conn.Get("cert").String()))

			http.DefaultClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{RootCAs: caCertPool},
			}
		} else {
			http.DefaultClient.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}

		for {
			switch backend.Kind {
			case "spark":
				spark := &lightning.Client{
					SparkURL:    conn.Get("endpoint").String(),
					SparkToken:  conn.Get("key").String(),
					CallTimeout: time.Minute * 15,
				}
				_, err := spark.Call("waitinvoice", hash)
				if err != nil {
					logger.Warn().Err(err).
						Msg("error on spark waitinvoice")
					continue
				}

				res <- true
				return
			case "lnd":
				endpoint := conn.Get("endpoint").String()
				macaroon := conn.Get("macaroon").String()

				// get the add_index for this invoice
				req, err := http.NewRequest("GET",
					endpoint+"/v1/invoice/"+hash, nil)
				if err != nil {
					logger.Warn().Err(err).Msg("error preparing lnd request")
					res <- false
					return
				}
				req.Header.Set("Grpc-Metadata-macaroon", macaroon)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					logger.Warn().Err(err).Msg("error on lnd invoice/")
					res <- false
					return
				}
				if resp.StatusCode >= 300 {
					logger.Warn().Err(err).Msg("error on lnd invoice/")
					res <- false
					return
				}
				defer resp.Body.Close()
				b, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					logger.Warn().Err(err).Msg("error reading lnd response")
					res <- false
					return
				}
				invdata := gjson.ParseBytes(b)
				if invdata.Get("settled").Bool() {
					// paid already, stop here
					res <- true
					return
				}
				addIndex := invdata.Get("add_index").String()

				// now that we have the add_index we can listen to lnd's stream
				req, err = http.NewRequest("GET",
					endpoint+"/v1/invoices/subscribe?add_index="+addIndex, nil)
				if err != nil {
					logger.Warn().Err(err).Msg("error preparing lnd request")
					return
				}
				req.Header.Set("Grpc-Metadata-macaroon", macaroon)
				resp, err = http.DefaultClient.Do(req)
				if err != nil {
					logger.Warn().Err(err).Msg("error on lnd invoices/subscribe")
					res <- false
					return
				}
				if resp.StatusCode >= 300 {
					logger.Warn().Err(err).Msg("error on lnd invoices/subscribe")
					res <- false
					return
				}

				defer resp.Body.Close()
				var settled struct {
					Settled bool `json:"settled"`
				}
				decoder := json.NewDecoder(resp.Body)
				err = decoder.Decode(&settled)
				if err != nil || !settled.Settled {
					logger.Warn().Err(err).
						Msg("error on lnd subscribe")
					continue
				}

				res <- true
				return
			}
		}
	}()

	return res
}
