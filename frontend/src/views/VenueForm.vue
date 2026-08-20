<template>
  <div class="page-container">
    <el-card style="max-width: 540px">
      <template #header><span>添加场地</span></template>
      <el-form :model="form" label-position="top">
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="地址"><el-input v-model="form.address" /></el-form-item>
        <el-row :gutter="8">
          <el-col :span="12"><el-form-item label="容量"><el-input-number v-model="form.capacity" :min="0" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="项目"><el-input v-model="form.sport" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%"><el-option label="可用" value="available" /><el-option label="维护" value="maintenance" /><el-option label="不可用" value="unavailable" /></el-select>
        </el-form-item>
        <el-button type="primary" :loading="loading" @click="submit">创建</el-button>
        <el-button @click="$router.back()">返回</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { venueApi } from '@/api'
const router = useRouter()
const loading = ref(false)
const form = reactive<any>({ code: '', name: '', address: '', capacity: 0, sport: '', status: 'available' })
async function submit() {
  if (!form.code || !form.name) { ElMessage.warning('编码和名称必填'); return }
  loading.value = true
  try { await venueApi.create(form); ElMessage.success('已创建'); router.push('/venues') } finally { loading.value = false }
}
</script>
