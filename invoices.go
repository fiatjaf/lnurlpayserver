package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/jmoiron/sqlx/types"
)

func NewInvoice(
	templateId string,
	shopId string,
	price int64,
	params map[string]string,
	encodedMetadata string,
) (invoice *Invoice, err error) {
	metadataHash := sha256.Sum256([]byte(encodedMetadata))
	expirySeconds := 1800 // 30 minutes
	preimage := make([]byte, 32)
	if _, err = io.ReadFull(rand.Reader, preimage); err != nil {
		return nil, err
	}
	preimageStr := hex.EncodeToString(preimage)
	hash := sha256.Sum256(preimage)
	hashStr := hex.EncodeToString(hash[:])

	var backend Backend
	err = pg.Get(&backend, `
      SELECT backend.* FROM backend
      INNER JOIN shop ON shop.backend = backend.id
      WHERE shop.id = $1
    `, shopId)
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
	Template   string         `db:"template" json:"template"`
	Params     types.JSONText `db:"params" json:"params"`
	AmountMsat int64          `db:"amount_msat" json:"amount_msat"`
	Bolt11     string         `db:"bolt11" json:"bolt11"`
	Creation   time.Time      `db:"creation" json:"creation"`
	Payment    *time.Time     `db:"payment" json:"payment"`
}

const INVOICEFIELDS = `hash, preimage, template, params, amount_msat, bolt11, creation, payment`
