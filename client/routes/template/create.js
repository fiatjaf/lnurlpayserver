/** @format */

import {h, Component} from 'preact'
import {route} from 'preact-router'
import {idb} from '../../idb'

export default class Template extends Component {
  state = {
    id: null,
    shop: null,
    path_params: [],
    query_params: [],
    description: null,
    image: null,
    currency: 'sat',
    min_price: null,
    max_price: null,
    auth: null,
    fetchOpts: null,
    loading: true
  }

  handleInputChange = event => {
    const target = event.target
    const value = target.type === 'checkbox' ? target.checked : target.value
    const name = target.name

    this.setState({
      [name]: value
    })
  }

  handleInputParams = params => {
    event.preventDefault()
    const target = event.target
    const value = target.value
    const idx = target.name
    const newParams = this.state[params]
    newParams[idx] = value
    this.setState({[params]: newParams})
  }

  addParams = params => {
    event.preventDefault()
    this.setState({[params]: [...this.state[params], '']})
  }

  submitTemplate = async () => {
    event.preventDefault()
    const template = this.state
    template.path_params = template.path_params
      .map(c => c.trim())
      .filter(c => c)
    template.query_params = template.query_params
      .map(c => c.trim())
      .filter(c => c)
    if (template.image && template.image.trim() === '') {
      delete template.image
    }
    if (!template.max_price) {
      template.max_price = template.min_price
    }
    const options = {
      method: 'PUT',
      body: JSON.stringify(template),
      headers: {'Authorization': 'Basic ' + this.state.auth}
    }
    await fetch(`/api/shop/${this.props.shop_id}/template/${template.id}`, options)
      .catch(err => console.error(err))
    return route(`/shop/${this.state.shop}`)
  }

  // gets called when this route is navigated to
  componentDidMount = async () => {
    const id = this.props.shop_id
    const auth = await idb.getShopToken(id)
    if (this.props.template_id) {
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
      this.setState({...data, loading: !this.state.loading, auth})
    } else {
      this.setState({loading: !this.state.loading, shop: this.props.shop_id, auth})
    }
  }

  // Note: `user` comes from the URL, courtesy of our router
  render(
    {template_id},
    {
      id,
      description,
      image,
      path_params,
      query_params,
      min_price,
      max_price,
      currency,
      loading
    }
  ) {
    return (
      <main class="container grid-lg">
        {loading ? (
          <div class="loading loading-lg"></div>
        ) : (
          <>
            {template_id ? <h1>Edit Template</h1> : <h1>Create Template</h1>}
            <div class="columns">
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <label class="form-label" for="id">
                  Template Name (ID)
                </label>
                <input
                  class="form-input"
                  type="text"
                  name="id"
                  placeholder="Cookie"
                  value={id}
                  onChange={this.handleInputChange}
                />
              </div>
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <label class="form-label" for="id">
                  Rigid Parameters
                </label>
                {path_params.map((c, i) => (
                  <div class="form-group">
                    <input
                      class="form-input"
                      type="text"
                      name={i}
                      placeholder="Cookie"
                      value={c}
                      onChange={() => this.handleInputParams('path_params')}
                    />
                  </div>
                ))}
              </div>
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <button
                  class="btn btn-primary input-group-btn tooltip"
                  data-tooltip="Add another"
                  onClick={() => this.addParams('path_params')}
                >
                  Add
                </button>
              </div>
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <label class="form-label" for="id">
                  Flexible Parameters
                </label>
                {query_params.map((c, i) => (
                  <div class="form-group">
                    <input
                      class="form-input"
                      type="text"
                      name={i}
                      placeholder="Cookie"
                      value={c}
                      onChange={() => this.handleInputParams('query_params')}
                    />
                  </div>
                ))}
              </div>
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <button
                  class="btn btn-primary input-group-btn tooltip"
                  data-tooltip="Add another"
                  onClick={() => this.addParams('query_params')}
                >
                  Add
                </button>
              </div>
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <label class="form-label" for="description">
                  Description
                </label>
                <textarea
                  class="form-input"
                  rows={5}
                  name="description"
                  placeholder={`Jan 3 2035\nMyShop, MyShopstreet 18, MyTown\nItem: MyItem\nQuantity: {{quantity}}\nColor: {{color}}`}
                  value={description}
                  onChange={this.handleInputChange}
                />
                <p class="form-input-hint text-dark">mustache template</p>
              </div>
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <label class="form-label" for="image">
                  Image
                </label>
                <textarea
                  class="form-input"
                  rows={5}
                  name="image"
                  placeholder="data:image/png;base64,..."
                  value={image}
                  onChange={this.handleInputChange}
                />
                <p class="form-input-hint text-dark">as base64</p>
              </div>
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <label class="form-label">Currency</label>
                <select
                  class="form-select"
                  value={currency}
                  name="currency"
                  onChange={this.handleInputChange}
                >
                  <option value="sat">satoshi</option>
                  <option value="usd">USD</option>
                  <option value="eur">EUR</option>
                  <option value="gbp">GBP</option>
                  <option value="cad">CAD</option>
                  <option value="jpy">JPY</option>
                </select>
              </div>
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <label class="form-label">Min Price</label>
                <input
                  class="form-input"
                  type="number"
                  name="min_price"
                  placeholder="quantity * 10000"
                  value={min_price}
                  onChange={this.handleInputChange}
                />
                <label class="form-label">Max Price</label>
                <input
                  class="form-input"
                  type="number"
                  name="max_price"
                  placeholder="quantity * 10000"
                  value={max_price}
                  onChange={this.handleInputChange}
                />
                <p class="form-input-hint text-dark">
                  You can use an expression to calculate price (ex. if quantity
                  parameter is set)
                </p>
              </div>
              <div class="column col-sm-12 col-8 col-mx-auto form-group">
                <button class="btn btn-primary" onClick={this.submitTemplate}>
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
