/** @format */

import {h, Component} from 'preact'
import {route} from 'preact-router'
import {idb} from '../../idb'

export default class CreateShop extends Component {
  state = {
    shopID: null,
    shopKey: parseInt(Math.random() * 1000000).toString(),
    shopVerif: 'none',
    shopSeqWords: null,
    shopSeqStart: 1,
    shopHmacKey: null,
    shopHmacInt: 10,
    shopThankYouMsg: '',
    shopBackEnd: 'lnd',
    shopEndpoint: null,
    shopMacaroon: null,
    shopInvoiceKey: null,
    shopTLS: null,
    loading: true
  }

  handleInputChange = event => {
    const target = event.target
    const value = target.type === 'checkbox' ? target.checked : target.value
    const name = target.name

    this.setState({
      [name]: value
    })

    if (name === 'shopBackEnd') {
      this.setState({
        shopEndpoint: null,
        shopInvoiceKey: null,
        shopMacaroon: null
      })
    }
  }

  customVerification = () => {
    if (this.state.shopVerif === 'sequential') {
      return (
        <div class="column col-sm-12 col-8 col-mx-auto form-group">
          <label class="form-label">Iterate over this words</label>
          <input
            class="form-input"
            type="text"
            name="shopSeqWords"
            placeholder="word1 word2 word..."
            value={this.state.shopSeqWords}
            onChange={this.handleInputChange}
          />
          <label class="form-label">Start at</label>
          <input
            class="form-input"
            type="number"
            name="shopSeqStart"
            value={this.state.shopSeqStart}
            min={1}
            onChange={this.handleInputChange}
          />
        </div>
      )
    } else {
      return (
        <div class="column col-sm-12 col-8 col-mx-auto form-group">
          <label class="form-label">HMAC Key</label>
          <input
            class="form-input"
            type="text"
            name="shopHmacKey"
            placeholder="HMAC Key"
            value={this.state.shopHmacKey}
            onChange={this.handleInputChange}
          />
          <label class="form-label">Set interval</label>
          <input
            class="form-input"
            type="number"
            name="shopHmacInt"
            value={this.state.shopHmacInt}
            min={1}
            onChange={this.handleInputChange}
          />
        </div>
      )
    }
  }

  lndBackend = () => {
    return (
      <>
        <div class="column col-sm-12 col-8 col-mx-auto form-group">
          <label class="form-label" for="shopEndpoint">
            HTTP Endpoint
          </label>
          <input
            class="form-input"
            type="text"
            name="shopEndpoint"
            placeholder="https://my.lnd:8080"
            value={this.state.shopEndpoint}
            onChange={this.handleInputChange}
          />
        </div>
        <div class="column col-sm-12 col-8 col-mx-auto form-group">
          <label class="form-label" for="shopMacaroon">
            Invoice Macaroon
          </label>
          <textarea
            class="form-input"
            rows={3}
            name="shopMacaroon"
            placeholder="https://my.lnd:8080"
            value={this.state.shopMacaroon}
            onChange={this.handleInputChange}
          />
          <p class="form-input-hint text-dark">as Hex</p>
        </div>
      </>
    )
  }

  sparkoBackend = () => {
    return (
      <>
        <div class="column col-sm-12 col-8 col-mx-auto form-group">
          <label class="form-label" for="shopEndpoint">
            HTTP Endpoint
          </label>
          <input
            class="form-input"
            type="text"
            name="shopEndpoint"
            placeholder="https://my.sparko:9737/rpc"
            value={this.state.shopEndpoint}
            onChange={this.handleInputChange}
          />
        </div>
        <div class="column col-sm-12 col-8 col-mx-auto form-group">
          <label class="form-label" for="shopInvoiceKey">
            Invoice Key
          </label>
          <input
            class="form-input"
            type="text"
            name="shopInvoiceKey"
            value={this.state.shopInvoiceKey}
            onChange={this.handleInputChange}
          />
        </div>
      </>
    )
  }

  lntxbotBackend = () => {
    return (
      <div class="column col-sm-12 col-8 col-mx-auto form-group">
        <label class="form-label" for="shopInvoiceKey">
          Invoice Key
        </label>
        <input
          class="form-input"
          type="text"
          name="shopInvoiceKey"
          value={this.state.shopInvoiceKey}
          onChange={this.handleInputChange}
        />
      </div>
    )
  }

