-- go-sports-league initial schema
-- MySQL 8.4  utf8mb4
-- Run order: top to bottom. FK checks temporarily disabled for clean re-creation.

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================================
-- 1. Users & Auth
-- ============================================================================
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username`     VARCHAR(64)  NOT NULL,
  `email`        VARCHAR(128) NOT NULL,
  `password_hash` VARCHAR(128) NOT NULL,
  `full_name`    VARCHAR(64)  NULL,
  `phone`        VARCHAR(32)  NULL,
  `status`       TINYINT     NOT NULL DEFAULT 1 COMMENT '1=active 0=disabled',
  `created_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `last_login_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_username` (`username`),
  UNIQUE KEY `uk_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `refresh_tokens`;
CREATE TABLE `refresh_tokens` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `token_id`    VARCHAR(64)  NOT NULL COMMENT 'opaque jti identifier',
  `user_id`     BIGINT UNSIGNED NOT NULL,
  `token_hash`  VARCHAR(128) NOT NULL,
  `issued_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `expires_at`  DATETIME(3) NOT NULL,
  `revoked_at`  DATETIME(3) NULL,
  `user_agent`  VARCHAR(255) NULL,
  `ip`          VARCHAR(64)  NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_rt_token_id` (`token_id`),
  KEY `idx_rt_user` (`user_id`),
  KEY `idx_rt_expires` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 2. Roles & Permissions (RBAC)
-- ============================================================================
DROP TABLE IF EXISTS `roles`;
CREATE TABLE `roles` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`       VARCHAR(32) NOT NULL,
  `name`       VARCHAR(64) NOT NULL,
  `description`VARCHAR(255) NULL,
  `is_builtin` TINYINT    NOT NULL DEFAULT 0,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_roles_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `permissions`;
