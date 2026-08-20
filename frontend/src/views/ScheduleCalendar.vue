<template>
  <div class="page-container">
    <div class="page-toolbar">
      <el-select v-model="seasonId" placeholder="赛季" style="width: 200px" @change="load">
        <el-option v-for="s in seasonStore.seasons" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
    </div>
    <el-card class="table-card">
      <el-calendar v-model="date">
        <template #date-cell="{ data }">
          <div style="height: 100%" @click="selectDay(data.day)">
            <div>{{ data.day.slice(8) }}</div>
            <div v-for="m in byDay(data.day)" :key="m.id" style="font-size: 11px; color: #409eff">
              #{{ m.home_team_id }} vs #{{ m.away_team_id }}
            </div>
          </div>
        </template>
      </el-calendar>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { scheduleApi } from '@/api'
import { useSeasonStore } from '@/stores/season'

const seasonStore = useSeasonStore()
const seasonId = ref<number | undefined>(undefined)
const date = ref(new Date())
const schedules = ref<any[]>([])

function byDay(day: string) {
  return schedules.value.filter((s) => (s.match_date || '').slice(0, 10) === day)
}
function selectDay(day: string) { /* could open detail */ }

async function load() {
  if (!seasonId.value) return
  const res = await scheduleApi.list({ season_id: seasonId.value, page: 1, page_size: 200 })
  schedules.value = res.list || []
}
onMounted(async () => {
  if (!seasonStore.seasons.length) await seasonStore.loadAll()
  seasonId.value = seasonStore.current?.id || seasonStore.seasons[0]?.id
  await load()
})
</script>
