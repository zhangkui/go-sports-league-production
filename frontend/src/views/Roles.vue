<template>
  <div class="page-container">
    <div class="page-toolbar">
      <span style="font-weight: 600">角色管理</span>
      <el-button type="primary" @click="openCreate">新建角色</el-button>
    </div>
    <el-card class="table-card">
      <el-table :data="roles" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="code" label="编码" width="160" />
        <el-table-column prop="name" label="名称" width="160" />
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column label="内置" width="80">
          <template #default="{ row }"><el-tag :type="row.is_builtin ? 'info' : 'success'">{{ row.is_builtin ? '是' : '否' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="权限数" width="90">
          <template #default="{ row }">{{ row.permission_codes?.length || 0 }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" :disabled="row.is_builtin" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dlg" :title="editing.id ? '编辑角色' : '新建角色'" width="520px">
      <el-form :model="editing" label-position="top">
        <el-form-item label="编码" v-if="!editing.id"><el-input v-model="editing.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="editing.name" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="editing.description" type="textarea" :rows="2" /></el-form-item>
        <el-form-item label="权限">
          <el-select v-model="editing.permission_ids" multiple filterable style="width: 100%">
            <el-option v-for="p in permissions" :key="p.id" :label="`${p.resource}:${p.action}`" :value="p.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dlg = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { roleApi } from '@/api'

const roles = ref<any[]>([])
const permissions = ref<any[]>([])
const dlg = ref(false)
const editing = reactive<any>({ id: null, code: '', name: '', description: '', permission_ids: [] })

async function load() {
  roles.value = await roleApi.list()
  permissions.value = await roleApi.permissions()
}
function openCreate() {
  Object.assign(editing, { id: null, code: '', name: '', description: '', permission_ids: [] })
  dlg.value = true
}
function openEdit(row: any) {
  Object.assign(editing, { id: row.id, code: row.code, name: row.name, description: row.description, permission_ids: [] })
  dlg.value = true
}
async function save() {
  if (editing.id) await roleApi.update(editing.id, { name: editing.name, description: editing.description, permission_ids: editing.permission_ids })
  else await roleApi.create({ code: editing.code, name: editing.name, description: editing.description, permission_ids: editing.permission_ids })
  ElMessage.success('已保存')
  dlg.value = false
  await load()
}
async function remove(row: any) {
  await ElMessageBox.confirm(`删除角色 ${row.name}？`, '确认', { type: 'warning' })
  await roleApi.remove(row.id)
  ElMessage.success('已删除')
  await load()
}
onMounted(load)
</script>
