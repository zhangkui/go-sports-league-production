<template>
  <div class="page-container">
    <el-card style="max-width: 640px">
      <template #header><span>创建赛季</span></template>
      <el-form :model="form" label-position="top">
        <el-row :gutter="12">
          <el-col :span="12"><el-form-item label="编码"><el-input v-model="form.code" placeholder="如 SEASON-2026-A" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="名称"><el-input v-model="form.name" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="运动项目">
            <el-select v-model="form.sport" style="width: 100%"><el-option label="篮球" value="basketball" /><el-option label="足球" value="football" /><el-option label="排球" value="volleyball" /></el-select>
          </el-form-item></el-col>
          <el-col :span="12"><el-form-item label="组别"><el-input v-model="form.division" placeholder="甲级/乙级" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="赛制">
            <el-select v-model="form.format" style="width: 100%">
              <el-option label="单循环" value="round_robin" /><el-option label="双循环" value="double_round" /><el-option label="混合" value="mixed" /><el-option label="淘汰赛" value="knockout" />
            </el-select>
          </el-form-item></el-col>
          <el-col :span="12"><el-form-item label="预计队伍数"><el-input-number v-model="form.team_count" :min="0" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="开始日期"><el-date-picker v-model="form.start_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="结束日期"><el-date-picker v-model="form.end_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="报名开始"><el-date-picker v-model="form.registration_start" type="date" value-format="YYYY-MM-DD" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="报名截止"><el-date-picker v-model="form.registration_end" type="date" value-format="YYYY-MM-DD" style="width: 100%" /></el-form-item></el-col>
        </el-row>
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
import { seasonApi } from '@/api'

const router = useRouter()
const loading = ref(false)
const form = reactive<any>({ code: '', name: '', sport: 'football', division: '', format: 'round_robin', team_count: 8, start_date: '', end_date: '', registration_start: '', registration_end: '' })

async function submit() {
  if (!form.code || !form.name) { ElMessage.warning('编码和名称必填'); return }
  loading.value = true
  try {
    await seasonApi.create(form)
    ElMessage.success('赛季已创建')
    router.push('/seasons')
  } finally { loading.value = false }
}
</script>
