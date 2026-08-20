import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'

const ACCESS_KEY = 'gsl_access'
const REFRESH_KEY = 'gsl_refresh'

export const tokenStore = {
  get access() { return localStorage.getItem(ACCESS_KEY) || '' },
  get refresh() { return localStorage.getItem(REFRESH_KEY) || '' },
  set(accessToken: string, refreshToken: string) {
    localStorage.setItem(ACCESS_KEY, accessToken)
    localStorage.setItem(REFRESH_KEY, refreshToken)
  },
  clear() {
    localStorage.removeItem(ACCESS_KEY)
    localStorage.removeItem(REFRESH_KEY)
  },
}

export const http: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

http.interceptors.request.use((config) => {
  const token = tokenStore.access
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

let refreshing = false
let waiters: Array<(t: string) => void> = []

async function doRefresh(): Promise<string> {
  const refresh = tokenStore.refresh
  if (!refresh) throw new Error('no refresh token')
  const { data } = await axios.post('/api/v1/auth/refresh', { refresh_token: refresh })
  tokenStore.set(data.access_token, data.refresh_token)
  return data.access_token as string
}

http.interceptors.response.use(
  (resp) => {
    const env = resp.data
    if (env && typeof env === 'object' && 'code' in env && env.code !== 0 && env.code != null) {
      ElMessage.error(env.message || '请求失败')
      return Promise.reject(new Error(env.message || 'business error'))
    }
    return resp
  },
  async (error) => {
    const status = error.response?.status
    const env = error.response?.data
    const message = env?.message || error.message
    if (status === 401 && !error.config.__retried) {
      if (tokenStore.refresh) {
        if (refreshing) {
          return new Promise((resolve, reject) => {
            waiters.push((token) => {
              error.config.headers.Authorization = `Bearer ${token}`
              resolve(http.request(error.config))
            })
          })
        }
        refreshing = true
        try {
          const token = await doRefresh()
          refreshing = false
          waiters.forEach((w) => w(token))
          waiters = []
          error.config.headers = error.config.headers || {}
          error.config.headers.Authorization = `Bearer ${token}`
          error.config.__retried = true
          return http.request(error.config)
        } catch (e) {
          refreshing = false
          waiters = []
          tokenStore.clear()
          ElMessage.error('登录已过期，请重新登录')
          setTimeout(() => (window.location.href = '/login'), 800)
          return Promise.reject(e)
        }
      }
      tokenStore.clear()
    }
    if (status && status >= 400) {
      ElMessage.error(message)
    }
    return Promise.reject(error)
  },
)

export async function get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
  const { data } = await http.get(url, config)
  return data.data !== undefined ? data.data : (data as T)
}

export async function getList<T = any>(url: string, config?: AxiosRequestConfig): Promise<{ list: T[]; total: number; page: number; page_size: number }> {
  const { data } = await http.get(url, config)
  return data.data
}

export async function post<T = any>(url: string, body?: any, config?: AxiosRequestConfig): Promise<T> {
  const { data } = await http.post(url, body, config)
  return data.data !== undefined ? data.data : (data as T)
}

export async function put<T = any>(url: string, body?: any, config?: AxiosRequestConfig): Promise<T> {
  const { data } = await http.put(url, body, config)
  return data.data !== undefined ? data.data : (data as T)
}

export async function del<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
  const { data } = await http.delete(url, config)
  return data.data !== undefined ? data.data : (data as T)
}
