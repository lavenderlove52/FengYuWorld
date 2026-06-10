import { createApp } from 'vue'
import { createPinia } from 'pinia'
import mermaid from 'mermaid'
import { config as mdEditorConfig } from 'md-editor-v3'
import App from './App.vue'
import router from './router'
import './styles/main.css'

mdEditorConfig({
  editorExtensions: {
    mermaid: {
      instance: mermaid,
      enableZoom: false,
    },
  },
  mermaidConfig: (base) => ({
    ...base,
    theme: 'base',
    themeVariables: {
      ...base.themeVariables,
      background: '#ffffff',
      primaryColor: '#f8fafc',
      primaryTextColor: '#0f172a',
      primaryBorderColor: '#64748b',
      lineColor: '#475569',
      textColor: '#0f172a',
      nodeTextColor: '#0f172a',
      edgeLabelBackground: '#ffffff',
      clusterBkg: '#ffffff',
      clusterBorder: '#cbd5e1',
    },
  }),
})

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

app.mount('#app')
