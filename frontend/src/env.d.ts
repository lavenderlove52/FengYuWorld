/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_GEETEST_LOGIN_CAPTCHA_ID: string
  readonly VITE_GEETEST_REGISTER_CAPTCHA_ID: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
