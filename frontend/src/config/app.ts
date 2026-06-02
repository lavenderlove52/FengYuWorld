/** Vite base，生产环境为 /blog/ */
export const appBase = import.meta.env.BASE_URL.endsWith('/')
  ? import.meta.env.BASE_URL
  : `${import.meta.env.BASE_URL}/`

/** API 前缀，如 /blog/api */
export const apiBase = `${appBase}api`.replace(/\/+/g, '/')
