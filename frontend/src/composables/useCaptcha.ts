import { ref } from 'vue'
import { getCaptcha } from '@/api/auth'

export function useCaptcha() {
  const captchaId = ref('')
  const captchaImage = ref('')
  const loading = ref(false)

  const refresh = async () => {
    loading.value = true
    try {
      const res = await getCaptcha()
      if (res.code === 0 && res.data) {
        captchaId.value = res.data.id
        captchaImage.value = res.data.image
      }
    } finally {
      loading.value = false
    }
  }

  return { captchaId, captchaImage, loading, refresh }
}
