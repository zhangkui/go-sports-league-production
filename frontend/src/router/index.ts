import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { tokenStore } from '@/api/client'

const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'login', component: () => import('@/views/auth/Login.vue'), meta: { public: true } },
  { path: '/register', name: 'register', component: () => import('@/views/auth/Register.vue'), meta: { public: true } },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', name: 'dashboard', component: () => import('@/views/Dashboard.vue') },
      { path: 'profile', name: 'profile', component: () => import('@/views/Profile.vue') },
      { path: 'profile/password', name: 'profile-password', component: () => import('@/views/ChangePassword.vue') },
      { path: 'users', name: 'users', component: () => import('@/views/Users.vue'), meta: { perm: 'users:manage' } },
      { path: 'roles', name: 'roles', component: () => import('@/views/Roles.vue'), meta: { perm: 'users:manage' } },
      { path: 'permissions', name: 'permissions', component: () => import('@/views/Permissions.vue'), meta: { perm: 'users:manage' } },
      { path: 'seasons', name: 'seasons', component: () => import('@/views/Seasons.vue') },
      { path: 'seasons/create', name: 'season-create', component: () => import('@/views/SeasonForm.vue'), meta: { perm: 'seasons:manage' } },
      { path: 'seasons/:id', name: 'season-detail', component: () => import('@/views/SeasonDetail.vue') },
      { path: 'teams', name: 'teams', component: () => import('@/views/Teams.vue') },
      { path: 'teams/create', name: 'team-create', component: () => import('@/views/TeamForm.vue') },
      { path: 'teams/:id', name: 'team-detail', component: () => import('@/views/TeamDetail.vue') },
      { path: 'teams/registrations', name: 'team-registrations', component: () => import('@/views/TeamRegistrations.vue') },
      { path: 'transfers', name: 'transfers', component: () => import('@/views/Transfers.vue') },
      { path: 'players', name: 'players', component: () => import('@/views/Players.vue') },
      { path: 'venues', name: 'venues', component: () => import('@/views/Venues.vue') },
      { path: 'venues/create', name: 'venue-create', component: () => import('@/views/VenueForm.vue'), meta: { perm: 'venues:manage' } },
      { path: 'venues/:id', name: 'venue-detail', component: () => import('@/views/VenueDetail.vue') },
      { path: 'schedules', name: 'schedules', component: () => import('@/views/Schedules.vue') },
      { path: 'schedules/calendar', name: 'schedules-calendar', component: () => import('@/views/ScheduleCalendar.vue') },
      { path: 'schedules/conflicts', name: 'schedules-conflicts', component: () => import('@/views/ScheduleConflicts.vue') },
      { path: 'matches', name: 'matches', component: () => import('@/views/Matches.vue') },
      { path: 'matches/:id', name: 'match-detail', component: () => import('@/views/MatchDetail.vue') },
      { path: 'disciplines', name: 'disciplines', component: () => import('@/views/Disciplines.vue') },
      { path: 'disciplines/create', name: 'discipline-create', component: () => import('@/views/DisciplineForm.vue'), meta: { perm: 'disciplines:manage' } },
      { path: 'appeals', name: 'appeals', component: () => import('@/views/Appeals.vue') },
      { path: 'standings', name: 'standings', component: () => import('@/views/Standings.vue') },
      { path: 'standings/history', name: 'standings-history', component: () => import('@/views/StandingsHistory.vue') },
      { path: 'reports/season', name: 'report-season', component: () => import('@/views/SeasonReport.vue') },
      { path: 'reports/players', name: 'report-players', component: () => import('@/views/PlayerReport.vue') },
      { path: 'audit-logs', name: 'audit-logs', component: () => import('@/views/AuditLogs.vue'), meta: { perm: 'audit_logs:view' } },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.meta.public) return true
  if (!tokenStore.access) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (!auth.isAuthed) {
    await auth.fetchMe()
  }
  if (to.meta.perm && !auth.has(to.meta.perm as string)) {
    return { name: 'dashboard' }
  }
  return true
})
