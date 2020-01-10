package main

func cleanupInvoices() {
	_, err := pg.Exec(`
      DELETE FROM invoice
      WHERE payment IS NULL
        AND creation < now() - interval '1 hour'
    `)
	if err != nil {
		log.Error().Err(err).Msg("error cleaning up invoices")
	}
}

func checkOldInvoices() {
	var invoices []Invoice
	err := pg.Select(&invoices, `
      SELECT `+INVOICEFIELDS+` FROM invoice
      WHERE payment IS NULL
    `)
	if err != nil {
		log.Error().Err(err).Msg("error checking old invoices")
		return
	}

	for _, inv := range invoices {
		log.Debug().Str("bolt11", inv.Bolt11).Msg("checking old invoice")
		go inv.Check()
	}
}
