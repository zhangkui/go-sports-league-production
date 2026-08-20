<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">球员管理</span>
      <el-input v-model="table.query.team_id" placeholder="球队ID" style="width: 140px" @keyup.enter="table.search" />
      <el-select v-model="table.query.status" placeholder="状态" clearable style="width: 120px" @change="table.search">
        <el-option v-for="s in statuses" :key="s" :label="statusLabels[s]" :value="s" />
      </el-select>
      <el-button @click="table.search">搜索</el-button>
      <el-button type="primary" @click="openCreate">添加球员</el-button>
    </div>
    <el-card class="table-card">
      <el-table :data="table.list.value" v-loading="table.loading.value" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="team_id" label="球队ID" width="90" />
        <el-table-column prop="number" label="号码" width="80" />
        <el-table-column prop="name" label="姓名" min-width="120" />
        <el-table-column prop="position" label="位置" width="90">
          <template #default="{ row }">{{ label(positionMap, row.position) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }"><el-tag :type="statusTagType(row.status)" size="small">{{ label(playerStatusMap, row.status) }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="eligibility" label="资格" width="90">
          <template #default="{ row }">{{ label(eligibilityMap, row.eligibility) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }"><el-button size="small" @click="openEdit(row)">编辑</el-button></template>
        </el-table-column>
      </el-table>
      <el-pagination style="margin-top: 12px; justify-content: flex-end; display: flex"
        :current-page="table.query.page" :page-size="table.query.page_size" :total="table.total.value"
        layout="total, prev, pager, next" @current-change="table.onPageChange" />
    </el-card>

    <el-dialog v-model="dlg" :title="editing.id ? '编辑球员' : '添加球员'" width="500px">
      <el-form :model="editing" label-position="top">
        <el-form-item label="球队ID" v-if="!editing.id"><el-input-number v-model="editing.team_id" :min="1" style="width: 100%" /></el-form-item>
        <el-form-item label="姓名"><el-input v-model="editing.name" /></el-form-item>
        <el-form-item label="号码"><el-input-number v-model="editing.number" :min="1" style="width: 100%" /></el-form-item>
        <el-form-item label="位置">
          <el-select v-model="editing.position" style="width: 100%" clearable><el-option label="门将 GK" value="GK" /><el-option label="后卫 DEF" value="DEF" /><el-option label="中场 MID" value="MID" /><el-option label="前锋 FWD" value="FWD" /></el-select>
        </el-form-item>
        <el-form-item label="出生日期"><el-date-picker v-model="editing.birth_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" /></el-form-item>
        <el-row :gutter="8">
          <el-col :span="12"><el-form-item label="身高(cm)"><el-input-number v-model="editing.height_cm" :min="0" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="体重(kg)"><el-input-number v-model="editing.weight_kg" :min="0" style="width: 100%" /></el-form-item></el-col>
        </el-row>
      </el-form>
      <template #footer><el-button @click="dlg=false">取消</el-button><el-button type="primary" @click="save">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { playerApi } from '@/api'
import { useTable } from '@/composables/useTable'
import { label, playerStatusMap, eligibilityMap, positionMap, statusTagType } from '@/utils/status'

const table = useTable((p) => playerApi.list(p))
const statuses = ['active', 'suspended', 'injured', 'transferred', 'retired']
const statusLabels: Record<string, string> = { active: '活跃', suspended: '停赛', injured: '受伤', transferred: '已转会', retired: '退役' }
const dlg = ref(false)
const editing = reactive<any>({ id: null, team_id: undefined, name: '', number: 1, position: '', birth_date: '', height_cm: undefined, weight_kg: undefined })

function openCreate() { Object.assign(editing, { id: null, team_id: undefined, name: '', number: 1, position: '', birth_date: '', height_cm: undefined, weight_kg: undefined }); dlg.value = true }
function openEdit(row: any) { Object.assign(editing, row); dlg.value = true }
async function save() {
  if (editing.id) await playerApi.update(editing.id, editing)
  else await playerApi.create(editing)
  ElMessage.success('已保存')
  dlg.value = false
  await table.load()
}
table.load()
</script>
