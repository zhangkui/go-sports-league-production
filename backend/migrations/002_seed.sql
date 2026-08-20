-- go-sports-league 演示数据种子
-- 仅在全新库执行一次（由 schema_migrations 跟踪）。
-- created_by=1 指向 bootstrap 创建的 admin。

-- ===== 赛季 =====
INSERT INTO `seasons`
  (`id`,`code`,`name`,`sport`,`division`,`team_count`,`format`,`rounds`,`start_date`,`end_date`,`registration_start`,`registration_end`,`status`,`current_rule_version`,`created_by`)
VALUES
  (1,'BL-2026','2026城市篮球甲级联赛','basketball','甲级',6,'double_round',10,'2026-09-01','2026-12-31','2026-08-01','2026-08-25','ongoing',1,1),
  (2,'FB-2026','2026城市足球甲级联赛','football','甲级',8,'round_robin',7,'2026-09-15','2026-12-20','2026-08-10','2026-09-05','registration',1,1);

-- ===== 积分规则（每赛季 v1）=====
INSERT INTO `scoring_rules`
  (`season_id`,`version`,`win_points`,`draw_points`,`loss_points`,`tiebreakers`,`is_active`,`created_by`)
VALUES
  (1,1,3,1,0,'points,goal_diff,goals_for,head_to_head,fair_play',1,1),
  (2,1,3,1,0,'points,goal_diff,goals_for,head_to_head,fair_play',1,1);

INSERT INTO `scoring_rule_versions` (`season_id`,`version`,`snapshot`,`created_by`) VALUES
  (1,1,'{"version":1,"win_points":3,"draw_points":1,"loss_points":0,"tiebreakers":"points,goal_diff,goals_for,head_to_head,fair_play"}',1),
  (2,1,'{"version":1,"win_points":3,"draw_points":1,"loss_points":0,"tiebreakers":"points,goal_diff,goals_for,head_to_head,fair_play"}',1);

-- ===== 球队（篮球赛季 6 支，已批准）=====
INSERT INTO `teams` (`id`,`code`,`name`,`season_id`,`contact`,`status`,`created_by`) VALUES
  (1,'LAK','湖人队',1,'张教练','approved',1),
  (2,'CEL','凯尔特人队',1,'李教练','approved',1),
  (3,'BUL','公牛队',1,'王教练','approved',1),
  (4,'WAR','勇士队',1,'赵教练','approved',1),
  (5,'SPU','马刺队',1,'钱教练','approved',1),
  (6,'ROC','火箭队',1,'孙教练','approved',1);

INSERT INTO `team_registrations` (`team_id`,`season_id`,`status`,`reviewer_id`,`reviewed_at`) VALUES
  (1,1,'approved',1,NOW()),(2,1,'approved',1,NOW()),(3,1,'approved',1,NOW()),
  (4,1,'approved',1,NOW()),(5,1,'approved',1,NOW()),(6,1,'approved',1,NOW());

