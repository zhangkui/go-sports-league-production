import { useAuthStore } from '@/stores/auth'

export function usePermission() {
  const auth = useAuthStore()
  return {
    has: (perm: string) => auth.has(perm),
    hasAny: (...perms: string[]) => auth.hasAny(...perms),
    isAdmin: () => auth.isAdmin,
    roleCodes: () => auth.roleCodes,
  }
}
