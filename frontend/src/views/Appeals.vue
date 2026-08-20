<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">申诉管理</span>
      <el-button @click="table.search">刷新</el-button>
    </div>
    <el-card class="table-card">
      <el-table :data="table.list.value" v-loading="table.loading.value" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="discipline_id" label="处罚ID" width="90" />
        <el-table-column prop="reason" label="申诉理由" min-width="220" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="statusTagType(row.status)" size="small">{{ label(appealStatusMap, row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="review_opinion" label="处理意见" min-width="180" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 'pending'">
              <el-button size="small" type="success" @click="review(row, 'accepted')">接受</el-button>
              <el-button size="small" type="danger" @click="review(row, 'rejected')">驳回</el-button>
            </template>
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
import { appealApi } from '@/api'
import { label, appealStatusMap, statusTagType } from '@/utils/status'
const table = useTable((p) => appealApi.list(p))
async function review(row: any, status: string) {
  await table.confirm(() => appealApi.review(row.id, { status, opinion: '处理意见' }), '已处理')
}
table.load()
</script>