-- ===== 球员（每队 6 人）=====
INSERT INTO `players` (`id`,`team_id`,`season_id`,`name`,`number`,`position`,`status`,`eligibility`) VALUES
  (1,1,1,'詹姆斯',23,'FWD','active','approved'),(2,1,1,'戴维斯',3,'DEF','active','approved'),(3,1,1,'威少',0,'MID','active','approved'),(4,1,1,'里夫斯',15,'FWD','active','approved'),(5,1,1,'施罗德',17,'MID','active','approved'),(6,1,1,'布朗',6,'DEF','active','approved'),
  (7,2,1,'塔图姆',0,'FWD','active','approved'),(8,2,1,'布朗',7,'FWD','active','approved'),(9,2,1,'斯玛特',36,'DEF','active','approved'),(10,2,1,'布罗格登',13,'MID','active','approved'),(11,2,1,'怀特',9,'MID','active','approved'),(12,2,1,'霍福德',42,'DEF','active','approved'),
  (13,3,1,'拉文',8,'FWD','active','approved'),(14,3,1,'德罗赞',11,'FWD','active','approved'),(15,3,1,'武切维奇',9,'DEF','active','approved'),(16,3,1,'鲍尔',2,'MID','injured','approved'),(17,3,1,'卡鲁索',4,'DEF','active','approved'),(18,3,1,'怀特',3,'MID','active','approved'),
  (19,4,1,'库里',30,'FWD','active','approved'),(20,4,1,'汤普森',11,'FWD','active','approved'),(21,4,1,'格林',23,'DEF','active','approved'),(22,4,1,'维金斯',22,'FWD','active','approved'),(23,4,1,'卢尼',5,'DEF','active','approved'),(24,4,1,'普尔',3,'MID','active','approved'),
  (25,5,1,'文班亚马',1,'DEF','active','approved'),(26,5,1,'瓦塞尔',24,'FWD','active','approved'),(27,5,1,'索汉',18,'MID','active','approved'),(28,5,1,'约翰逊',3,'FWD','active','approved'),(29,5,1,'凯尔登',5,'FWD','active','approved'),(30,5,1,'科林斯',22,'DEF','active','approved'),
  (31,6,1,'格林',4,'MID','active','approved'),(32,6,1,'波特',3,'FWD','active','approved'),(33,6,1,'申京',28,'DEF','active','approved'),(34,6,1,'范弗利特',5,'MID','active','approved'),(35,6,1,'史密斯',10,'FWD','active','approved'),(36,6,1,'汤普森',1,'DEF','active','approved');

-- 队长关联
UPDATE `teams` SET `captain_id`=1 WHERE id=1;
UPDATE `teams` SET `captain_id`=7 WHERE id=2;
UPDATE `teams` SET `captain_id`=13 WHERE id=3;
UPDATE `teams` SET `captain_id`=19 WHERE id=4;
UPDATE `teams` SET `captain_id`=25 WHERE id=5;
UPDATE `teams` SET `captain_id`=31 WHERE id=6;

-- ===== 场地 =====
INSERT INTO `venues` (`id`,`code`,`name`,`address`,`capacity`,`sport`,`status`) VALUES
  (1,'ARENA-A','城东体育馆','城东体育路1号',2000,'basketball','available'),
  (2,'ARENA-B','城西体育馆','城西大道88号',1500,'basketball,football','available');

INSERT INTO `venue_availability` (`venue_id`,`weekday`,`start_time`,`end_time`) VALUES
  (1,1,'18:00','22:00'),(1,3,'18:00','22:00'),(1,5,'18:00','22:00'),
  (2,2,'19:00','22:00'),(2,4,'19:00','22:00'),(2,6,'09:00','21:00'),(2,0,'09:00','21:00');

-- ===== 赛程（篮球赛季 单循环演示 5 轮 15 场）=====
-- 第 1 轮
INSERT INTO `schedules` (`id`,`season_id`,`round`,`home_team_id`,`away_team_id`,`venue_id`,`match_date`,`start_time`,`status`) VALUES
  (1,1,1,1,2,1,'2026-09-05','19:00','completed'),
  (2,1,1,3,4,2,'2026-09-05','19:30','completed'),
  (3,1,1,5,6,1,'2026-09-06','19:00','completed');
-- 第 2 轮
INSERT INTO `schedules` (`id`,`season_id`,`round`,`home_team_id`,`away_team_id`,`venue_id`,`match_date`,`start_time`,`status`) VALUES
  (4,1,2,1,3,1,'2026-09-12','19:00','completed'),
  (5,1,2,2,5,2,'2026-09-12','19:30','scheduled'),
  (6,1,2,4,6,1,'2026-09-13','19:00','scheduled');
-- 第 3 轮
INSERT INTO `schedules` (`id`,`season_id`,`round`,`home_team_id`,`away_team_id`,`venue_id`,`match_date`,`start_time`,`status`) VALUES
  (7,1,3,1,4,1,'2026-09-19','19:00','scheduled'),
  (8,1,3,2,6,2,'2026-09-19','19:30','scheduled'),
  (9,1,3,3,5,1,'2026-09-20','19:00','scheduled');
