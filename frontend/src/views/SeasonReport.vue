<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">赛季汇总报告</span>
      <el-select v-model="seasonId" placeholder="赛季" style="width: 200px" @change="load">
        <el-option v-for="s in seasonStore.seasons" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
    </div>
    <div v-loading="loading">
      <el-row :gutter="16" v-if="summary">
        <el-col :span="6" v-for="c in cards" :key="c.label">
          <el-card class="stat-card"><div class="stat-label">{{ c.label }}</div><div class="stat-value">{{ c.value }}</div></el-card>
        </el-col>
      </el-row>
      <el-card class="table-card" style="margin-top: 16px" v-if="summary?.top_scorer">
        <template #header><span>最佳射手</span></template>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="球员">{{ summary.top_scorer.player_name }}</el-descriptions-item>
          <el-descriptions-item label="进球">{{ summary.top_scorer.goals }}</el-descriptions-item>
          <el-descriptions-item label="助攻">{{ summary.top_scorer.assists }}</el-descriptions-item>
        </el-descriptions>
      </el-card>
      <el-card class="table-card" style="margin-top: 16px" v-if="summary?.top_team">
        <template #header><span>榜首球队</span></template>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="球队">{{ summary.top_team.team_name }}</el-descriptions-item>
          <el-descriptions-item label="积分">{{ summary.top_team.points }}</el-descriptions-item>
          <el-descriptions-item label="净胜球">{{ summary.top_team.goal_diff }}</el-descriptions-item>
        </el-descriptions>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { reportApi } from '@/api'
import { useSeasonStore } from '@/stores/season'

const seasonStore = useSeasonStore()
const seasonId = ref<number | undefined>(undefined)
const summary = ref<any>(null)
const loading = ref(false)
const cards = computed(() => summary.value ? [
  { label: '球队数', value: summary.value.team_count },
  { label: '球员数', value: summary.value.player_count },
  { label: '比赛数', value: summary.value.match_count },
  { label: '已完成', value: summary.value.completed_matches },
] : [])

async function load() {
  if (!seasonId.value) return
  loading.value = true
  try { summary.value = await reportApi.seasonSummary(seasonId.value) } finally { loading.value = false }
}
onMounted(async () => {
  if (!seasonStore.seasons.length) await seasonStore.loadAll()
  seasonId.value = seasonStore.current?.id || seasonStore.seasons[0]?.id
  await load()
})
</script>
