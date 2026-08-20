import { defineStore } from 'pinia'
import { ref } from 'vue'
import { seasonApi } from '@/api'

export const useSeasonStore = defineStore('season', () => {
  const seasons = ref<any[]>([])
  const current = ref<any | null>(null)

  async function loadAll() {
    const res = await seasonApi.list({ page: 1, page_size: 100 })
    seasons.value = res.list || []
  }
  function setCurrent(s: any) {
    current.value = s
    if (s) localStorage.setItem('gsl_season', String(s.id))
  }
  function restore() {
    const id = localStorage.getItem('gsl_season')
    if (id) current.value = { id: Number(id) }
  }
  return { seasons, current, loadAll, setCurrent, restore }
})
