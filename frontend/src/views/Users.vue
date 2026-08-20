<template>
  <div class="page-container">
    <div class="page-toolbar">
      <el-input v-model="table.query.q" placeholder="搜索用户名/邮箱" style="width: 240px" clearable @clear="table.search" @keyup.enter="table.search" />
      <el-button @click="table.search">搜索</el-button>
    </div>
    <el-card class="table-card">
      <el-table :data="table.list.value" v-loading="table.loading.value" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="username" label="用户名" width="140" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column prop="full_name" label="姓名" width="120" />
        <el-table-column label="角色" min-width="160">
          <template #default="{ row }">
            <el-tag v-for="r in row.roles" :key="r.id" style="margin-right: 4px">{{ r.code }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openRoles(row)">分配角色</el-button>
            <el-button size="small" :type="row.status === 1 ? 'danger' : 'success'" @click="toggle(row)">{{ row.status === 1 ? '停用' : '启用' }}</el-button>
            <el-button size="small" @click="openReset(row)">重置密码</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination style="margin-top: 12px; justify-content: flex-end; display: flex"
        :current-page="table.query.page" :page-size="table.query.page_size" :total="table.total.value"
        layout="total, prev, pager, next, sizes" @current-change="table.onPageChange" @size-change="table.onSizeChange" />
    </el-card>

    <el-dialog v-model="roleDlg" title="分配角色" width="420px">
      <el-select v-model="selectedRoles" multiple placeholder="选择角色" style="width: 100%">
        <el-option v-for="r in roles" :key="r.id" :label="`${r.code} - ${r.name}`" :value="r.id" />
      </el-select>
      <template #footer>
        <el-button @click="roleDlg = false">取消</el-button>
        <el-button type="primary" @click="saveRoles">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="resetDlg" title="重置密码" width="420px">
      <el-input v-model="newPassword" type="password" show-password placeholder="新密码（至少8位字母+数字）" />
      <template #footer>
        <el-button @click="resetDlg = false">取消</el-button>
        <el-button type="primary" @click="doReset">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { userApi, roleApi } from '@/api'
import { useTable } from '@/composables/useTable'

const table = useTable((p) => userApi.list(p))
const roles = ref<any[]>([])
const roleDlg = ref(false)
const resetDlg = ref(false)
const selectedUser = ref<any>(null)
const selectedRoles = ref<number[]>([])
const newPassword = ref('')

async function loadRoles() { roles.value = await roleApi.list() }
function openRoles(row: any) {
  selectedUser.value = row
  selectedRoles.value = (row.roles || []).map((r: any) => r.id)
  roleDlg.value = true
}
async function saveRoles() {
  await table.confirm(() => userApi.assignRoles(selectedUser.value.id, selectedRoles.value), '角色已分配')
  roleDlg.value = false
}
function openReset(row: any) { selectedUser.value = row; newPassword.value = ''; resetDlg.value = true }
async function doReset() {
  await table.confirm(() => userApi.resetPassword(selectedUser.value.id, { new_password: newPassword.value }), '密码已重置')
  resetDlg.value = false
}
async function toggle(row: any) {
  await table.confirm(() => userApi.enable(row.id, row.status !== 1), '状态已更新')
}

onMounted(() => { loadRoles(); table.load() })
</script>