  submitShop = e => {
    e.preventDefault()
    const state = this.state

    if (!state.shopBackEnd) return
    if (state.shopBackEnd === 'spark' && !state.shopEndpoint) return
    if (state.shopBackEnd === 'lnd' && !state.shopMacaroon) return
    if (state.shopBackEnd === 'lntxbot' && !state.shopInvoiceKey) return

    const backend = {kind: 'lnd', connection: {}}
    const shop = {verification: {}}

    backend.kind = state.shopBackEnd
    backend.connection.endpoint = state.shopEndpoint
    if (state.shopBackEnd === 'lnd') {
      backend.connection.macaroon = state.shopMacaroon
    } else {
      backend.connection.key = state.shopInvoiceKey
    }
    if (state.shopTLS) {
      backend.connection.cert = state.shopTLS
    }

    shop.id = state.shopID
    shop.message = state.shopThankYouMsg
    shop.verification.kind = state.shopVerif

    if (state.shopVerif === 'sequential') {
      if (state.shopSeqWords && state.shopSeqWords.length) {
        shop.verification.words = state.shopSeqWords
          .split(' ')
          .map(x => x.trim())
          .filter(x => x)
        shop.verification.init = state.shopSeqStart
      }
    }
    if (state.shopVerif === 'hmac') {
      shop.verification.key = state.shopHmacKey
      shop.verification.interval = state.shopHmacInt
    }
    const body = JSON.stringify({...backend, ...shop})

    const options = {
      method: 'PUT',
      body,
      headers: this.props.id ? {Authorization: 'Basic ' + state.auth} : {}
    }
    return fetch(`/api/shop/${shop.id}`, options)
      .then(res => res.json())
      .then(key => {
        idb.addShop(state.shopID, key)
        return key
      })
      .then(() => route(`/shop/${state.shopID}`))
      .catch(err => console.error(err))
  }

  componentDidMount = async () => {
    if (this.props.shop_id) {
      const shopID = this.props.shop_id
      const auth = await idb.getShopToken(shopID)
      const options = {
        method: 'GET',
        headers: {Authorization: 'Basic ' + auth}
      }
      return fetch(`/api/shop/${shopID}`, options)
        .then(res => res.json())
        .then(data => {
          this.setState({
            loaded: !this.state.loaded,
            auth,
            shopID,
            shopKey: data.key,
            shopVerif: data.verification.kind,
            shopSeqWords: data.verification.words || null,
            shopSeqStart: data.verification.init || null,
            shopHmacKey:
              data.verification.kind === 'hmac' ? data.verification.key : null,
            shopHmacInt:
              data.verification.kind === 'hmac'
                ? data.verification.interval
                : 10,
            shopThankYouMsg: data.message,
            shopBackEnd: null,
            shopEndpoint: null,
            shopMacaroon: null,
            shopInvoiceKey: null,
            shopTLS: null,
            loading: !this.state.loading
          })
          console.log(data)
        })
    }
    this.setState({loading: !this.state.loading})
  }

  // Note: `user` comes from the URL, courtesy of our router
  render(
    {shop_id},
    {loading, shopID, shopVerif, shopThankYouMsg, shopBackEnd, shopTLS}
  ) {
    return (
      <main class="container grid-lg">
        {loading ? (
          <div class="loading loading-lg"></div>
        ) : (
          <>
            <h1>{`${shop_id ? 'Edit Shop' : 'Create Shop'}`}</h1>
            <br />
            <div class="columns">
              <div class="column col-sm-12 col-8 col-mx-auto">
                <h3>Shop Details</h3>
              </div>
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <label class="form-label" for="shopID">
                  Shop Name
                </label>
                <input
                  class="form-input"
                  type="text"
                  name="shopID"
                  placeholder="Cookie Store"
                  value={shopID}
                  onChange={this.handleInputChange}
                />
              </div>
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <label class="form-label">Show verification after buy</label>
                <select
                  class="form-select"
                  value={shopVerif}
                  name="shopVerif"
                  onChange={this.handleInputChange}
                >
                  <option value="none">None</option>
                  <option value="sequential">Sequential</option>
                  <option value="hmac">Hmac</option>
                </select>
              </div>
              {shopVerif !== 'none' && this.customVerification()}
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <label class="form-label" for="shopThankYouMsg">
                  Send message after buy
                </label>
                <input
                  class="form-input"
                  type="text"
                  name="shopThankYouMsg"
                  placeholder="Thank you..."
                  value={shopThankYouMsg}
                  onChange={this.handleInputChange}
                />
              </div>
            </div>
            <br />
            <div class="columns">
              <div class="column col-sm-12 col-8 col-mx-auto">
                <h3>Lightning Backend</h3>
              </div>
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <label class="form-label">Choose backend</label>
                <select
                  class="form-select"
                  value={shopBackEnd}
                  name="shopBackEnd"
                  onChange={this.handleInputChange}
                >
                  <option value="lnd">LND</option>
                  <option value="spark">Spark</option>
                  <option value="lntxbot">@lntxbot</option>
                </select>
              </div>
              {shopBackEnd === 'lnd'
                ? this.lndBackend()
                : shopBackEnd === 'spark'
                ? this.sparkoBackend()
                : this.lntxbotBackend()}
              {shopBackEnd !== 'lntxbot' && (
                <div class="column col-sm-12 col-8 col-mx-auto form-group">
                  <label class="form-label" for="shopTLS">
                    Include TLS certificate
                  </label>
                  <textarea
                    class="form-input"
                    rows={3}
                    name="shopTLS"
                    placeholder="-----BEGIN CERTIFICATE-----
																						MIICPDCCAeKgAwIBAgIRAMzWW6TGNgkeFEW4QfJY0U0wCgYIKoZI
																						MB0GA1UEChMWbG5kIGF1dG9nZW5lcmF0ZWQgY2VydDEUMBIGA1UE
																						..."
                    value={shopTLS}
                  />
                  <p class="form-input-hint text-dark">Optional</p>
                </div>
              )}
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <button class="btn btn-primary" onClick={this.submitShop}>
                  Save
                </button>
              </div>
            </div>
          </>
        )}
      </main>
    )
  }
}