-- 第 4 轮
INSERT INTO `schedules` (`id`,`season_id`,`round`,`home_team_id`,`away_team_id`,`venue_id`,`match_date`,`start_time`,`status`) VALUES
  (10,1,4,1,5,1,'2026-09-26','19:00','scheduled'),
  (11,1,4,2,3,2,'2026-09-26','19:30','scheduled'),
  (12,1,4,4,6,1,'2026-09-27','19:00','scheduled');
-- 第 5 轮
INSERT INTO `schedules` (`id`,`season_id`,`round`,`home_team_id`,`away_team_id`,`venue_id`,`match_date`,`start_time`,`status`) VALUES
  (13,1,5,1,6,1,'2026-10-03','19:00','scheduled'),
  (14,1,5,2,4,2,'2026-10-03','19:30','scheduled'),
  (15,1,5,3,5,1,'2026-10-04','19:00','scheduled');

-- ===== 比赛记录（4 场已完成）=====
INSERT INTO `matches`
  (`id`,`schedule_id`,`season_id`,`home_team_id`,`away_team_id`,`venue_id`,`match_date`,`start_time`,`home_score`,`away_score`,`home_half_score`,`away_half_score`,`duration_min`,`status`,`confirm_status`,`confirmed_by`,`confirmed_at`)
VALUES
  (1,1,1,1,2,1,'2026-09-05','19:00',88,82,45,40,48,'completed','confirmed',1,NOW()),
  (2,2,1,3,4,2,'2026-09-05','19:30',76,76,38,38,48,'completed','confirmed',1,NOW()),
  (3,3,1,5,6,1,'2026-09-06','19:00',91,85,48,39,48,'completed','confirmed',1,NOW()),
  (4,4,1,1,3,1,'2026-09-12','19:00',95,88,50,42,48,'completed','confirmed',1,NOW());

-- 比赛事件（第 1 场）
INSERT INTO `match_events` (`match_id`,`team_id`,`player_id`,`event_type`,`minute`,`description`,`confirm_status`,`created_by`) VALUES
  (1,1,1,'goal',12,'跳投命中','confirmed',1),
  (1,1,1,'goal',28,'三分球','confirmed',1),
  (1,2,7,'goal',15,'突破上篮','confirmed',1),
  (1,1,4,'yellow',33,'防守犯规','confirmed',1),
  (1,2,8,'goal',40,'中距离','confirmed',1);

-- 上场统计（第 1 场，部分）
INSERT INTO `player_match_stats` (`match_id`,`player_id`,`team_id`,`is_starter`,`played_min`,`position`,`rating`) VALUES
  (1,1,1,1,38,'FWD',8.5),(1,2,1,1,40,'DEF',7.5),(1,7,2,1,36,'FWD',8.0),(1,8,2,1,34,'FWD',7.0);

-- ===== 积分榜（依据 4 场已完成比赛累计）=====
-- 队1: 胜队2(+3) 胜队3(+3) = 赛2 胜2 进183 失170 净+13 积6
-- 队2: 负队1 平队4... 简化按演示：队2 负队1(进82失88) 平队4? 实际只打了1场 vs 队1
-- 为保证 UI 准确，这里严格按已插入的 4 场累计：
--   M1 队1 88:82 队2  -> 队1胜3分 队2负0分
--   M2 队3 76:76 队4  -> 队3平1分 队4平1分
--   M3 队5 91:85 队6  -> 队5胜3分 队6负0分
--   M4 队1 95:88 队3  -> 队1胜3分 队3负0分
-- 合计:
--   队1: 赛2 胜2 进183 失170 净+13 积6
--   队2: 赛1 负1 进82  失88  净-6  积0
--   队3: 赛1 平1 进164 失171 净-7  积1   (76+88=164, 76+95=171)
--   队4: 赛1 平1 进76  失76  净0   积1
--   队5: 赛1 胜1 进91  失85  净+6  积3
--   队6: 赛1 负1 进85  失91  净-6  积0
INSERT INTO `standings`
  (`id`,`season_id`,`team_id`,`played`,`wins`,`draws`,`losses`,`goals_for`,`goals_against`,`goal_diff`,`points`,`fair_play`,`rank`,`prev_rank`)
