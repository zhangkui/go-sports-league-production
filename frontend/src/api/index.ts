import { get, getList, post, put, del } from './client'

// ===== Auth =====
export const authApi = {
  register: (body: any) => post('/auth/register', body),
  login: (body: { username: string; password: string }) => post('/auth/login', body),
  refresh: (refresh_token: string) => post('/auth/refresh', { refresh_token }),
  logout: (refresh_token: string) => post('/auth/logout', { refresh_token }),
  me: () => get('/me'),
  changePassword: (body: { old_password: string; new_password: string }) => put('/me/password', body),
}

// ===== Users & Roles =====
export const userApi = {
  list: (params: any) => getList('/users', { params }),
  get: (id: number) => get(`/users/${id}`),
  update: (id: number, body: any) => put(`/users/${id}`, body),
  enable: (id: number, enabled: boolean) => put(`/users/${id}/enable?enabled=${enabled}`),
  assignRoles: (id: number, role_ids: number[]) => put(`/users/${id}/roles`, { role_ids }),
  resetPassword: (id: number, body: { new_password: string }) => post(`/users/${id}/password`, body),
}
export const roleApi = {
  list: () => get('/roles'),
  get: (id: number) => get(`/roles/${id}`),
  create: (body: any) => post('/roles', body),
  update: (id: number, body: any) => put(`/roles/${id}`, body),
  remove: (id: number) => del(`/roles/${id}`),
  permissions: () => get('/permissions'),
}

// ===== Seasons & Rules =====
export const seasonApi = {
  list: (params: any) => getList('/seasons', { params }),
  get: (id: number) => get(`/seasons/${id}`),
  create: (body: any) => post('/seasons', body),
  update: (id: number, body: any) => put(`/seasons/${id}`, body),
  changeStatus: (id: number, status: string) => put(`/seasons/${id}/status`, { status }),
  activeRule: (id: number) => get(`/seasons/${id}/rules`),
  setRule: (id: number, body: any) => post(`/seasons/${id}/rules`, body),
  ruleVersions: (id: number) => get(`/seasons/${id}/rules/versions`),
}

// ===== Teams & Players =====
export const teamApi = {
  list: (params: any) => getList('/teams', { params }),
  get: (id: number) => get(`/teams/${id}`),
  create: (body: any) => post('/teams', body),
  update: (id: number, body: any) => put(`/teams/${id}`, body),
  review: (id: number, body: any) => put(`/teams/${id}/review`, body),
  registrations: (params: any) => getList('/team-registrations', { params }),
}
export const playerApi = {
  list: (params: any) => getList('/players', { params }),
  get: (id: number) => get(`/players/${id}`),
  create: (body: any) => post('/players', body),
  update: (id: number, body: any) => put(`/players/${id}`, body),
}
export const transferApi = {
  list: (params: any) => getList('/transfers', { params }),
  create: (body: any) => post('/transfers', body),
  review: (id: number, body: any) => put(`/transfers/${id}/review`, body),
}

// ===== Venues =====
export const venueApi = {
  list: (params: any) => getList('/venues', { params }),
  get: (id: number) => get(`/venues/${id}`),
  create: (body: any) => post('/venues', body),
  update: (id: number, body: any) => put(`/venues/${id}`, body),
  remove: (id: number) => del(`/venues/${id}`),
  availability: (id: number) => get(`/venues/${id}/availability`),
  setAvailability: (id: number, slots: any[]) => put(`/venues/${id}/availability`, slots),
}

// ===== Schedules =====
export const scheduleApi = {
  list: (params: any) => getList('/schedules', { params }),
  get: (id: number) => get(`/schedules/${id}`),
  generate: (body: any) => post('/schedules/generate', body),
  update: (id: number, body: any) => put(`/schedules/${id}`, body),
  conflicts: (season_id?: number) => get('/schedules/conflicts', { params: { season_id } }),
}

// ===== Matches =====
export const matchApi = {
  list: (params: any) => getList('/matches', { params }),
  get: (id: number) => get(`/matches/${id}`),
  record: (id: number, body: any) => put(`/matches/${id}`, body),
  setStatus: (id: number, status: string) => put(`/matches/${id}/status?status=${status}`),
  confirm: (id: number) => put(`/matches/${id}/confirm`),
  dispute: (id: number, reason: string) => put(`/matches/${id}/dispute?reason=${encodeURIComponent(reason)}`),
  events: (id: number) => get(`/matches/${id}/events`),
  createEvent: (id: number, body: any) => post(`/matches/${id}/events`, body),
  updateEvent: (id: number, eid: number, body: any) => put(`/matches/${id}/events/${eid}`, body),
  deleteEvent: (id: number, eid: number) => del(`/matches/${id}/events/${eid}`),
  stats: (id: number) => get(`/matches/${id}/stats`),
  upsertStat: (id: number, body: any) => post(`/matches/${id}/stats`, body),
}

// ===== Discipline =====
export const disciplineApi = {
  list: (params: any) => getList('/disciplines', { params }),
  get: (id: number) => get(`/disciplines/${id}`),
  create: (body: any) => post('/disciplines', body),
  overturn: (id: number) => put(`/disciplines/${id}/overturn`),
  playerHistory: (id: number, params: any) => getList(`/players/${id}/disciplines`, { params }),
}
export const appealApi = {
  list: (params: any) => getList('/appeals', { params }),
  create: (body: any) => post('/appeals', body),
  review: (id: number, body: any) => put(`/appeals/${id}/review`, body),
}

// ===== Standings & Reports =====
export const standingApi = {
  list: (season_id: number) => get('/standings', { params: { season_id } }),
  snapshots: (season_id: number) => get('/standings/snapshots', { params: { season_id } }),
  snapshot: (id: number) => get(`/standings/snapshots/${id}`),
}
export const reportApi = {
  seasonSummary: (season_id: number) => get('/reports/season-summary', { params: { season_id } }),
  playerRanking: (season_id: number, limit = 50) => get('/reports/player-ranking', { params: { season_id, limit } }),
}
export const auditApi = {
  list: (params: any) => getList('/audit-logs', { params }),
  get: (id: number) => get(`/audit-logs/${id}`),
}
