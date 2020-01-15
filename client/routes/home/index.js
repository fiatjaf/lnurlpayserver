/** @format */

import {h, Component} from 'preact'
import {Link} from 'preact-router/match'

import {idb} from '../../idb'

const emptyShops = () => (
  <div class="empty">
    <div class="empty-icon">
      <i class="icon icon-apps"></i>
    </div>
    <p class="empty-title h5">You have no shops</p>
    <p class="empty-subtitle">Select an option.</p>
    <div class="empty-action">
      <Link href="/shop/edit" class="btn btn-primary">
        Register
      </Link>
      {/* <button class="btn ml-1" onClick={getShopKey}>Adopt by Key</button> */}
    </div>
  </div>
)

const showShops = shops => (
  <div class="empty">
    <div class="columns">
      {shops.map((shop, i) => (
        <div key={i} class={`column col-sm-12 col-md-6 col-4 my-2`}>
          <p class="empty-title h5">{shop}</p>
          <p class="empty-subtitle">Select an option.</p>
          <div class="empty-action">
            <Link href={`/shop/${shop}`} class="btn btn-primary">
              Open Shop
            </Link>
            <Link href={`/shop/edit/${shop}`} class="btn ml-1">
              Edit Shop
            </Link>
          </div>
        </div>
      ))}
      <div class="column col-12 my-2">
        <div class="divider text-center" data-content="OR"></div>
      </div>
      <div class="column col-12 my-2">
        <Link href="/shop/edit" class="btn btn-primary">
          Register
        </Link>
        {/* <button class="btn ml-1" onClick={getShopKey}>Adopt by Key</button> */}
      </div>
    </div>
  </div>
)

export default class Home extends Component {
  state = {
    shops: null
  }

  componentDidMount() {
    const shops = idb.getShops().then(s => {
      if (Object.keys(s).length) {
        this.setState({shops: Object.keys(s)})
      }
    })
  }

  render({}, {shops}) {
    return (
      <main class="container grid-lg">
        <h1>Your Shops</h1>
        {shops ? showShops(shops) : emptyShops()}
      </main>
    )
  }
}
