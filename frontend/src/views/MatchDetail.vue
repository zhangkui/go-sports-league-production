<template>
  <div class="page-container" v-loading="loading">
    <el-page-header @back="$router.back()" content="比赛详情" style="margin-bottom: 16px" />
    <el-row :gutter="16" v-if="match">
      <el-col :span="16">
        <el-card>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="日期">{{ match.match_date }}</el-descriptions-item>
            <el-descriptions-item label="时间">{{ match.start_time }}</el-descriptions-item>
            <el-descriptions-item label="主队">#{{ match.home_team_id }}</el-descriptions-item>
            <el-descriptions-item label="客队">#{{ match.away_team_id }}</el-descriptions-item>
            <el-descriptions-item label="比分">{{ match.home_score ?? '-' }} : {{ match.away_score ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="半场">{{ match.home_half_score ?? '-' }} : {{ match.away_half_score ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="状态"><el-tag :type="statusTagType(match.status)">{{ label(matchStatusMap, match.status) }}</el-tag></el-descriptions-item>
            <el-descriptions-item label="确认"><el-tag :type="match.confirm_status === 'confirmed' ? 'success' : 'warning'">{{ label(confirmStatusMap, match.confirm_status) }}</el-tag></el-descriptions-item>
          </el-descriptions>
          <div style="margin-top: 16px; display: flex; gap: 8px; flex-wrap: wrap" v-if="hasAny('matches:record','matches:confirm')">
            <el-button size="small" @click="recordDlg = true">录入比分</el-button>
            <el-button size="small" type="success" :disabled="match.confirm_status === 'confirmed'" @click="confirm">确认比赛</el-button>
          </div>
        </el-card>

        <el-card style="margin-top: 16px">
          <template #header><span>比赛事件</span></template>
          <el-table :data="events" size="small" stripe>
            <el-table-column prop="minute" label="分钟" width="80" />
            <el-table-column prop="event_type" label="类型" width="100">
              <template #default="{ row }"><el-tag size="small" :type="row.event_type === 'red' ? 'danger' : row.event_type === 'yellow' ? 'warning' : 'primary'">{{ label(eventTypeMap, row.event_type) }}</el-tag></template>
            </el-table-column>
            <el-table-column prop="team_id" label="球队ID" width="90" />
            <el-table-column prop="player_id" label="球员ID" width="90" />
            <el-table-column prop="description" label="描述" min-width="180" />
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header><span>上场统计</span></template>
          <el-table :data="stats" size="small">
            <el-table-column prop="player_id" label="球员ID" width="90" />
            <el-table-column prop="is_starter" label="首发" width="70"><template #default="{ row }">{{ row.is_starter ? '是' : '否' }}</template></el-table-column>
            <el-table-column prop="played_min" label="上场" width="70" />
            <el-table-column prop="rating" label="评分" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="recordDlg" title="录入比分" width="420px">
      <el-form :model="rec" label-position="top">
        <el-row :gutter="8">
          <el-col :span="12"><el-form-item label="主队得分"><el-input-number v-model="rec.home_score" :min="0" style="width: 100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="客队得分"><el-input-number v-model="rec.away_score" :min="0" style="width: 100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="比赛时长(分)"><el-input-number v-model="rec.duration_min" :min="0" style="width: 100%" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="recordDlg=false">取消</el-button><el-button type="primary" @click="saveRecord">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { matchApi } from '@/api'
import { usePermission } from '@/composables/usePermission'
import { label, matchStatusMap, confirmStatusMap, eventTypeMap, statusTagType } from '@/utils/status'

const route = useRoute()
const { hasAny } = usePermission()
const id = Number(route.params.id)
const loading = ref(true)
const match = ref<any>(null)
const events = ref<any[]>([])
const stats = ref<any[]>([])
const recordDlg = ref(false)
const rec = reactive<any>({ home_score: 0, away_score: 0, duration_min: 90, status: 'completed' })

async function load() {
  loading.value = true
  try {
    match.value = await matchApi.get(id)
    events.value = await matchApi.events(id)
    stats.value = await matchApi.stats(id)
    rec.home_score = match.value?.home_score ?? 0
    rec.away_score = match.value?.away_score ?? 0
    rec.duration_min = match.value?.duration_min ?? 90
  } finally { loading.value = false }
}
async function saveRecord() {
  await matchApi.record(id, rec)
  ElMessage.success('比分已录入')
  recordDlg.value = false
  await load()
}
async function confirm() {
  await matchApi.confirm(id)
  ElMessage.success('比赛已确认，积分已更新')
  await load()
}
onMounted(load)
</script>
