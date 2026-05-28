import { ref, onUnmounted } from 'vue'

export interface GeetestResult {
  lot_number: string
  captcha_output: string
  pass_token: string
  gen_time: string
}

export function useGeetest(containerId: string, captchaId: string) {
  const verified = ref(false)
  const result = ref<GeetestResult | null>(null)
  let captchaObj: any = null

  const init = () => {
    if (!captchaId) return
    const win = window as any
    if (!win.initGeetest4) return

    win.initGeetest4(
      {
        captchaId,
        product: 'bind', // 手动触发模式
        language: 'zho',
      },
      (obj: any) => {
        captchaObj = obj

        obj.appendTo(`#${containerId}`)

        obj.onSuccess(() => {
          const r = obj.getValidate()
          result.value = {
            lot_number: r.lot_number,
            captcha_output: r.captcha_output,
            pass_token: r.pass_token,
            gen_time: r.gen_time,
          }
          verified.value = true
        })

        obj.onError(() => {
          verified.value = false
          result.value = null
        })

        obj.onClose(() => {
          if (!verified.value) {
            result.value = null
          }
        })
      },
    )
  }

  const showCaptcha = (): Promise<GeetestResult> => {
    return new Promise((resolve, reject) => {
      // GeeTest 未配置时，直接跳过验证，返回空字段
      if (!captchaId) {
        resolve({ lot_number: '', captcha_output: '', pass_token: '', gen_time: '' })
        return
      }
      if (!captchaObj) {
        reject(new Error('验证组件未初始化'))
        return
      }
      // 重置上一次结果
      verified.value = false
      result.value = null

      captchaObj.showCaptcha()

      let resolved = false
      const unwatch = setInterval(() => {
        if (result.value) {
          clearInterval(unwatch)
          resolved = true
          resolve(result.value)
        }
      }, 200)

      // 30秒超时，确保所有路径都清理 interval
      setTimeout(() => {
        clearInterval(unwatch)
        if (!resolved) reject(new Error('验证超时，请重试'))
      }, 30000)
    })
  }

  const reset = () => {
    verified.value = false
    result.value = null
    captchaObj?.reset?.()
  }

  onUnmounted(() => {
    captchaObj?.destroy?.()
  })

  return { verified, result, init, showCaptcha, reset }
}
