import { ref } from 'vue'
import { getCaptcha } from '@/api/auth'

function toDataUrl(image: string): string {
  if (!image) return ''
  if (image.startsWith('data:')) return image
  return `data:image/png;base64,${image}`
}

export function useCaptcha() {
  const captchaId = ref('')
  const captchaImage = ref('')
  const loading = ref(false)
  const loadError = ref('')

  const refresh = async () => {
    loading.value = true
    loadError.value = ''
    captchaId.value = ''
    captchaImage.value = ''
    try {
      const res = await getCaptcha()
      if (res.code === 0 && res.data?.id && res.data?.image) {
        captchaId.value = res.data.id
        captchaImage.value = toDataUrl(res.data.image)
      } else {
        loadError.value = res.message || '验证码加载失败'
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : ''
      loadError.value = msg || '验证码加载失败，请检查网络'
    } finally {
      loading.value = false
    }
  }

  return { captchaId, captchaImage, loading, loadError, refresh }
}
