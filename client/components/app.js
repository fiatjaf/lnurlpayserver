/** @format */

import {h, Component} from 'preact'
import {Router} from 'preact-router'

import Header from './header'

// Code-splitting is automated for routes
import Home from '../routes/home'
import Shop from '../routes/shop'
import CreateShop from '../routes/shop/createShop'
import Template from '../routes/template'
import EditTemplate from '../routes/template/create'

export default class App extends Component {
  /** Gets fired when the route changes.
   *	@param {Object} event		"change" event from [preact-router](http://git.io/preact-router)
   *	@param {string} event.url	The newly routed URL
   */
  handleRoute = e => {
    this.currentUrl = e.url
  }

  render() {
    return (
      <div id="app">
        <Header />
        <Router onChange={this.handleRoute}>
          <Home path="/" />
          <Shop path="/shop/:shop_id" />
          <CreateShop path="/shop/edit/:shop_id?" />
          <Template path="/shop/:shop_id/template/:template_id" />
          <EditTemplate path="/shop/:shop_id/template/edit/:template_id?" />
        </Router>
      </div>
    )
  }
}
