<template>
  <div class="page-container">
    <el-card style="max-width: 540px">
      <template #header><span>创建球队</span></template>
      <el-form :model="form" label-position="top">
        <el-form-item label="编码"><el-input v-model="form.code" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="所属赛季">
          <el-select v-model="form.season_id" style="width: 100%">
            <el-option v-for="s in seasonStore.seasons" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="联系人"><el-input v-model="form.contact" /></el-form-item>
        <el-form-item label="队徽URL"><el-input v-model="form.logo_url" /></el-form-item>
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
import { teamApi } from '@/api'
import { useSeasonStore } from '@/stores/season'

const router = useRouter()
const seasonStore = useSeasonStore()
const loading = ref(false)
const form = reactive<any>({ code: '', name: '', season_id: undefined, contact: '', logo_url: '' })

async function submit() {
  if (!form.code || !form.name || !form.season_id) { ElMessage.warning('编码、名称、赛季必填'); return }
  loading.value = true
  try { await teamApi.create(form); ElMessage.success('已创建'); router.push('/teams') } finally { loading.value = false }
}
</script>
