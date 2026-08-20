<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">实时积分榜</span>
      <el-select v-model="seasonId" placeholder="赛季" style="width: 200px" @change="load">
        <el-option v-for="s in seasonStore.seasons" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
      <el-button @click="load">刷新</el-button>
    </div>
    <el-card class="table-card" v-loading="loading">
      <el-table :data="standings" stripe>
        <el-table-column type="index" label="排名" width="70" />
        <el-table-column prop="team_name" label="球队" min-width="140" />
        <el-table-column prop="played" label="赛" width="60" />
        <el-table-column prop="wins" label="胜" width="60" />
        <el-table-column prop="draws" label="平" width="60" />
        <el-table-column prop="losses" label="负" width="60" />
        <el-table-column prop="goals_for" label="进球" width="70" />
        <el-table-column prop="goals_against" label="失球" width="70" />
        <el-table-column prop="goal_diff" label="净胜球" width="80" />
        <el-table-column prop="points" label="积分" width="80">
          <template #default="{ row }"><span style="font-weight: 700; color: #409eff">{{ row.points }}</span></template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { standingApi } from '@/api'
import { useSeasonStore } from '@/stores/season'

const seasonStore = useSeasonStore()
const seasonId = ref<number | undefined>(undefined)
const standings = ref<any[]>([])
const loading = ref(false)

async function load() {
  if (!seasonId.value) return
  loading.value = true
  try { standings.value = await standingApi.list(seasonId.value) } finally { loading.value = false }
}
onMounted(async () => {
  if (!seasonStore.seasons.length) await seasonStore.loadAll()
  seasonId.value = seasonStore.current?.id || seasonStore.seasons[0]?.id
  await load()
})
</script>
