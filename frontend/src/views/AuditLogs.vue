<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">审计日志</span>
      <el-input v-model="table.query.user_id" placeholder="用户ID" style="width: 120px" @keyup.enter="table.search" />
      <el-input v-model="table.query.resource" placeholder="资源" style="width: 140px" @keyup.enter="table.search" />
      <el-button @click="table.search">查询</el-button>
    </div>
    <el-card class="table-card">
      <el-table :data="table.list.value" v-loading="table.loading.value" stripe size="small">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户" width="110" />
        <el-table-column prop="action" label="动作" width="160" />
        <el-table-column prop="resource" label="资源" width="110" />
        <el-table-column prop="method" label="方法" width="80" />
        <el-table-column prop="path" label="路径" min-width="180" />
        <el-table-column prop="status_code" label="状态码" width="80" />
        <el-table-column prop="ip" label="IP" width="120" />
        <el-table-column prop="created_at" label="时间" width="180" />
      </el-table>
      <el-pagination style="margin-top: 12px; justify-content: flex-end; display: flex"
        :current-page="table.query.page" :page-size="table.query.page_size" :total="table.total.value"
        layout="total, sizes, prev, pager, next" @current-change="table.onPageChange" @size-change="table.onSizeChange" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useTable } from '@/composables/useTable'
import { auditApi } from '@/api'
const table = useTable((p) => auditApi.list(p))
table.load()
</script>
