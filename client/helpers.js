/**
 * @prettier
 */

import {shopKeys} from './shopStore'

export async function fetchShopData(shopId) {
  let r = await window.fetch(`/shop/${shopId}`, {
    method: 'GET',
    headers: {Authorization: 'Basic ' + shopKeys.getShopToken(shopId)}
  })
  let res = await r.json()
  if (res.error) {
    throw new Error(res.error)
  }

  if (res.verification.words) {
    res.verification.words = res.verification.words.join(' ')
  }

  return res
}

export async function fetchTemplateData(shopId, templateId) {
  let r = await window.fetch(`/shop/${shopId}/template/${templateId}`, {
    method: 'GET',
    headers: {Authorization: 'Basic ' + shopKeys.getShopToken(shopId)}
  })
  let res = await r.json()
  if (res.error) {
    throw new Error(res.error)
  }
  return res
}

export async function fetchTemplates(shopId) {
  let r = await window.fetch(`/shop/${shopId}/templates`, {
    method: 'GET',
    headers: {Authorization: 'Basic ' + shopKeys.getShopToken(shopId)}
  })
  let res = await r.json()
  if (res.error) {
    throw new Error(res.error)
  }
  return res
}
