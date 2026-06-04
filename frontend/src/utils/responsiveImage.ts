const POST_DESKTOP_VARIANT = '/objects/post-image-desktop-v2/'
const POST_MOBILE_VARIANT = '/objects/post-image-mobile-v1/'

export function postMobileImageURL(url: string) {
  if (!url.includes(POST_DESKTOP_VARIANT)) return ''
  return url.replace(POST_DESKTOP_VARIANT, POST_MOBILE_VARIANT)
}
