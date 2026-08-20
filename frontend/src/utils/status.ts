// 状态枚举 → 中文映射。所有业务状态统一在此维护。

export const seasonStatusMap: Record<string, string> = {
  draft: '草稿',
  registration: '报名中',
  open: '开放',
  ongoing: '进行中',
  completed: '已完成',
  archived: '已归档',
  cancelled: '已取消',
}

export const teamStatusMap: Record<string, string> = {
  pending: '待审核',
  approved: '已批准',
  rejected: '已拒绝',
  suspended: '已停赛',
}

export const matchStatusMap: Record<string, string> = {
  scheduled: '已排期',
  confirmed: '已确认',
  in_progress: '进行中',
  completed: '已完成',
  cancelled: '已取消',
  postponed: '已延期',
}

export const confirmStatusMap: Record<string, string> = {
  pending: '待确认',
  confirmed: '已确认',
  disputed: '有争议',
}

export const playerStatusMap: Record<string, string> = {
  active: '活跃',
  suspended: '停赛',
  injured: '受伤',
  transferred: '已转会',
  retired: '退役',
}

export const eligibilityMap: Record<string, string> = {
  pending: '待审核',
  approved: '已批准',
  rejected: '已拒绝',
}

export const disciplineStatusMap: Record<string, string> = {
  active: '生效中',
  served: '已执行',
  appealed: '申诉中',
  overturned: '已撤销',
}

export const punishmentMap: Record<string, string> = {
  yellow: '黄牌',
  red: '红牌',
  fine: '罚款',
  suspension: '停赛',
  warning: '警告',
}

export const severityMap: Record<string, string> = {
  minor: '轻微',
  medium: '中等',
  severe: '严重',
}

export const appealStatusMap: Record<string, string> = {
  pending: '待审核',
  reviewed: '已复议',
  accepted: '已接受',
  rejected: '已驳回',
}

export const transferStatusMap: Record<string, string> = {
  pending: '待审核',
  approved: '已批准',
  rejected: '已拒绝',
}

export const venueStatusMap: Record<string, string> = {
  available: '可用',
  maintenance: '维护中',
  unavailable: '不可用',
}

export const eventTypeMap: Record<string, string> = {
  goal: '进球',
  assist: '助攻',
  foul: '犯规',
  yellow: '黄牌',
  red: '红牌',
  substitution: '换人',
  injury: '伤病',
}

export const sportMap: Record<string, string> = {
  basketball: '篮球',
  football: '足球',
  volleyball: '排球',
}

export const formatMap: Record<string, string> = {
  round_robin: '单循环',
  double_round: '双循环',
  mixed: '混合',
  knockout: '淘汰赛',
}

export const positionMap: Record<string, string> = {
  GK: '门将',
  DEF: '后卫',
  MID: '中场',
  FWD: '前锋',
}

// label 在 map 中查找中文，找不到则原样返回。
export function label(map: Record<string, string>, v: string): string {
  return (v && map[v]) || v || '-'
}

// tagType 给状态返回 el-tag 的 type，便于区分颜色。
export type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

export function statusTagType(v: string): TagType {
  switch (v) {
    case 'ongoing':
    case 'approved':
    case 'active':
    case 'completed':
    case 'confirmed':
    case 'accepted':
      return 'success'
    case 'pending':
    case 'registration':
    case 'disputed':
    case 'appealed':
      return 'warning'
    case 'rejected':
    case 'cancelled':
    case 'suspended':
    case 'injured':
    case 'unavailable':
    case 'maintenance':
      return 'danger'
    case 'open':
      return 'primary'
    default:
      return 'info'
  }
}
