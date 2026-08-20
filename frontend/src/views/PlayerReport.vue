<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">球员数据排行</span>
      <el-select v-model="seasonId" placeholder="赛季" style="width: 200px" @change="load">
        <el-option v-for="s in seasonStore.seasons" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
    </div>
    <el-card class="table-card" v-loading="loading">
      <el-table :data="rankings" stripe>
        <el-table-column type="index" label="#" width="60" />
        <el-table-column prop="player_name" label="球员" min-width="140" />
        <el-table-column prop="team_name" label="球队" min-width="120" />
        <el-table-column prop="goals" label="进球" width="80" sortable />
        <el-table-column prop="assists" label="助攻" width="80" sortable />
        <el-table-column prop="yellow_cards" label="黄牌" width="80" />
        <el-table-column prop="red_cards" label="红牌" width="80" />
        <el-table-column prop="matches" label="出场" width="80" />
        <el-table-column prop="avg_rating" label="评分" width="90" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { reportApi } from '@/api'
import { useSeasonStore } from '@/stores/season'

const seasonStore = useSeasonStore()
const seasonId = ref<number | undefined>(undefined)
const rankings = ref<any[]>([])
const loading = ref(false)

async function load() {
  if (!seasonId.value) return
  loading.value = true
  try { rankings.value = await reportApi.playerRanking(seasonId.value) } finally { loading.value = false }
}
onMounted(async () => {
  if (!seasonStore.seasons.length) await seasonStore.loadAll()
  seasonId.value = seasonStore.current?.id || seasonStore.seasons[0]?.id
  await load()
})
</script>