VALUES
  (1,1,1,2,2,0,0,183,170,13,6,1,1,1),
  (2,1,5,1,1,0,0,91,85,6,3,0,2,2),
  (3,1,3,1,0,1,0,164,171,-7,1,0,3,3),
  (4,1,4,1,0,1,0,76,76,0,1,0,4,4),
  (5,1,2,1,0,0,1,82,88,-6,0,0,5,5),
  (6,1,6,1,0,0,1,85,91,-6,0,0,6,6);

-- 第 2 轮快照
INSERT INTO `standings_snapshots` (`season_id`,`round`,`snapshot`) VALUES
  (1,2,'[{"id":1,"season_id":1,"team_id":1,"played":2,"wins":2,"draws":0,"losses":0,"goals_for":183,"goals_against":170,"goal_diff":13,"points":6,"fair_play":1,"rank":1,"prev_rank":1,"team_name":"湖人队","team_code":"LAK"},{"id":2,"season_id":1,"team_id":5,"played":1,"wins":1,"draws":0,"losses":0,"goals_for":91,"goals_against":85,"goal_diff":6,"points":3,"fair_play":0,"rank":2,"prev_rank":2,"team_name":"马刺队","team_code":"SPU"},{"id":3,"season_id":1,"team_id":3,"played":1,"wins":0,"draws":1,"losses":0,"goals_for":164,"goals_against":171,"goal_diff":-7,"points":1,"fair_play":0,"rank":3,"prev_rank":3,"team_name":"公牛队","team_code":"BUL"},{"id":4,"season_id":1,"team_id":4,"played":1,"wins":0,"draws":1,"losses":0,"goals_for":76,"goals_against":76,"goal_diff":0,"points":1,"fair_play":0,"rank":4,"prev_rank":4,"team_name":"勇士队","team_code":"WAR"},{"id":5,"season_id":1,"team_id":2,"played":1,"wins":0,"draws":0,"losses":1,"goals_for":82,"goals_against":88,"goal_diff":-6,"points":0,"fair_play":0,"rank":5,"prev_rank":5,"team_name":"凯尔特人队","team_code":"CEL"},{"id":6,"season_id":1,"team_id":6,"played":1,"wins":0,"draws":0,"losses":1,"goals_for":85,"goals_against":91,"goal_diff":-6,"points":0,"fair_play":0,"rank":6,"prev_rank":6,"team_name":"火箭队","team_code":"ROC"}]');

-- ===== 纪律处罚（演示）=====
INSERT INTO `disciplines`
  (`id`,`player_id`,`team_id`,`season_id`,`match_id`,`punishment`,`reason`,`severity`,`suspend_games`,`fine_amount`,`status`,`issued_by`)
VALUES
  (1,4,1,1,1,'yellow','防守动作过大','minor',0,0,'served',1),
  (2,16,3,1,NULL,'suspension','赛季中累计技术犯规','medium',2,500,'active',1);

INSERT INTO `suspensions` (`discipline_id`,`player_id`,`total_games`,`served_games`,`start_date`,`status`) VALUES
  (2,16,2,0,'2026-09-10','active');

INSERT INTO `appeals` (`discipline_id`,`appellant_id`,`reason`,`status`) VALUES
  (2,3,'球员认为判罚过重，请求复议','pending');

-- 同步自增指针，避免后续 INSERT 主键冲突
ALTER TABLE `seasons` AUTO_INCREMENT = 3;
ALTER TABLE `teams` AUTO_INCREMENT = 7;
ALTER TABLE `players` AUTO_INCREMENT = 37;
ALTER TABLE `venues` AUTO_INCREMENT = 3;
ALTER TABLE `schedules` AUTO_INCREMENT = 16;
ALTER TABLE `matches` AUTO_INCREMENT = 5;
ALTER TABLE `standings` AUTO_INCREMENT = 7;
ALTER TABLE `disciplines` AUTO_INCREMENT = 3;
