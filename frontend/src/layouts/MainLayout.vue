<template>
  <el-container class="main-layout" style="height: 100vh">
    <el-aside :width="collapsed ? '64px' : '220px'" class="aside">
      <div class="logo">
        <span v-if="!collapsed">业余联赛系统</span>
        <span v-else>GSL</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="collapsed"
        :collapse-transition="false"
        router
        background-color="#001529"
        text-color="#c9d1d9"
        active-text-color="#409eff"
      >
        <template v-for="item in visibleMenus" :key="item.path">
          <el-sub-menu v-if="item.children" :index="item.path">
            <template #title>
              <el-icon><component :is="item.icon" /></el-icon>
              <span>{{ item.title }}</span>
            </template>
            <el-menu-item
              v-for="child in item.children"
              :key="child.path"
              :index="child.path"
              :disabled="child.perm && !has(child.perm)"
            >
              <span>{{ child.title }}</span>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item v-else :index="item.path" :disabled="item.perm && !has(item.perm)">
            <el-icon><component :is="item.icon" /></el-icon>
            <template #title>{{ item.title }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-icon class="toggle" @click="collapsed = !collapsed"><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
          <span class="season-picker" v-if="seasonStore.seasons.length">
            <span class="muted">当前赛季：</span>
            <el-select
              v-model="seasonId"
              size="small"
              style="width: 200px"
              placeholder="选择赛季"
              @change="onSeasonChange"
            >
              <el-option v-for="s in seasonStore.seasons" :key="s.id" :label="s.name" :value="s.id" />
            </el-select>
          </span>
        </div>
        <el-dropdown @command="onCommand">
          <span class="user-chip">
            <el-icon><User /></el-icon>
            {{ auth.user?.username || 'guest' }}
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人资料</el-dropdown-item>
              <el-dropdown-item command="password">修改密码</el-dropdown-item>
              <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSeasonStore } from '@/stores/season'
import {
  HomeFilled, User, UserFilled, Trophy, Setting, Location, Calendar,
  Football, Warning, DataLine, Document, Fold, Expand, ArrowDown, Tickets,
} from '@element-plus/icons-vue'

const auth = useAuthStore()
const seasonStore = useSeasonStore()
const router = useRouter()
const route = useRoute()
const collapsed = ref(false)
const seasonId = ref<number | undefined>(undefined)

const menus = [
  { path: '/dashboard', title: '仪表盘', icon: HomeFilled },
  { path: '/seasons', title: '赛季管理', icon: Trophy },
  { path: '/teams', title: '球队球员', icon: UserFilled, children: [
    { path: '/teams', title: '球队列表' },
    { path: '/teams/registrations', title: '注册审核' },
    { path: '/players', title: '球员管理' },
    { path: '/transfers', title: '转会管理' },
  ]},
  { path: '/venues', title: '场地管理', icon: Location },
  { path: '/schedules', title: '赛程管理', icon: Calendar, children: [
    { path: '/schedules', title: '赛程列表' },
    { path: '/schedules/calendar', title: '日历视图' },
    { path: '/schedules/conflicts', title: '冲突检测' },
  ]},
  { path: '/matches', title: '比赛管理', icon: Football },
  { path: '/disciplines', title: '纪律管理', icon: Warning, children: [
    { path: '/disciplines', title: '处罚记录' },
    { path: '/appeals', title: '申诉管理' },
  ]},
  { path: '/standings', title: '积分排名', icon: DataLine, children: [
    { path: '/standings', title: '实时积分榜' },
    { path: '/standings/history', title: '历史快照' },
  ]},
  { path: '/reports/season', title: '报表', icon: Document, children: [
    { path: '/reports/season', title: '赛季汇总' },
    { path: '/reports/players', title: '球员排行' },
  ]},
  { path: '/users', title: '用户权限', icon: Setting, perm: 'users:manage', children: [
    { path: '/users', title: '用户管理' },
    { path: '/roles', title: '角色管理' },
    { path: '/permissions', title: '权限管理' },
  ]},
  { path: '/audit-logs', title: '审计日志', icon: Tickets, perm: 'audit_logs:view' },
] as any[]

const has = (perm: string) => auth.has(perm)
const visibleMenus = computed(() => menus)
const activeMenu = computed(() => route.path)

function onSeasonChange(id: number) {
  const s = seasonStore.seasons.find((x) => x.id === id)
  if (s) seasonStore.setCurrent(s)
}

async function onCommand(cmd: string) {
  if (cmd === 'profile') router.push('/profile')
  else if (cmd === 'password') router.push('/profile/password')
  else if (cmd === 'logout') {
    await auth.logout()
    router.push('/login')
  }
}

onMounted(async () => {
  await auth.fetchMe()
  await seasonStore.loadAll()
  seasonStore.restore()
  seasonId.value = seasonStore.current?.id
})
</script>

<style scoped>
.aside { background: #001529; transition: width 0.2s; overflow: hidden; }
.logo { height: 56px; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 16px; font-weight: 700; }
.header { display: flex; align-items: center; justify-content: space-between; background: #fff; border-bottom: 1px solid #ebeef5; }
.header-left { display: flex; align-items: center; gap: 16px; }
.toggle { cursor: pointer; font-size: 20px; }
.user-chip { display: flex; align-items: center; gap: 6px; cursor: pointer; }
.main { background: var(--gsl-bg); }
.el-menu { border-right: none; }
.muted { color: #909399; font-size: 13px; }
</style>
