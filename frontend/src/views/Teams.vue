<template>
  <div class="page-container">
    <div class="page-toolbar">
      <el-select v-model="table.query.season_id" placeholder="赛季" clearable style="width: 200px" @change="table.search">
        <el-option v-for="s in seasonStore.seasons" :key="s.id" :label="s.name" :value="s.id" />
      </el-select>
      <el-select v-model="table.query.status" placeholder="状态" clearable style="width: 120px" @change="table.search">
        <el-option v-for="s in statuses" :key="s" :label="statusLabels[s]" :value="s" />
      </el-select>
      <el-input v-model="table.query.q" placeholder="球队名称" style="width: 200px" clearable @keyup.enter="table.search" />
      <el-button @click="table.search">搜索</el-button>
      <el-button type="primary" @click="$router.push('/teams/create')">创建球队</el-button>
    </div>
    <el-card class="table-card">
      <el-table :data="table.list.value" v-loading="table.loading.value" stripe>
        <el-table-column prop="code" label="编码" width="120" />
        <el-table-column prop="name" label="名称" min-width="160" />
        <el-table-column prop="contact" label="联系人" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="statusTagType(row.status)">{{ label(teamStatusMap, row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }"><el-button size="small" @click="$router.push(`/teams/${row.id}`)">详情</el-button></template>
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
import { useSeasonStore } from '@/stores/season'
import { label, teamStatusMap, statusTagType } from '@/utils/status'

const seasonStore = useSeasonStore()
const table = useTable((p) => teamApi.list(p))
const statuses = ['pending', 'approved', 'rejected', 'suspended']
const statusLabels: Record<string, string> = { pending: '待审核', approved: '已批准', rejected: '已拒绝', suspended: '已停赛' }
table.load()
</script>
