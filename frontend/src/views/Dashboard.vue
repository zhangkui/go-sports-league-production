<template>
  <div class="page-container">
    <el-row :gutter="16" v-loading="loading">
      <el-col :span="6" v-for="card in stats" :key="card.label">
        <el-card class="stat-card">
          <div class="stat-label">{{ card.label }}</div>
          <div class="stat-value">{{ card.value }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="14">
        <el-card class="table-card">
          <template #header><span>近期比赛</span></template>
          <el-table :data="recentMatches" size="small" stripe>
            <el-table-column prop="match_date" label="日期" width="110" />
            <el-table-column label="对阵" min-width="160">
              <template #default="{ row }">{{ teamName(row.home_team_id) }} vs {{ teamName(row.away_team_id) }}</template>
            </el-table-column>
            <el-table-column label="比分" width="90">
              <template #default="{ row }">{{ row.home_score ?? '-' }} : {{ row.away_score ?? '-' }}</template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="90" />
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card class="table-card">
          <template #header><span>积分榜 TOP 5</span></template>
          <el-table :data="topStandings" size="small" stripe>
            <el-table-column type="index" label="#" width="50" />
            <el-table-column prop="team_name" label="球队" />
            <el-table-column prop="points" label="积分" width="80" />
            <el-table-column prop="goal_diff" label="净胜球" width="90" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { matchApi, standingApi, seasonApi, teamApi } from '@/api'
import { useSeasonStore } from '@/stores/season'

const seasonStore = useSeasonStore()
const loading = ref(true)
const recentMatches = ref<any[]>([])
const topStandings = ref<any[]>([])
const teams = ref<any[]>([])

const stats = computed(() => [
  { label: '赛季总数', value: seasonStore.seasons.length },
  { label: '近期比赛', value: recentMatches.value.length },
  { label: '积分榜球队', value: topStandings.value.length },
  { label: '当前用户', value: 'admin' },
])

function teamName(id: number) {
  return teams.value.find((t) => t.id === id)?.name || `#${id}`
}

onMounted(async () => {
  loading.value = true
  try {
    if (!seasonStore.seasons.length) await seasonStore.loadAll()
    const recent = await matchApi.list({ page: 1, page_size: 5 })
    recentMatches.value = recent.list || []
    const seasonId = seasonStore.current?.id || seasonStore.seasons[0]?.id
    if (seasonId) {
      const stands = await standingApi.list(seasonId)
      topStandings.value = (stands as any[]).slice(0, 5)
      const t = await teamApi.list({ season_id: seasonId, page: 1, page_size: 100 })
      teams.value = t.list || []
    }
  } finally {
    loading.value = false
  }
})
</script>
