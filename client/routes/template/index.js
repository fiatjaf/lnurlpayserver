/** @format */

import {h, Component} from 'preact'
import {Link} from 'preact-router/match'
import QRCode from 'qrcode'
import {idb} from '../../idb'

import {TileCompact} from '../../components/tile_compact'
import {route} from 'preact-router'

export default class Template extends Component {
  state = {
    id: null,
    shop: null,
    path_params: [],
    query_params: [],
    description: '',
    image: null,
    currency: 'sat',
    min_price: '',
    max_price: '',
    lnurl: null,
    lnurlQR: null,
    lnurlparams: {},
    auth: null,
    fetchOpts: null,
    loading: true
  }

  handleInputLnurlParams = event => {
    event.preventDefault()
    const target = event.target
    const value = target.value
    const name = target.name

    this.setState(prevState => ({
      lnurlparams: {
        ...prevState.lnurlparams,
        [name]: value
      }
    }))
  }

  generateQR = async url => {
    try {
      return await QRCode.toDataURL(url)
    } catch (err) {
      console.error(err)
    }
  }

  generateLNURL = e => {
    e.preventDefault()
    let qs = new URLSearchParams()
    const {lnurlparams, shop, id, fetchOpts} = this.state
    for (const k in lnurlparams) {
      qs.set(k, lnurlparams[k])
    }
    console.log(qs, lnurlparams)
    fetch(`/api/shop/${shop}/template/${id}/lnurl?${qs.toString()}`, fetchOpts)
      .then(res => res.json())
      .then(async lnurl => {
        if (lnurl.error) {
          throw new Error(lnurl.error)
        }
        const QR = await this.generateQR(lnurl)
        console.log(lnurl, QR)
        this.setState({lnurl, lnurlQR: QR})
      })
      .catch(err => console.error(err))
  }

  // gets called when this route is navigated to
  componentDidMount = async () => {
    // if(!this.props.template_id) {route(`/shop/${this.props.shop_id}/template/edit`)}
    if (this.props.shop_id && this.props.template_id) {
      const id = this.props.shop_id
      const auth = await idb.getShopToken(id)
      const options = {
        method: 'GET',
        headers: {Authorization: 'Basic ' + auth}
      }
      const call = await fetch(
        `/api/shop/${id}/template/${this.props.template_id}`,
        options
      )
      const data = await call.json()
      if (data.error) {
        throw new Error(data.error)
      }
      console.log(data)
      this.setState({
        ...data,
        loading: !this.state.loading,
        auth,
        fetchOpts: options
      })
    }
  }

  // Note: `user` comes from the URL, courtesy of our router
  render(
    {edit},
    {
      id,
      description,
      image,
      path_params,
      query_params,
      lnurlparams,
      lnurl,
      lnurlQR,
      min_price,
      max_price,
      currency,
      shop,
      loading
    }
  ) {
    return (
      <main class="container grid-lg">
        {loading && <div class="loading loading-lg"></div>}
        {id && (
          <>
            <h1>{id}</h1>
            <br />
            <div class="columns">
              <TileCompact
                title="Params"
                subtitle={`${[...path_params, ...query_params].join()}`}
              />
              <TileCompact title="Description" subtitle={description} />
              {image && <TileCompact title="Image" image={image} />}
              <TileCompact title="Curency" subtitle={currency.toUpperCase()} />
              <TileCompact
                title="Price"
                subtitle={`${
                  min_price != max_price
                    ? 'Min: ' + min_price + ' | Max: ' + max_price
                    : min_price
                }`}
              />
            </div>
            <div class="my-2">
              <br />
              <Link href={`/shop/${shop}/template/edit/${id}`} class="btn">
                Edit Template
              </Link>
            </div>
            <br />
            <h2>Generate QR (lnurl)</h2>
            <div class="columns">
              {[...path_params, ...query_params].map((p, i) => (
                <div
                  class="column col-sm-12 col-8 col-mx-auto form-group"
                  key={i}
                >
                  <label class="form-label" for={`${p}`}>
                    {p}
                  </label>
                  <input
                    class="form-input"
                    type="text"
                    name={`${p}`}
                    value={lnurlparams[p]}
                    onChange={this.handleInputLnurlParams}
                  />
                </div>
              ))}
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <button class="btn btn-primary" onClick={this.generateLNURL}>
                  Generate
                </button>
              </div>
              <br />
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                {lnurl && (
                  <div class="card">
                    <div class="card-image p-centered">
                      <img src={lnurlQR} class="img-responsive" />
                    </div>
                    <div class="card-header">
                      <div class="card-title h5">Generated lnurl-pay</div>
                      <div class="card-subtitle text-gray"></div>
                    </div>
                    <div class="card-body" style={`overflow-wrap: break-word;`}>
                      {lnurl}
                    </div>
                    <div class="card-footer">
                      <button class="btn btn-primary">Save</button>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </>
        )}
      </main>
    )
  }
}
