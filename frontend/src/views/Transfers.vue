<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">转会管理</span>
      <el-button @click="table.search">刷新</el-button>
    </div>
    <el-card class="table-card">
      <el-table :data="table.list.value" v-loading="table.loading.value" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="player_id" label="球员ID" width="90" />
        <el-table-column prop="from_team_id" label="原球队" width="90" />
        <el-table-column prop="to_team_id" label="新球队" width="90" />
        <el-table-column prop="reason" label="原因" min-width="160" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="statusTagType(row.status)" size="small">{{ label(transferStatusMap, row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 'pending'" size="small" type="success" @click="review(row, 'approved')">批准</el-button>
            <el-button v-if="row.status === 'pending'" size="small" type="danger" @click="review(row, 'rejected')">拒绝</el-button>
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
import { transferApi } from '@/api'
import { label, transferStatusMap, statusTagType } from '@/utils/status'
const table = useTable((p) => transferApi.list(p))
async function review(row: any, status: string) {
  await table.confirm(() => transferApi.review(row.id, { status }), '已处理')
}
table.load()
</script>
