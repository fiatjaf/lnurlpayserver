/** @format */

import {h, Component} from 'preact'
import {route} from 'preact-router'
import {Link} from 'preact-router/match'
import {idb} from '../../idb'

import {TileCompact} from '../../components/tile_compact'
import {Empty} from '../../components/empty'

const emptyTemplates = id => (
  <div class="empty">
    <div class="empty-icon">
      <i class="icon icon-apps"></i>
    </div>
    <p class="empty-title h5">You have no templates</p>
    <p class="empty-subtitle">Click to start</p>
    <div class="empty-action">
      <Link href={`/shop/${id}/template/edit`} class="btn">
        Create
      </Link>
    </div>
  </div>
)

const showTemplates = (templates, shop_id) => (
  <div class="empty">
    <div class="columns">
      {templates.map((t, i) => (
        <div key={i} class={`column col-sm-12 col-md-6 col-4 my-2`}>
          <p class="empty-title h5">{t.id}</p>
          <p class="empty-subtitle">Select an option.</p>
          <div class="empty-action">
            <Link
              href={`/shop/${shop_id}/template/${t.id}`}
              class="btn btn-primary"
            >
              Open
            </Link>
            <Link
              href={`/shop/${shop_id}/template/edit/${t.id}`}
              class="btn ml-1"
            >
              Edit
            </Link>
          </div>
        </div>
      ))}
      <div class="column col-12 my-2">
        <div class="divider text-center" data-content="OR"></div>
      </div>
      <div class="column col-12 my-2">
        <Link href={`/shop/${shop_id}/template/edit`} class="btn btn-primary">
          Create New
        </Link>
      </div>
    </div>
  </div>
)

export default class Shop extends Component {
  state = {
    loading: true
  }

  componentDidMount = async () => {
    if (this.props.shop_id) {
      const shopID = this.props.shop_id
      const auth = await idb.getShopToken(shopID)
      const options = {
        method: 'GET',
        headers: {Authorization: 'Basic ' + auth}
      }
      fetch(`/api/shop/${shopID}`, options)
        .then(res => res.json())
        .then(data => {
          this.setState({
            shop: data
          })
          console.log(data)
        })
        .catch(err => console.error(err))
      fetch(`/api/shop/${shopID}/templates`, options)
        .then(res => res.json())
        .then(data => {
          this.setState({
            templates: data,
            loading: !this.state.loading
          })
          console.log('Templates', data)
        })
        .catch(err => console.error(err))
    }
  }

  // Note: `user` comes from the URL, courtesy of our router
  render({shop_id}, {loading, shop, templates}) {
    return (
      <main class="container grid-lg">
        {loading ? (
          <div class="loading loading-lg"></div>
        ) : (
          <>
            <h1>{shop.id}</h1>
            <br />
            <div class="columns">
              <TileCompact title="Backend" subtitle={shop.backend} />
              <TileCompact title="Key" subtitle={shop.key} />
              <TileCompact title="Message" subtitle={shop.message} />
              <TileCompact
                title="Verification"
                subtitle={shop.verification.kind}
              />
            </div>
            <div class="my-2">
              <br />
              <Link href={`/shop/edit/${shop.id}`} class="btn">
                Edit Shop
              </Link>
            </div>
            <br />
            <h2>Templates</h2>
            {templates.length
              ? showTemplates(templates, shop_id)
              : emptyTemplates(shop_id)}
            <br />
            <h2>Invoices</h2>
            <Empty title="You have no invoices yet!" />
          </>
        )}
      </main>
    )
  }
}
