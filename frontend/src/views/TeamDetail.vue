<template>
  <div class="page-container" v-loading="loading">
    <el-page-header @back="$router.back()" :content="team?.name || '球队详情'" style="margin-bottom: 16px" />
    <el-row :gutter="16" v-if="team">
      <el-col :span="10">
        <el-card>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="编码">{{ team.code }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ team.name }}</el-descriptions-item>
            <el-descriptions-item label="联系人">{{ team.contact || '-' }}</el-descriptions-item>
            <el-descriptions-item label="状态"><el-tag :type="statusTagType(team.status)">{{ label(teamStatusMap, team.status) }}</el-tag></el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card>
          <template #header><span>球员名单</span></template>
          <el-table :data="players" size="small" stripe>
            <el-table-column prop="number" label="号码" width="70" />
            <el-table-column prop="name" label="姓名" />
            <el-table-column prop="position" label="位置" width="90"><template #default="{ row }">{{ label(positionMap, row.position) }}</template></el-table-column>
            <el-table-column prop="status" label="状态" width="90"><template #default="{ row }"><el-tag size="small" :type="statusTagType(row.status)">{{ label(playerStatusMap, row.status) }}</el-tag></template></el-table-column>
            <el-table-column prop="eligibility" label="资格" width="90"><template #default="{ row }">{{ label(eligibilityMap, row.eligibility) }}</template></el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { teamApi, playerApi } from '@/api'
import { label, teamStatusMap, playerStatusMap, eligibilityMap, positionMap, statusTagType } from '@/utils/status'

const route = useRoute()
const id = Number(route.params.id)
const loading = ref(true)
const team = ref<any>(null)
const players = ref<any[]>([])

onMounted(async () => {
  try {
    team.value = await teamApi.get(id)
    const res = await playerApi.list({ team_id: id, page: 1, page_size: 100 })
    players.value = res.list || []
  } finally { loading.value = false }
})
</script>
