# go-sports-league 业余体育联赛与赛程积分系统

面向城市业余联赛组织者的**联赛全流程管理平台**：球队注册 → 智能排赛 → 比赛记录 → 积分排名，全链路数字化、可审计。

- **后端**：Go 1.22+ REST API（gorilla/mux + MySQL 8 + Redis 7 + JWT）
- **前端**：Vue 3 + TypeScript + Vite + Element Plus + Pinia
- **部署**：Docker Compose 全容器化，一键启动

## 功能模块

| 模块 | 能力 |
|------|------|
| 认证授权 | 注册/登录/刷新/退出、bcrypt、JWT(15m/7d)、登录限流、RBAC |
| 用户权限 | 用户/角色/权限 CRUD，7 个内置角色 |
| 联赛赛季 | 赛季生命周期(草稿→报名→开放→进行→完成→归档)、积分规则版本化 |
| 球队球员 | 球队注册审核、球员管理、转会审批 |
| 场地赛程 | 场地可用时段、单/双循环自动排赛、冲突检测(球队/场地/轮次/时间窗口) |
| 比赛记录 | 比分录入、事件(进球/黄红牌/换人等)、上场统计、确认即事务更新积分 |
| 纪律处罚 | 处罚记录、停赛执行、申诉复议 |
| 积分排名 | 实时积分榜、每轮快照、可配置同分规则 |
| 报表审计 | 赛季汇总、球员排行、全链路审计日志 |

## 目录结构

```
goxm2/
├── docker-compose.yml          # 一键启动 MySQL + Redis + 后端 + 前端
├── .env.example                # 环境变量样例
├── backend/                    # Go REST API
│   ├── cmd/server/main.go      # 入口(优雅关闭)
│   ├── internal/
│   │   ├── config/  models/  database/
│   │   ├── repository/  service/  middleware/  handler/  router/  auth/
│   │   └── pkg/(response,errorsx,hash,validator,pagination,logx)
│   ├── migrations/001_init.sql # 20+ 表 + 角色/权限种子
│   └── scripts/count.sh        # 生产代码行数/文件数门禁
└── frontend/                   # Vue 3 管理前端
    └── src/{api,stores,router,layouts,views,composables}
```

## 快速开始

### 一键启动（Docker Compose）

> 前置要求：Docker 20.10+ 与 Docker Compose v2（`docker compose` 子命令）。

**1. 准备配置文件**

```bash
cp .env.example .env      # 按需修改密码 / JWT 密钥等
```

**2. 构建并启动全部服务**

```bash
docker compose up -d --build
```

首次启动会：拉取 MySQL 8.4 / Redis 7.4 镜像 → 构建 Go 后端与 Vue 前端镜像 → 启动容器。
后端启动后会自动执行数据库迁移（`001_init.sql`，建表 + 角色/权限种子）并创建默认管理员 `admin`。

**3. 检查服务状态**

```bash
docker compose ps                 # 四个容器均应为 Up / healthy
```

逐项健康验证：

```bash
# 后端健康检查
curl -s http://localhost:8080/health          # {"code":0,...,"data":{"status":"ok"}}
curl -s http://localhost:8080/ready           # {"code":0,...,"data":{"status":"ready"}}

# 前端
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:5173   # 200

# MySQL（容器内）
docker compose exec mysql mysqladmin -uroot -proot_pass ping     # mysqld is alive

# Redis（容器内）
docker compose exec redis redis-cli ping                        # PONG
```

**4. 验证默认管理员登录**

```bash
# 登录换取 token
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin123!"}'

# 用返回的 access_token 调用 /me
curl -s http://localhost:8080/api/v1/me -H "Authorization: Bearer <access_token>"
```

浏览器访问前端 **http://localhost:5173**，用 `admin / Admin123!` 登录即可。

### 服务清单

| 容器 | 端口 | 说明 |
|------|------|------|
| gsl-mysql | 3307→3306 | MySQL 8.4 |
| gsl-redis | 6380→6379 | Redis 7.4.10 |
| gsl-backend | 8080 | Go API |
| gsl-frontend | 5173→80 | Vue 3 (nginx 静态 + /api 反代) |

启动顺序：`mysql`(healthy) → `redis`(healthy) → `backend` → `frontend`，由 `depends_on` + `healthcheck` 保证。

### 常用运维命令

