<template>
  <div class="login-wrap">
    <div class="login-card">
      <h2 class="login-title">业余体育联赛管理系统</h2>
      <el-form :model="form" label-position="top" @submit.prevent="onLogin">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="请输入用户名" :prefix-icon="User" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" :prefix-icon="Lock" @keyup.enter="onLogin" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="auth.loading" style="width: 100%" @click="onLogin">登录</el-button>
        </el-form-item>
        <div style="display: flex; justify-content: space-between">
          <router-link to="/register" style="font-size: 13px">没有账号？注册</router-link>
          <span class="muted" style="font-size: 12px">默认管理员 admin / Admin123!</span>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { User, Lock } from '@element-plus/icons-vue'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const form = reactive({ username: 'admin', password: 'Admin123!' })

async function onLogin() {
  await auth.login(form.username, form.password)
  const redirect = (route.query.redirect as string) || '/dashboard'
  router.push(redirect)
}
</script>
