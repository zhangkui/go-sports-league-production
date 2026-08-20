<template>
  <div class="login-wrap">
    <div class="login-card">
      <h2 class="login-title">注册账号</h2>
      <el-form :model="form" label-position="top" :rules="rules" ref="formRef" @submit.prevent="onSubmit">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="3-32位字母数字_.-" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="example@league.local" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password placeholder="至少8位字母+数字" />
        </el-form-item>
        <el-form-item label="姓名（可选）">
          <el-input v-model="form.full_name" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" style="width: 100%" @click="onSubmit">注册</el-button>
        </el-form-item>
        <router-link to="/login" style="font-size: 13px">已有账号？登录</router-link>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance } from 'element-plus'
import { authApi } from '@/api'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ username: '', email: '', password: '', full_name: '' })
const rules = {
  username: [{ required: true, min: 3, max: 32, message: '用户名 3-32 位', trigger: 'blur' }],
  email: [{ required: true, type: 'email' as const, message: '邮箱格式不正确', trigger: 'blur' }],
  password: [{ required: true, min: 8, message: '密码至少 8 位', trigger: 'blur' }],
}

async function onSubmit() {
  if (!formRef.value) return
  await formRef.value.validate()
  loading.value = true
  try {
    await authApi.register(form)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } finally {
    loading.value = false
  }
}
</script>
