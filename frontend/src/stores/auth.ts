import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api'
import { tokenStore } from '@/api/client'

export interface UserPerm {
  id: number
  username: string
  email: string
  full_name: string
  status: number
  roles: { id: number; code: string; name: string }[]
  permissions: string[]
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserPerm | null>(null)
  const loading = ref(false)

  const isLoggedIn = computed(() => !!tokenStore.access)
  const isAuthed = computed(() => !!user.value)
  const isAdmin = computed(() => !!user.value?.roles?.some((r) => r.code === 'admin'))
  const roleCodes = computed(() => user.value?.roles?.map((r) => r.code) || [])

  function has(perm: string): boolean {
    if (!user.value) return false
    if (user.value.roles?.some((r) => r.code === 'admin')) return true
    return user.value.permissions?.includes(perm) || false
  }
  function hasAny(...perms: string[]): boolean {
    return perms.some((p) => has(p))
  }

  async function login(username: string, password: string) {
    loading.value = true
    try {
      const data = await authApi.login({ username, password })
      tokenStore.set(data.access_token, data.refresh_token)
      await fetchMe()
    } finally {
      loading.value = false
    }
  }

  async function fetchMe() {
    if (!tokenStore.access) return
    try {
      user.value = await authApi.me()
    } catch {
      tokenStore.clear()
      user.value = null
    }
  }

  async function logout() {
    try {
      await authApi.logout(tokenStore.refresh)
    } catch { /* ignore */ }
    tokenStore.clear()
    user.value = null
  }

  return { user, loading, isLoggedIn, isAuthed, isAdmin, roleCodes, has, hasAny, login, fetchMe, logout }
})
