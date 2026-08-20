<template>
  <div class="page-container">
    <div class="page-toolbar">
      <el-select v-model="table.query.status" placeholder="状态" clearable style="width: 140px" @change="table.search">
        <el-option v-for="s in statuses" :key="s" :label="statusLabels[s]" :value="s" />
      </el-select>
      <el-button @click="table.search">刷新</el-button>
      <el-button type="primary" v-if="has('seasons:manage')" @click="$router.push('/seasons/create')">创建赛季</el-button>
    </div>
    <el-card class="table-card">
      <el-table :data="table.list.value" v-loading="table.loading.value" stripe>
        <el-table-column prop="code" label="编码" width="120" />
        <el-table-column prop="name" label="名称" min-width="160" />
        <el-table-column prop="sport" label="项目" width="90">
          <template #default="{ row }">{{ label(sportMap, row.sport) }}</template>
        </el-table-column>
        <el-table-column prop="division" label="组别" width="90" />
        <el-table-column prop="format" label="赛制" width="120">
          <template #default="{ row }">{{ label(formatMap, row.format) }}</template>
        </el-table-column>
        <el-table-column prop="team_count" label="队伍数" width="80" />
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }"><el-tag :type="statusTagType(row.status)">{{ label(seasonStatusMap, row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="current_rule_version" label="规则版本" width="90" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }"><el-button size="small" @click="$router.push(`/seasons/${row.id}`)">详情</el-button></template>
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
import { usePermission } from '@/composables/usePermission'
import { seasonApi } from '@/api'
import { label, seasonStatusMap, sportMap, formatMap, statusTagType } from '@/utils/status'

const table = useTable((p) => seasonApi.list(p))
const { has } = usePermission()
const statuses = ['draft', 'registration', 'open', 'ongoing', 'completed', 'archived']
const statusLabels: Record<string, string> = {
  draft: '草稿', registration: '报名中', open: '开放', ongoing: '进行中', completed: '已完成', archived: '已归档',
}
table.load()
</script>
