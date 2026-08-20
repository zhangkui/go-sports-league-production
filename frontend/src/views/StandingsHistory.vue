<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">历史快照</span>
      <el-select v-model="seasonId" placeholder="赛季" style="width: 200px" @change="load">
        <el-option v-for="s in seasonStore.seasons" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
    </div>
    <el-card class="table-card" v-loading="loading">
      <el-table :data="snapshots" stripe>
        <el-table-column prop="round" label="轮次" width="100" />
        <el-table-column prop="created_at" label="生成时间" width="200" />
        <el-table-column label="数据" min-width="200">
          <template #default="{ row }">
            <el-button size="small" @click="view(row)">查看JSON</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dlg" title="快照数据" width="640px">
      <el-input :model-value="snapText" type="textarea" :rows="16" readonly />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { standingApi } from '@/api'
import { useSeasonStore } from '@/stores/season'

const seasonStore = useSeasonStore()
const seasonId = ref<number | undefined>(undefined)
const snapshots = ref<any[]>([])
const loading = ref(false)
const dlg = ref(false)
const snapText = ref('')

async function load() {
  if (!seasonId.value) return
  loading.value = true
  try { snapshots.value = await standingApi.snapshots(seasonId.value) } finally { loading.value = false }
}
async function view(row: any) {
  const snap = await standingApi.snapshot(row.id)
  snapText.value = JSON.stringify(JSON.parse(snap.snapshot || snap), null, 2)
  dlg.value = true
}
onMounted(async () => {
  if (!seasonStore.seasons.length) await seasonStore.loadAll()
  seasonId.value = seasonStore.current?.id || seasonStore.seasons[0]?.id
  await load()
})
</script>