```bash
docker compose logs -f backend            # 实时查看后端日志
docker compose logs -f                     # 全部服务日志
docker compose restart backend             # 重启某服务
docker compose up -d --build backend       # 仅改动后端代码后重新构建并启动
docker compose down                        # 停止并删除容器（保留数据卷）
docker compose down -v                     # 停止并删除数据卷（彻底重置数据库）
docker compose exec mysql mysql -uroot -proot_pass league_db   # 进入 MySQL
docker compose exec backend sh             # 进入后端容器 shell
```

### 故障排查

| 现象 | 排查 |
|------|------|
| backend 不断重启 | `docker compose logs backend`，多为连不上 MySQL/Redis——确认 mysql/redis 已 `healthy` |
| 前端 502 / `/api` 报错 | 后端未就绪，等 backend 健康检查通过；或确认 nginx 反代到 `backend:8080` |
| `unknown variable 'default-authentication-plugin'` | 不应再出现：8.4 已改用 `MYSQL_AUTHENTICATION_POLICY`，请确认 docker-compose.yml 为最新 |
| 端口被占用 | 修改 `.env` 或 `docker-compose.yml` 中左侧端口（3307/6380/8080/5173） |
| 想完全重来 | `docker compose down -v` 后再 `docker compose up -d --build` |

### 本地开发

后端：
```bash
cd backend
go mod download
# 启动本地 MySQL/Redis 并设置 APP_MYSQL_HOST=localhost 等
go run ./cmd/server
```

前端：
```bash
cd frontend
npm install
npm run dev   # http://localhost:5173 ，开发代理 /api -> http://localhost:8080
```

## API 概览

统一响应：`{ code, message, data, timestamp, request_id }`（`code=0` 表示成功）
认证：`Authorization: Bearer <access_token>`
幂等：`POST/PUT` 请求可携带 `Idempotency-Key`
分页：`page` / `page_size` / `order_by` / `order`

主要资源（均挂载在 `/api/v1` 下）：

- 认证：`auth/register|login|refresh|logout`、`me`、`me/password`
- 用户/角色/权限：`users`、`roles`、`permissions`
- 赛季/规则：`seasons`、`seasons/{id}/rules`、`/rules/versions`、`/status`
- 球队/球员/转会：`teams`、`team-registrations`、`players`、`transfers`
- 场地/赛程：`venues`、`venues/{id}/availability`、`schedules`、`/generate`、`/conflicts`
- 比赛/事件/统计：`matches`、`/{id}/events`、`/{id}/stats`、`/confirm`、`/status`
- 纪律/申诉：`disciplines`、`/overturn`、`appeals`、`/review`、`players/{id}/disciplines`
- 积分/报表：`standings`、`standings/snapshots`、`reports/season-summary|player-ranking`
- 审计：`audit-logs`
- 健康检查：`/health`、`/ready`

## 核心设计要点

- **积分更新事务边界（INV-05）**：比赛确认在单事务内完成 `更新主/客队积分 → 更新比赛确认状态 → 重新排名 → 生成快照 → 写审计日志`。
- **赛程冲突检测**：球队同日多赛、场地同时段多赛、同轮多赛、场地可用时段窗口。
- **积分规则版本化（INV-06）**：每次修改生成新版本，旧版本入 `scoring_rule_versions` 快照。
- **幂等性（INV-07）**：关键写操作支持 `Idempotency-Key`（Redis 缓存 + 抢占锁）。
- **RBAC**：所有受保护端点经中间件 `RequireAuth` + `RequirePermission`，权限码 `resource:action`。
- **限流**：登录 `5次/分钟/IP`、`10次/分钟/用户`；Redis 滑动窗口。

## 代码门禁

后端生产代码门禁（排除测试/vendor/生成代码/migrations）：

```bash
cd backend && bash scripts/count.sh
# 生产Go文件数: 67   (≥ 51 ✅)
# 生产Go代码行数: 7476 (≥ 5001 ✅)
```

## 默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | Admin123! | 系统管理员 |

## 技术栈

- Go 1.22+ / gorilla/mux / go-sql-driver/mysql / redis/go-redis/v9
- golang-jwt/v5 / golang.org/x/crypto(bcrypt) / spf13/viper / uber-go/zap
- Vue 3 / TypeScript / Vite / Pinia / Element Plus / ECharts / Axios
