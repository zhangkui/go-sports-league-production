<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">处罚记录</span>
      <el-input v-model="table.query.player_id" placeholder="球员ID" style="width: 120px" @keyup.enter="table.search" />
      <el-button @click="table.search">刷新</el-button>
      <el-button type="primary" v-if="has('disciplines:manage')" @click="$router.push('/disciplines/create')">记录处罚</el-button>
    </div>
    <el-card class="table-card">
      <el-table :data="table.list.value" v-loading="table.loading.value" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="player_id" label="球员" width="80" />
        <el-table-column prop="punishment" label="处罚" width="90">
          <template #default="{ row }">{{ label(punishmentMap, row.punishment) }}</template>
        </el-table-column>
        <el-table-column prop="severity" label="等级" width="90">
          <template #default="{ row }">{{ label(severityMap, row.severity) }}</template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" min-width="200" />
        <el-table-column prop="suspend_games" label="停赛场次" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }"><el-tag :type="statusTagType(row.status)" size="small">{{ label(disciplineStatusMap, row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 'active'" size="small" type="warning" @click="overturn(row)">撤销</el-button>
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
import { disciplineApi } from '@/api'
import { usePermission } from '@/composables/usePermission'
import { label, punishmentMap, severityMap, disciplineStatusMap, statusTagType } from '@/utils/status'
const table = useTable((p) => disciplineApi.list(p))
const { has } = usePermission()
async function overturn(row: any) { await table.confirm(() => disciplineApi.overturn(row.id), '已撤销') }
table.load()
</script>
