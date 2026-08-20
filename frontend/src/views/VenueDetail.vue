<template>
  <div class="page-container" v-loading="loading">
    <el-page-header @back="$router.back()" :content="venue?.name || '场地详情'" style="margin-bottom: 16px" />
    <el-row :gutter="16" v-if="venue">
      <el-col :span="10">
        <el-card>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="编码">{{ venue.code }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ venue.name }}</el-descriptions-item>
            <el-descriptions-item label="地址">{{ venue.address || '-' }}</el-descriptions-item>
            <el-descriptions-item label="容量">{{ venue.capacity }}</el-descriptions-item>
            <el-descriptions-item label="项目">{{ venue.sport || '-' }}</el-descriptions-item>
            <el-descriptions-item label="状态"><el-tag>{{ venue.status }}</el-tag></el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="14">
        <el-card>
          <template #header><span>可用时间段（每周）</span></template>
          <el-table :data="slots" size="small">
            <el-table-column prop="weekday" label="星期" width="90"><template #default="{ row }">{{ weekdayName(row.weekday) }}</template></el-table-column>
            <el-table-column prop="start_time" label="开始" />
            <el-table-column prop="end_time" label="结束" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { venueApi } from '@/api'
const route = useRoute()
const id = Number(route.params.id)
const loading = ref(true)
const venue = ref<any>(null)
const slots = ref<any[]>([])
const wd = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
function weekdayName(n: number) { return wd[n] || n }
onMounted(async () => {
  try { venue.value = await venueApi.get(id); slots.value = await venueApi.availability(id) } finally { loading.value = false }
})
</script>
