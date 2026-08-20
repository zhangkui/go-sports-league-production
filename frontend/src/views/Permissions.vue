<template>
  <div class="page-container">
    <el-card class="table-card">
      <template #header><span>权限清单（只读）</span></template>
      <el-table :data="perms" stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="resource" label="资源" width="160" />
        <el-table-column prop="action" label="操作" width="120" />
        <el-table-column prop="name" label="说明" min-width="200" />
        <el-table-column label="编码" width="200">
          <template #default="{ row }"><el-tag>{{ row.resource }}:{{ row.action }}</el-tag></template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { roleApi } from '@/api'
const perms = ref<any[]>([])
onMounted(async () => { perms.value = await roleApi.permissions() })
</script>
