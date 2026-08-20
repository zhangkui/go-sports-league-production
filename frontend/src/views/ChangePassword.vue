<template>
  <div class="page-container">
    <el-card style="max-width: 480px">
      <template #header><span>修改密码</span></template>
      <el-form :model="form" label-position="top" ref="formRef" :rules="rules">
        <el-form-item label="原密码" prop="old_password">
          <el-input v-model="form.old_password" type="password" show-password />
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input v-model="form.new_password" type="password" show-password />
        </el-form-item>
        <el-form-item label="确认新密码" prop="confirm">
          <el-input v-model="form.confirm" type="password" show-password />
        </el-form-item>
        <el-button type="primary" :loading="loading" @click="submit">提交</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, type FormInstance } from 'element-plus'
import { authApi } from '@/api'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ old_password: '', new_password: '', confirm: '' })
const rules = {
  old_password: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  new_password: [{ required: true, min: 8, message: '至少8位', trigger: 'blur' }],
  confirm: [{ required: true, validator: (_r: any, v: string, cb: any) => v === form.new_password ? cb() : cb(new Error('两次输入不一致')), trigger: 'blur' }],
}

async function submit() {
  if (!formRef.value) return
  await formRef.value.validate()
  loading.value = true
  try {
    await authApi.changePassword({ old_password: form.old_password, new_password: form.new_password })
    ElMessage.success('密码已修改，请重新登录')
    await auth.logout()
    location.href = '/login'
  } finally {
    loading.value = false
  }
}
</script>
