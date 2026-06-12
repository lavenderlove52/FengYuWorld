import http from './http'

export interface PushSettings {
  bark_endpoint: string
  ntfy_base_url: string
  ntfy_topic: string
  ntfy_public_enabled: boolean
  ntfy_public_token?: string
}

export type BarkMode = 'message' | 'title' | 'sound' | 'call' | 'critical'
export type NtfyMode = 'message' | 'title' | 'urgent' | 'priority' | 'tags'

export const getPushSettings = () =>
  http.get('/me/push-settings').then((response) => response.data)

export const updatePushSettings = (barkEndpoint: string) =>
  http.put('/me/push-settings', { bark_endpoint: barkEndpoint }).then((response) => response.data)

export const sendBarkPush = (payload: { mode: BarkMode; title?: string; message: string }) =>
  http.post('/me/push/bark', payload).then((response) => response.data)

export const updateNtfyTopic = (topicSuffix: string) =>
  http.put('/me/push-settings/ntfy', { topic_suffix: topicSuffix }).then((response) => response.data)

export const updateNtfyPublic = (enabled: boolean) =>
  http.put('/me/push-settings/ntfy/public', { enabled }).then((response) => response.data)

export interface NtfyPushPayload {
  mode: NtfyMode
  title?: string
  message: string
  priority?: number
  tags?: string
}

export const sendNtfyPush = (payload: NtfyPushPayload) =>
  http.post('/me/push/ntfy', payload).then((response) => response.data)

export const getPublicNtfyPush = (token: string) =>
  http.get(`/push/public/${encodeURIComponent(token)}`).then((response) => response.data)

export const sendPublicNtfyPush = (token: string, payload: NtfyPushPayload) =>
  http.post(`/push/public/${encodeURIComponent(token)}`, payload).then((response) => response.data)
