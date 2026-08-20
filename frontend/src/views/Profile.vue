<template>
  <div class="page-container">
    <el-card>
      <template #header><span>个人资料</span></template>
      <el-descriptions :column="2" border v-if="auth.user">
        <el-descriptions-item label="用户名">{{ auth.user.username }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ auth.user.email }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ auth.user.full_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="auth.user.status === 1 ? 'success' : 'danger'">{{ auth.user.status === 1 ? '启用' : '停用' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="角色" :span="2">
          <el-tag v-for="r in auth.user.roles" :key="r.id" style="margin-right: 6px">{{ r.name }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="权限" :span="2">
          <el-tag v-for="p in (auth.user.permissions || []).slice(0, 24)" :key="p" type="info" style="margin: 2px">{{ p }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
const auth = useAuthStore()
</script>
