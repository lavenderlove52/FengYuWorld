import { config as mdEditorConfig } from 'md-editor-v3'

let markdownExtensionsReady: Promise<void> | null = null

export const ensureMarkdownExtensions = () => {
  if (!markdownExtensionsReady) {
    markdownExtensionsReady = import('mermaid').then(({ default: mermaid }) => {
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
    })
  }
  return markdownExtensionsReady
}
