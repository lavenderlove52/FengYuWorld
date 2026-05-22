import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './styles/main.css'
import { useAuthStore } from './stores/auth'
import { getAccessToken } from './api/http'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

if (getAccessToken()) {
  useAuthStore(pinia).fetchMe()
}

app.mount('#app')
