/**
 * @prettier
 */

import {writable} from 'svelte/store'
const {subscribe, set, update} = writable(
  JSON.parse(localStorage.getItem('keys') || '{}')
)

export const shopKeys = {
  subscribe,
  addShop: (id, key) => {
    let current = JSON.parse(localStorage.getItem('keys') || '{}')
    current[id] = key
    localStorage.setItem('keys', JSON.stringify(current))
    return set(current)
  },
  getShopToken: id => {
    let current = JSON.parse(localStorage.getItem('keys') || '{}')
    let key = current[id]
    return btoa('key:' + key)
  }
}
