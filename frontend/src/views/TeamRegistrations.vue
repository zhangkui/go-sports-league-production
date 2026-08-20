<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">球队注册审核</span>
      <el-select v-model="table.query.status" placeholder="状态" clearable style="width: 140px" @change="table.search">
        <el-option v-for="s in statuses" :key="s" :label="statusLabels[s]" :value="s" />
      </el-select>
    </div>
    <el-card class="table-card">
      <el-table :data="table.list.value" v-loading="table.loading.value" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="team_id" label="球队ID" width="90" />
        <el-table-column prop="season_id" label="赛季ID" width="90" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="statusTagType(row.status)">{{ label(teamStatusMap, row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="review_note" label="审核备注" min-width="180" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="success" @click="review(row, 'approved')">批准</el-button>
            <el-button size="small" type="danger" @click="review(row, 'rejected')">拒绝</el-button>
          </template>
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
import { teamApi } from '@/api'
import { label, teamStatusMap, statusTagType } from '@/utils/status'
const table = useTable((p) => teamApi.registrations(p))
const statuses = ['pending', 'approved', 'rejected', 'suspended']
const statusLabels: Record<string, string> = { pending: '待审核', approved: '已批准', rejected: '已拒绝', suspended: '已停赛' }
async function review(row: any, status: string) {
  await table.confirm(() => teamApi.review(row.team_id, { status }), '已审核')
}
table.load()
</script>
