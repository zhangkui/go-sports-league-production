<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">场地管理</span>
      <el-button @click="table.search">刷新</el-button>
      <el-button type="primary" @click="$router.push('/venues/create')">添加场地</el-button>
    </div>
    <el-card class="table-card">
      <el-table :data="table.list.value" v-loading="table.loading.value" stripe>
        <el-table-column prop="code" label="编码" width="120" />
        <el-table-column prop="name" label="名称" min-width="160" />
        <el-table-column prop="address" label="地址" min-width="200" />
        <el-table-column prop="capacity" label="容量" width="90" />
        <el-table-column prop="sport" label="项目" width="90" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="statusTagType(row.status)" size="small">{{ label(venueStatusMap, row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }"><el-button size="small" @click="$router.push(`/venues/${row.id}`)">详情</el-button>
            <el-button size="small" type="danger" @click="remove(row)">删除</el-button></template>
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
import { venueApi } from '@/api'
import { label, venueStatusMap, statusTagType } from '@/utils/status'
const table = useTable((p) => venueApi.list(p))
async function remove(row: any) { await table.confirm(() => venueApi.remove(row.id), '已删除') }
table.load()
</script>
