<template>
  <div class="page-container" v-loading="loading">
    <el-page-header @back="$router.back()" :content="season?.name || '赛季详情'" style="margin-bottom: 16px" />
    <el-row :gutter="16" v-if="season">
      <el-col :span="16">
        <el-card>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="编码">{{ season.code }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ season.name }}</el-descriptions-item>
            <el-descriptions-item label="项目">{{ season.sport }}</el-descriptions-item>
            <el-descriptions-item label="组别">{{ season.division || '-' }}</el-descriptions-item>
            <el-descriptions-item label="赛制">{{ season.format }}</el-descriptions-item>
            <el-descriptions-item label="队伍数">{{ season.team_count }}</el-descriptions-item>
            <el-descriptions-item label="开始">{{ season.start_date?.slice(0, 10) || '-' }}</el-descriptions-item>
            <el-descriptions-item label="结束">{{ season.end_date?.slice(0, 10) || '-' }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="statusTagType(season.status)">{{ label(seasonStatusMap, season.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="当前规则版本">v{{ season.current_rule_version }}</el-descriptions-item>
          </el-descriptions>
          <div style="margin-top: 16px; display: flex; gap: 8px" v-if="has('seasons:manage')">
            <el-select v-model="newStatus" size="small" style="width: 160px" :placeholder="season.status">
              <el-option v-for="s in statusFlow" :key="s" :label="label(seasonStatusMap, s)" :value="s" />
            </el-select>
            <el-button size="small" type="primary" @click="changeStatus">流转状态</el-button>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header><span>当前积分规则</span></template>
          <div v-if="rule">
            <el-descriptions :column="1" border>
              <el-descriptions-item label="版本">v{{ rule.version }}</el-descriptions-item>
              <el-descriptions-item label="胜">{{ rule.win_points }}</el-descriptions-item>
              <el-descriptions-item label="平">{{ rule.draw_points }}</el-descriptions-item>
              <el-descriptions-item label="负">{{ rule.loss_points }}</el-descriptions-item>
              <el-descriptions-item label="同分">{{ rule.tiebreakers }}</el-descriptions-item>
            </el-descriptions>
            <el-button v-if="has('seasons:manage')" size="small" type="primary" style="margin-top: 12px" @click="ruleDlg = true">修改规则</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="ruleDlg" title="设置积分规则（新版本）" width="480px">
      <el-form :model="ruleForm" label-position="top">
        <el-row :gutter="8">
          <el-col :span="8"><el-form-item label="胜"><el-input-number v-model="ruleForm.win_points" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="平"><el-input-number v-model="ruleForm.draw_points" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="负"><el-input-number v-model="ruleForm.loss_points" :min="0" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="同分规则（逗号分隔）"><el-input v-model="ruleForm.tiebreakers" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="ruleDlg=false">取消</el-button><el-button type="primary" @click="saveRule">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { seasonApi } from '@/api'
import { usePermission } from '@/composables/usePermission'
import { label, seasonStatusMap, statusTagType } from '@/utils/status'

const route = useRoute()
const { has } = usePermission()
const id = Number(route.params.id)
const loading = ref(true)
const season = ref<any>(null)
const rule = ref<any>(null)
const newStatus = ref('')
const ruleDlg = ref(false)
const statusFlow = ['draft', 'registration', 'open', 'ongoing', 'completed', 'archived']
const ruleForm = reactive<any>({ win_points: 3, draw_points: 1, loss_points: 0, tiebreakers: 'points,goal_diff,goals_for,head_to_head,fair_play' })

async function load() {
  loading.value = true
  try {
    season.value = await seasonApi.get(id)
    rule.value = await seasonApi.activeRule(id)
    Object.assign(ruleForm, { win_points: rule.value?.win_points, draw_points: rule.value?.draw_points, loss_points: rule.value?.loss_points, tiebreakers: rule.value?.tiebreakers })
  } finally { loading.value = false }
}
async function changeStatus() {
  if (!newStatus.value) return
  await seasonApi.changeStatus(id, newStatus.value)
  ElMessage.success('状态已更新')
  await load()
}
async function saveRule() {
  await seasonApi.setRule(id, ruleForm)
  ElMessage.success('规则已版本化保存')
  ruleDlg.value = false
  await load()
}
onMounted(load)
</script>
