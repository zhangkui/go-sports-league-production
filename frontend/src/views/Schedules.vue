<template>
  <div class="page-container">
    <div class="page-toolbar">
      <el-select v-model="table.query.season_id" placeholder="赛季" clearable style="width: 200px" @change="table.search">
        <el-option v-for="s in seasonStore.seasons" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
      <el-input v-model="table.query.round" placeholder="轮次" style="width: 100px" @keyup.enter="table.search" />
      <el-button @click="table.search">刷新</el-button>
      <el-button type="primary" v-if="hasAny('schedules:generate','schedules:manage')" @click="genDlg = true">生成赛程</el-button>
    </div>
    <el-card class="table-card">
      <el-table :data="table.list.value" v-loading="table.loading.value" stripe>
        <el-table-column prop="round" label="轮次" width="80" />
        <el-table-column prop="match_date" label="日期" width="120" />
        <el-table-column prop="start_time" label="时间" width="90" />
        <el-table-column label="对阵" min-width="180">
          <template #default="{ row }">#{{ row.home_team_id }} vs #{{ row.away_team_id }}</template>
        </el-table-column>
        <el-table-column prop="venue_id" label="场地ID" width="90" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="statusTagType(row.status)" size="small">{{ label(matchStatusMap, row.status) }}</el-tag></template>
        </el-table-column>
      </el-table>
      <el-pagination style="margin-top: 12px; justify-content: flex-end; display: flex"
        :current-page="table.query.page" :page-size="table.query.page_size" :total="table.total.value"
        layout="total, prev, pager, next" @current-change="table.onPageChange" />
    </el-card>

    <el-dialog v-model="genDlg" title="生成赛程" width="460px">
      <el-form :model="gen" label-position="top">
        <el-form-item label="赛季ID"><el-input-number v-model="gen.season_id" :min="1" style="width: 100%" /></el-form-item>
        <el-form-item label="赛制"><el-select v-model="gen.format" style="width: 100%"><el-option label="单循环" value="round_robin" /><el-option label="双循环" value="double_round" /><el-option label="混合" value="mixed" /><el-option label="淘汰赛" value="knockout" /></el-select></el-form-item>
        <el-form-item label="场地ID（可选）"><el-input-number v-model="gen.venue_id" :min="0" style="width: 100%" /></el-form-item>
        <el-form-item label="开始日期"><el-date-picker v-model="gen.start_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" /></el-form-item>
        <el-form-item label="每日场次"><el-input-number v-model="gen.games_per_day" :min="1" style="width: 100%" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="genDlg=false">取消</el-button><el-button type="primary" :loading="genLoading" @click="generate">生成</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { scheduleApi } from '@/api'
import { useTable } from '@/composables/useTable'
import { useSeasonStore } from '@/stores/season'
import { usePermission } from '@/composables/usePermission'
import { label, matchStatusMap, statusTagType } from '@/utils/status'

const seasonStore = useSeasonStore()
const { hasAny } = usePermission()
const table = useTable((p) => scheduleApi.list(p))
const genDlg = ref(false)
const genLoading = ref(false)
const gen = reactive<any>({ season_id: undefined, format: 'round_robin', venue_id: 0, start_date: '', games_per_day: 2 })

async function generate() {
  genLoading.value = true
  try {
    const res = await scheduleApi.generate(gen)
    ElMessage.success(`已生成 ${res.count} 场比赛，冲突 ${res.conflicts?.length || 0} 项`)
    genDlg.value = false
    await table.load()
  } finally { genLoading.value = false }
}
table.load()
</script>
