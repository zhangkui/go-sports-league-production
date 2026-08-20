<template>
  <div class="page-container">
    <el-card style="max-width: 560px">
      <template #header><span>记录处罚</span></template>
      <el-form :model="form" label-position="top">
        <el-row :gutter="8">
          <el-col :span="8"><el-form-item label="球员ID"><el-input-number v-model="form.player_id" :min="1" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="球队ID"><el-input-number v-model="form.team_id" :min="1" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="赛季ID"><el-input-number v-model="form.season_id" :min="1" style="width: 100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="处罚类型">
          <el-select v-model="form.punishment" style="width: 100%">
            <el-option label="黄牌" value="yellow" /><el-option label="红牌" value="red" /><el-option label="罚款" value="fine" /><el-option label="停赛" value="suspension" /><el-option label="警告" value="warning" />
          </el-select>
        </el-form-item>
        <el-form-item label="原因"><el-input v-model="form.reason" type="textarea" :rows="2" /></el-form-item>
        <el-row :gutter="8">
          <el-col :span="8"><el-form-item label="等级"><el-select v-model="form.severity" style="width: 100%"><el-option label="轻微" value="minor" /><el-option label="中等" value="medium" /><el-option label="严重" value="severe" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="停赛场次"><el-input-number v-model="form.suspend_games" :min="0" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="罚款金额"><el-input-number v-model="form.fine_amount" :min="0" style="width: 100%" /></el-form-item></el-col>
        </el-row>
        <el-button type="primary" :loading="loading" @click="submit">提交</el-button>
        <el-button @click="$router.back()">返回</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { disciplineApi } from '@/api'
const router = useRouter()
const loading = ref(false)
const form = reactive<any>({ player_id: undefined, team_id: undefined, season_id: undefined, punishment: 'yellow', reason: '', severity: 'medium', suspend_games: 0, fine_amount: 0 })
async function submit() {
  if (!form.reason) { ElMessage.warning('请填写处罚原因'); return }
  loading.value = true
  try { await disciplineApi.create(form); ElMessage.success('已记录'); router.push('/disciplines') } finally { loading.value = false }
}
</script>
