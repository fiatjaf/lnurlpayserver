/** @format */

import {h} from 'preact'
import {Link} from 'preact-router/match'

const Header = () => (
  <header class="navbar bg-secondary">
    <div class="container grid-lg">
      <div class="columns" style={'height: 100%;'}>
        <section class="navbar-section"></section>
        <section class="navbar-center">
          <Link href="/" class="navbar-brand mr-2">
            lnurl Pay Server
          </Link>
        </section>
        <section class="navbar-section">
          {/* <nav>
						<Link class='btn btn-link' href="/">Home</Link>
						<Link class='btn btn-link' href="/profile">Docs</Link>
					</nav> */}
        </section>
      </div>
    </div>
  </header>
)

export default Header
