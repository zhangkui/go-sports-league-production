<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">比赛列表</span>
      <el-select v-model="table.query.status" placeholder="状态" clearable style="width: 140px" @change="table.search">
        <el-option v-for="s in statuses" :key="s" :label="statusLabels[s]" :value="s" />
      </el-select>
      <el-button @click="table.search">刷新</el-button>
    </div>
    <el-card class="table-card">
      <el-table :data="table.list.value" v-loading="table.loading.value" stripe>
        <el-table-column prop="match_date" label="日期" width="120" />
        <el-table-column prop="start_time" label="时间" width="80" />
        <el-table-column label="对阵" min-width="160">
          <template #default="{ row }">#{{ row.home_team_id }} vs #{{ row.away_team_id }}</template>
        </el-table-column>
        <el-table-column label="比分" width="100">
          <template #default="{ row }">{{ row.home_score ?? '-' }} : {{ row.away_score ?? '-' }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="statusTagType(row.status)" size="small">{{ label(matchStatusMap, row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="confirm_status" label="确认" width="100">
          <template #default="{ row }"><el-tag :type="row.confirm_status === 'confirmed' ? 'success' : 'warning'" size="small">{{ label(confirmStatusMap, row.confirm_status) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }"><el-button size="small" @click="$router.push(`/matches/${row.id}`)">详情</el-button></template>
        </el-table-column>
      </el-table>
      <el-pagination style="margin-top: 12px; justify-content: flex-end; display: flex"
        :current-page="table.query.page" :page-size="table.query.page_size" :total="table.total.value"
        layout="total, prev, pager, next" @current-change="table.onPageChange" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useTable } from '@/composables/useTable'
import { matchApi } from '@/api'
import { label, matchStatusMap, confirmStatusMap, statusTagType } from '@/utils/status'
const table = useTable((p) => matchApi.list(p))
const statuses = ['scheduled', 'confirmed', 'in_progress', 'completed', 'cancelled', 'postponed']
const statusLabels: Record<string, string> = { scheduled: '已排期', confirmed: '已确认', in_progress: '进行中', completed: '已完成', cancelled: '已取消', postponed: '已延期' }
table.load()
</script>