CREATE TABLE `permissions` (
  `id`      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `resource` VARCHAR(64) NOT NULL,
  `action`  VARCHAR(32) NOT NULL,
  `name`    VARCHAR(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_perm_res_act` (`resource`,`action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `user_roles`;
CREATE TABLE `user_roles` (
  `user_id` BIGINT UNSIGNED NOT NULL,
  `role_id` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`user_id`,`role_id`),
  KEY `idx_ur_role` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `role_permissions`;
CREATE TABLE `role_permissions` (
  `role_id`      BIGINT UNSIGNED NOT NULL,
  `permission_id` BIGINT UNSIGNED NOT NULL,
  `created_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`role_id`,`permission_id`),
  KEY `idx_rp_perm` (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 3. Seasons & Scoring rules
-- ============================================================================
DROP TABLE IF EXISTS `seasons`;
CREATE TABLE `seasons` (
  `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`           VARCHAR(32) NOT NULL,
  `name`           VARCHAR(128) NOT NULL,
  `sport`          VARCHAR(16) NOT NULL COMMENT 'basketball/football/volleyball/...',
  `division`       VARCHAR(32) NULL COMMENT '甲级/乙级',
  `team_count`     INT NOT NULL DEFAULT 0,
  `format`         VARCHAR(16) NOT NULL DEFAULT 'round_robin' COMMENT 'round_robin/double_round/mixed/knockout',
  `rounds`         INT NOT NULL DEFAULT 0,
  `start_date`     DATE NULL,
  `end_date`       DATE NULL,
  `registration_start` DATE NULL,
  `registration_end`   DATE NULL,
  `status`         VARCHAR(16) NOT NULL DEFAULT 'draft',
  `current_rule_version` INT NOT NULL DEFAULT 1,
  `created_by`     BIGINT UNSIGNED NOT NULL,
  `created_at`     DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`     DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_seasons_code` (`code`),
  KEY `idx_seasons_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `scoring_rules`;
CREATE TABLE `scoring_rules` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `season_id`   BIGINT UNSIGNED NOT NULL,
  `version`     INT NOT NULL DEFAULT 1,
  `win_points`  INT NOT NULL DEFAULT 3,
  `draw_points` INT NOT NULL DEFAULT 1,
  `loss_points` INT NOT NULL DEFAULT 0,
  `tiebreakers` VARCHAR(255) NOT NULL DEFAULT 'goal_diff,goals_for,head_to_head,fair_play',
  `is_active`   TINYINT NOT NULL DEFAULT 1,
  `created_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `created_by`   BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sr_season_version` (`season_id`,`version`),
  KEY `idx_sr_season` (`season_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `scoring_rule_versions`;
CREATE TABLE `scoring_rule_versions` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `season_id`   BIGINT UNSIGNED NOT NULL,
  `version`     INT NOT NULL,
  `snapshot`    JSON NOT NULL,
  `created_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `created_by`  BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_srv_season` (`season_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 4. Teams & Players
-- ============================================================================
DROP TABLE IF EXISTS `teams`;
CREATE TABLE `teams` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`       VARCHAR(32) NOT NULL,
  `name`       VARCHAR(128) NOT NULL,
  `season_id`  BIGINT UNSIGNED NOT NULL,
  `logo_url`   VARCHAR(255) NULL,
  `captain_id` BIGINT UNSIGNED NULL,
  `contact`    VARCHAR(64) NULL,
  `status`     VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT 'pending/approved/rejected/suspended',
  `created_by` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_teams_season_name` (`season_id`,`name`),
  UNIQUE KEY `uk_teams_code` (`code`),
  KEY `idx_teams_season` (`season_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `team_registrations`;
CREATE TABLE `team_registrations` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `team_id`     BIGINT UNSIGNED NOT NULL,
  `season_id`   BIGINT UNSIGNED NOT NULL,
  `status`      VARCHAR(16) NOT NULL DEFAULT 'pending',
  `reviewer_id` BIGINT UNSIGNED NULL,
  `review_note` VARCHAR(500) NULL,
  `reviewed_at` DATETIME(3) NULL,
  `created_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_tr_season` (`season_id`),
  KEY `idx_tr_team` (`team_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `players`;
CREATE TABLE `players` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `team_id`      BIGINT UNSIGNED NOT NULL,
  `season_id`    BIGINT UNSIGNED NOT NULL,
  `name`         VARCHAR(64) NOT NULL,
  `number`       INT NOT NULL,
  `position`     VARCHAR(16) NULL COMMENT 'GK/DEF/MID/FWD',
  `birth_date`   DATE NULL,
  `height_cm`    INT NULL,
  `weight_kg`    INT NULL,
  `status`       VARCHAR(16) NOT NULL DEFAULT 'active',
  `eligibility`  VARCHAR(16) NOT NULL DEFAULT 'pending',
  `created_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_players_team_number` (`team_id`,`number`),
  UNIQUE KEY `uk_players_season_player` (`season_id`,`id`) COMMENT 'player one team per season enforced in app',
  KEY `idx_players_team` (`team_id`),
  KEY `idx_players_season` (`season_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `transfers`;
CREATE TABLE `transfers` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `player_id`    BIGINT UNSIGNED NOT NULL,
  `from_team_id` BIGINT UNSIGNED NOT NULL,
  `to_team_id`   BIGINT UNSIGNED NOT NULL,
  `season_id`    BIGINT UNSIGNED NOT NULL,
  `reason`       VARCHAR(255) NULL,
  `status`       VARCHAR(16) NOT NULL DEFAULT 'pending',
  `requested_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `effective_at` DATE NULL,
  `reviewer_id`  BIGINT UNSIGNED NULL,
  `reviewed_at`  DATETIME(3) NULL,
  `review_note`  VARCHAR(500) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_transfers_player` (`player_id`),
  KEY `idx_transfers_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 5. Venues & Schedules
-- ============================================================================
DROP TABLE IF EXISTS `venues`;
CREATE TABLE `venues` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code`        VARCHAR(32) NOT NULL,
  `name`        VARCHAR(128) NOT NULL,
  `address`     VARCHAR(255) NULL,
  `capacity`    INT NULL,
  `sport`       VARCHAR(64) NULL,
  `status`      VARCHAR(16) NOT NULL DEFAULT 'available',
  `created_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_venues_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `venue_availability`;
CREATE TABLE `venue_availability` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `venue_id`    BIGINT UNSIGNED NOT NULL,
  `weekday`     TINYINT NOT NULL COMMENT '0=Sun..6=Sat',
  `start_time`  TIME NOT NULL,
  `end_time`    TIME NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_va_venue_wd_time` (`venue_id`,`weekday`,`start_time`),
  KEY `idx_va_venue` (`venue_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `schedules`;
CREATE TABLE `schedules` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `season_id`    BIGINT UNSIGNED NOT NULL,
  `round`        INT NOT NULL,
  `home_team_id` BIGINT UNSIGNED NOT NULL,
  `away_team_id` BIGINT UNSIGNED NOT NULL,
  `venue_id`     BIGINT UNSIGNED NOT NULL,
  `match_date`   DATE NOT NULL,
  `start_time`   TIME NOT NULL,
  `status`       VARCHAR(16) NOT NULL DEFAULT 'scheduled',
  `created_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sch_venue_dt` (`venue_id`,`match_date`,`start_time`),
  KEY `idx_sch_home_date` (`home_team_id`,`match_date`),
  KEY `idx_sch_away_date` (`away_team_id`,`match_date`),
  KEY `idx_sch_season_round` (`season_id`,`round`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `schedule_conflicts`;
CREATE TABLE `schedule_conflicts` (
  `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `schedule_id`    BIGINT UNSIGNED NOT NULL,
  `conflict_type`  VARCHAR(16) NOT NULL COMMENT 'team/venue/round/window',
  `description`    VARCHAR(255) NULL,
  `detected_at`    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_sc_schedule` (`schedule_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 6. Matches & Events
-- ============================================================================
DROP TABLE IF EXISTS `matches`;
CREATE TABLE `matches` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `schedule_id`   BIGINT UNSIGNED NOT NULL,
  `season_id`     BIGINT UNSIGNED NOT NULL,
  `home_team_id`  BIGINT UNSIGNED NOT NULL,
  `away_team_id`  BIGINT UNSIGNED NOT NULL,
  `venue_id`      BIGINT UNSIGNED NOT NULL,
  `match_date`    DATE NOT NULL,
  `start_time`   TIME NOT NULL,
  `home_score`    INT NULL,
  `away_score`    INT NULL,
  `home_half_score` INT NULL,
  `away_half_score` INT NULL,
  `referee_id`    BIGINT UNSIGNED NULL,
  `recorder_id`   BIGINT UNSIGNED NULL,
  `duration_min`  INT NULL,
  `status`        VARCHAR(16) NOT NULL DEFAULT 'scheduled',
  `confirm_status` VARCHAR(16) NOT NULL DEFAULT 'pending',
  `confirmed_by`  BIGINT UNSIGNED NULL,
  `confirmed_at`  DATETIME(3) NULL,
  `created_at`     DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_match_schedule` (`schedule_id`),
  KEY `idx_match_season` (`season_id`),
  KEY `idx_match_status` (`status`),
  KEY `idx_match_home` (`home_team_id`),
  KEY `idx_match_away` (`away_team_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `match_events`;
CREATE TABLE `match_events` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `match_id`     BIGINT UNSIGNED NOT NULL,
  `team_id`      BIGINT UNSIGNED NOT NULL,
  `player_id`    BIGINT UNSIGNED NULL,
  `event_type`   VARCHAR(16) NOT NULL COMMENT 'goal/assist/foul/yellow/red/substitution/injury',
  `minute`       INT NOT NULL,
  `description`  VARCHAR(255) NULL,
  `confirm_status` VARCHAR(16) NOT NULL DEFAULT 'pending',
  `created_by`   BIGINT UNSIGNED NOT NULL,
  `created_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_me_match` (`match_id`),
  KEY `idx_me_player` (`player_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `player_match_stats`;
CREATE TABLE `player_match_stats` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `match_id`     BIGINT UNSIGNED NOT NULL,
  `player_id`    BIGINT UNSIGNED NOT NULL,
  `team_id`      BIGINT UNSIGNED NOT NULL,
  `is_starter`   TINYINT NOT NULL DEFAULT 0,
  `played_min`   INT NULL,
  `position`     VARCHAR(16) NULL,
  `rating`       DECIMAL(4,2) NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pms_match_player` (`match_id`,`player_id`),
  KEY `idx_pms_player` (`player_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 7. Discipline & Appeals
-- ============================================================================
DROP TABLE IF EXISTS `disciplines`;
CREATE TABLE `disciplines` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `player_id`     BIGINT UNSIGNED NOT NULL,
  `team_id`       BIGINT UNSIGNED NOT NULL,
  `season_id`     BIGINT UNSIGNED NOT NULL,
  `match_id`      BIGINT UNSIGNED NULL,
  `punishment`    VARCHAR(16) NOT NULL COMMENT 'yellow/red/fine/suspension/warning',
  `reason`        VARCHAR(255) NOT NULL,
  `severity`      VARCHAR(16) NOT NULL DEFAULT 'medium' COMMENT 'minor/medium/severe',
  `suspend_games` INT NOT NULL DEFAULT 0,
  `fine_amount`   DECIMAL(10,2) NOT NULL DEFAULT 0,
  `status`        VARCHAR(16) NOT NULL DEFAULT 'active',
  `issued_by`     BIGINT UNSIGNED NOT NULL,
  `created_at`    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at`    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_disc_player` (`player_id`),
  KEY `idx_disc_status` (`status`),
  KEY `idx_disc_season` (`season_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `suspensions`;
CREATE TABLE `suspensions` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `discipline_id`   BIGINT UNSIGNED NOT NULL,
  `player_id`       BIGINT UNSIGNED NOT NULL,
  `total_games`     INT NOT NULL,
  `served_games`    INT NOT NULL DEFAULT 0,
  `start_date`      DATE NOT NULL,
  `end_date`        DATE NULL,
  `status`          VARCHAR(16) NOT NULL DEFAULT 'active',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_susp_discipline` (`discipline_id`),
  KEY `idx_susp_player` (`player_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `appeals`;
CREATE TABLE `appeals` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `discipline_id` BIGINT UNSIGNED NOT NULL,
  `appellant_id`  BIGINT UNSIGNED NOT NULL,
  `reason`        VARCHAR(500) NOT NULL,
  `status`        VARCHAR(16) NOT NULL DEFAULT 'pending',
  `reviewer_id`   BIGINT UNSIGNED NULL,
  `review_opinion` VARCHAR(500) NULL,
  `reviewed_at`   DATETIME(3) NULL,
  `created_at`    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_appeal_discipline` (`discipline_id`),
  KEY `idx_appeal_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 8. Standings & Snapshots
-- ============================================================================
DROP TABLE IF EXISTS `standings`;
CREATE TABLE `standings` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `season_id`    BIGINT UNSIGNED NOT NULL,
  `team_id`      BIGINT UNSIGNED NOT NULL,
  `played`       INT NOT NULL DEFAULT 0,
  `wins`         INT NOT NULL DEFAULT 0,
  `draws`        INT NOT NULL DEFAULT 0,
  `losses`       INT NOT NULL DEFAULT 0,
  `goals_for`    INT NOT NULL DEFAULT 0,
  `goals_against` INT NOT NULL DEFAULT 0,
  `goal_diff`    INT NOT NULL DEFAULT 0,
  `points`       INT NOT NULL DEFAULT 0,
  `fair_play`    INT NOT NULL DEFAULT 0,
  `rank`         INT NOT NULL DEFAULT 0,
  `prev_rank`    INT NOT NULL DEFAULT 0,
  `updated_at`   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_st_season_team` (`season_id`,`team_id`),
  KEY `idx_st_season` (`season_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `standings_snapshots`;
CREATE TABLE `standings_snapshots` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `season_id`   BIGINT UNSIGNED NOT NULL,
  `round`       INT NOT NULL,
  `snapshot`    JSON NOT NULL,
  `created_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_ss_season_round` (`season_id`,`round`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- 9. Audit logs
-- ============================================================================
DROP TABLE IF EXISTS `audit_logs`;
CREATE TABLE `audit_logs` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT UNSIGNED NULL,
  `username`   VARCHAR(64) NULL,
  `action`     VARCHAR(64) NOT NULL,
  `resource`   VARCHAR(64) NOT NULL,
  `resource_id` VARCHAR(64) NULL,
  `method`     VARCHAR(8) NULL,
  `path`       VARCHAR(255) NULL,
  `status_code` INT NULL,
  `ip`         VARCHAR(64) NULL,
  `request_id` VARCHAR(64) NULL,
  `detail`     TEXT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_audit_user_time` (`user_id`,`created_at`),
  KEY `idx_audit_resource` (`resource`),
  KEY `idx_audit_time` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================================
-- Seed: built-in roles
-- ============================================================================
INSERT INTO `roles` (`code`,`name`,`description`,`is_builtin`) VALUES
('admin','系统管理员','系统配置与用户权限管理',1),
('league_admin','联赛管理员','赛季与赛程管理',1),
('team_admin','球队管理员','球队与球员管理',1),
('referee','裁判/记录员','比赛数据录入',1),
('discipline','纪律委员','违规处罚与申诉',1),
('auditor','审计人员','只读审计',1),
('spectator','普通观众','查看赛程积分',1);

-- ============================================================================
-- Seed: permissions (resource, action)
-- ============================================================================
INSERT INTO `permissions` (`resource`,`action`,`name`) VALUES
('users','manage','用户与角色管理'),
('seasons','manage','赛季管理'),
('seasons','view','查看赛季'),
('teams','manage','球队管理'),
('teams','view','查看球队'),
('teams','approve','球队审核'),
('players','manage','球员管理'),
('players','view','查看球员'),
('venues','manage','场地管理'),
('venues','view','查看场地'),
('schedules','manage','赛程管理'),
('schedules','generate','生成赛程'),
('schedules','view','查看赛程'),
('matches','manage','比赛管理'),
('matches','record','比赛录入'),
('matches','confirm','比赛确认'),
('matches','view','查看比赛'),
('disciplines','manage','纪律处罚'),
('disciplines','view','查看处罚'),
('appeals','manage','申诉管理'),
('standings','view','查看积分榜'),
('reports','view','查看报表'),
('audit_logs','view','查看审计日志');

-- role -> permission mapping (use known ids starting at 1)
-- admin: all (1..23), league_admin: seasons/teams-approve/players-view/venues/schedules/matches-discipline/standings/reports
INSERT INTO `role_permissions` (`role_id`,`permission_id`)
SELECT r.id, p.id FROM `roles` r JOIN `permissions` p
WHERE r.code='admin';

INSERT INTO `role_permissions` (`role_id`,`permission_id`)
SELECT r.id, p.id FROM `roles` r JOIN `permissions` p
WHERE r.code='league_admin' AND p.action IN ('manage','view','approve','generate','record','confirm') ;

INSERT INTO `role_permissions` (`role_id`,`permission_id`)
SELECT r.id, p.id FROM `roles` r JOIN `permissions` p
WHERE r.code='team_admin' AND p.resource IN ('teams','players') AND p.action IN ('manage','view');

INSERT INTO `role_permissions` (`role_id`,`permission_id`)
SELECT r.id, p.id FROM `roles` r JOIN `permissions` p
WHERE r.code='referee' AND p.resource IN ('matches') AND p.action IN ('record','view');

INSERT INTO `role_permissions` (`role_id`,`permission_id`)
SELECT r.id, p.id FROM `roles` r JOIN `permissions` p
WHERE r.code='discipline' AND p.resource IN ('disciplines','appeals') AND p.action IN ('manage','view');

INSERT INTO `role_permissions` (`role_id`,`permission_id`)
SELECT r.id, p.id FROM `roles` r JOIN `permissions` p
WHERE r.code='auditor' AND p.action='view';

INSERT INTO `role_permissions` (`role_id`,`permission_id`)
SELECT r.id, p.id FROM `roles` r JOIN `permissions` p
WHERE r.code='spectator' AND p.resource IN ('seasons','teams','schedules','matches','standings') AND p.action='view';
