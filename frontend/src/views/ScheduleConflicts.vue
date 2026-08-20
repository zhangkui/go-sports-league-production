<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">冲突检测</span>
      <el-select v-model="seasonId" placeholder="赛季" clearable style="width: 200px" @change="load">
        <el-option v-for="s in seasonStore.seasons" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
    </div>
    <el-card class="table-card" v-loading="loading">
      <el-empty v-if="!conflicts.length" description="未检测到冲突" />
      <el-table v-else :data="conflicts" stripe>
        <el-table-column prop="schedule_id" label="赛程ID" width="100" />
        <el-table-column prop="conflict_type" label="类型" width="120" />
        <el-table-column prop="description" label="说明" min-width="280" />
        <el-table-column prop="detected_at" label="检测时间" width="180" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { scheduleApi } from '@/api'
import { useSeasonStore } from '@/stores/season'

const seasonStore = useSeasonStore()
const seasonId = ref<number | undefined>(undefined)
const conflicts = ref<any[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try { conflicts.value = await scheduleApi.conflicts(seasonId.value) } finally { loading.value = false }
}
onMounted(async () => {
  if (!seasonStore.seasons.length) await seasonStore.loadAll()
  seasonId.value = seasonStore.current?.id || seasonStore.seasons[0]?.id
  await load()
})
</script>
