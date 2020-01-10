package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx/types"
)

func NewInvoice(
	templateId string,
	shopId string,
	price int64,
	params map[string]string,
	encodedMetadata string,
) (*Invoice, error) {
	metadataHash := sha256.Sum256([]byte(encodedMetadata))
	expirySeconds := 1800 // 30 minutes
	preimage := make([]byte, 32)
	if _, err = io.ReadFull(rand.Reader, preimage); err != nil {
		return nil, err
	}
	preimageStr := hex.EncodeToString(preimage)
	hash := sha256.Sum256(preimage)
	hashStr := hex.EncodeToString(hash[:])

	backend, err := BackendFromShop(shopId)
	if err != nil {
		return nil, fmt.Errorf("failed to get backend info to generate invoice: %w", err)
	}

	bolt11, err := backend.MakeInvoice(price, metadataHash, preimage, expirySeconds)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice: %w", err)
	}

	jparams, _ := json.Marshal(params)
	var inv Invoice
	err = pg.Get(&inv, `
      INSERT INTO invoice
        (preimage, hash, shop, template, params, amount_msat, bolt11)
      VALUES ($1, $2, $3, $4, $5, $6, $7)
      RETURNING `+INVOICEFIELDS+`
    `, preimageStr, hashStr, shopId, templateId, jparams, price, bolt11)
	if err != nil {
		return nil, fmt.Errorf("failed to save invoice on database: %w", err)
	}

	return &inv, nil
}

type Invoice struct {
	Hash       string         `db:"hash" json:"hash"`
	Preimage   string         `db:"preimage" json:"preimage"`
	Shop       string         `db:"shop" json:"shop"`
	Template   string         `db:"template" json:"template"`
	Params     types.JSONText `db:"params" json:"params"`
	AmountMsat int64          `db:"amount_msat" json:"amount_msat"`
	Bolt11     string         `db:"bolt11" json:"bolt11"`
	Creation   time.Time      `db:"creation" json:"creation"`
	Payment    *time.Time     `db:"payment" json:"payment"`

	backend *Backend
}

const INVOICEFIELDS = `hash, preimage, template, shop, params, amount_msat, bolt11, creation, payment`

func (inv Invoice) Wait() {
	if inv.backend == nil {
		backend, err := BackendFromShop(inv.Shop)
		if err != nil {
			log.Error().Err(err).Msg("failed to get backend from invoice")
			return
		}
		inv.backend = backend
	}

	paid := inv.backend.waitInvoicePaid(inv.Hash)
	if !paid {
		log.Debug().Interface("invoice", inv).
			Msg("waited, but invoice wasn't paid")
		return
	}

	inv.markAsPaid()
	inv.sendWebhook()
}

func (inv Invoice) Check() {
	if inv.backend == nil {
		backend, err := BackendFromShop(inv.Shop)
		if err != nil {
			log.Error().Err(err).Msg("failed to get backend from invoice")
			return
		}
		inv.backend = backend
	}

	paid := inv.backend.checkInvoice(inv.Hash)
	if paid {
		inv.markAsPaid()
		inv.sendWebhook()
	}
}

func (inv Invoice) markAsPaid() {
	_, err := pg.Exec(`
      UPDATE invoice
      SET payment = now()
      WHERE hash = $1
    `, inv.Hash)
	if err != nil {
		log.Error().Err(err).Interface("invoice", inv).
			Msg("failed to mark invoice as paid")
		return
	}
}

func (inv Invoice) sendWebhook() {
	var webhook string
	err = pg.Get(&webhook, `
      SELECT webhook
      FROM shop
      WHERE shop.id = $1
    `, inv.Shop)
	if err == nil {
		jinv, _ := json.Marshal(inv)
		_, err := http.Post(webhook, "application/json", bytes.NewBuffer(jinv))
		if err != nil {
			log.Warn().Err(err).Str("url", webhook).Msg("webhook error")
		} else {
			log.Info().Str("url", webhook).Msg("webhook dispatched")
		}
	}
}
