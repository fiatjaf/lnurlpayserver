/** @format */

import {Store, set, get} from 'idb-keyval'

const lnurlData = new Store('lnurl-db', 'lnurlData')

export const idb = {
  //set(key, value, store)
  addShop: (id, key) => {
    return get('keys', lnurlData).then(current => {
      if (!current) {
        current = {}
      }
      current[id] = key
      set('keys', current, lnurlData)
      return current
    })
  },
  getShopToken: async id => {
    let current = await get('keys', lnurlData)
    if (!current) {
      current = {}
    }
    const key = current[id]
    console.log(id, key)
    return btoa(`key:${key}`)
  },
  getShops: async () => {
    const shops = await get('keys', lnurlData)
    return shops || {}
  }
}

//Shop Store_1 created with key d7e1b63aea5b02f5f6d869bf2addaca5!
