import axios from 'axios'
import { ElMessage } from 'element-plus'

export const http = axios.create({ timeout: 20000 })

http.interceptors.request.use((cfg) => {
  const token = localStorage.getItem('gowiki.token')
  if (token) cfg.headers.Authorization = `Bearer ${token}`
  return cfg
})

http.interceptors.response.use(
  (res) => res.data,
  (err) => {
    const msg = err?.response?.data?.message || '请求失败'
    if (err?.response?.status === 401) {
      localStorage.removeItem('gowiki.token')
      if (!location.hash.includes('/login')) location.hash = '#/login'
    }
    ElMessage.error(msg)
    return Promise.reject(err)
  },
)

export type Api<T> = Promise<{ code: string; message: string; data: T; timestamp: string }>
