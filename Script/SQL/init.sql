-- ==========================================
-- 千年江湖 数据库初始化脚本
-- 版本: V1.1
-- 日期: 2026-06-15
-- 描述: 千年武侠MMORPG核心数据库 - 完整武学系统
-- ==========================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS qiannian DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE qiannian;

-- ==========================================
-- 1. 账号表 (account)
-- ==========================================
DROP TABLE IF EXISTS `account`;
CREATE TABLE `account` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '账号ID',
    `username` VARCHAR(50) NOT NULL COMMENT '用户名',
    `password` VARCHAR(64) NOT NULL COMMENT '密码(MD5)',
    `salt` VARCHAR(32) NOT NULL COMMENT '盐值',
    `email` VARCHAR(100) DEFAULT NULL COMMENT '邮箱',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `last_login_time` DATETIME DEFAULT NULL COMMENT '最后登录时间',
    `last_login_ip` VARCHAR(50) DEFAULT NULL COMMENT '最后登录IP',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 0=封号, 1=正常',
    `ban_reason` VARCHAR(255) DEFAULT NULL COMMENT '封号原因',
    `ban_time` DATETIME DEFAULT NULL COMMENT '封号时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    KEY `idx_status` (`status`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='账号表';

-- ==========================================
-- 2. 角色表 (role)
-- ==========================================
DROP TABLE IF EXISTS `role`;
CREATE TABLE `role` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色ID',
    `account_id` BIGINT UNSIGNED NOT NULL COMMENT '账号ID',
    `name` VARCHAR(20) NOT NULL COMMENT '角色名',
    `level` INT NOT NULL DEFAULT 1 COMMENT '等级',
    `exp` BIGINT NOT NULL DEFAULT 0 COMMENT '经验值',
    `gold` BIGINT NOT NULL DEFAULT 0 COMMENT '金币',
    `bind_gold` BIGINT NOT NULL DEFAULT 0 COMMENT '绑定金币',
    `yuanbao` INT NOT NULL DEFAULT 0 COMMENT '元宝',
    `gender` TINYINT NOT NULL DEFAULT 0 COMMENT '性别: 0=男, 1=女',
    `appearance` INT NOT NULL DEFAULT 0 COMMENT '形象ID',
    `title` VARCHAR(50) DEFAULT NULL COMMENT '称号',
    
    -- 基础属性
    `hp` INT NOT NULL DEFAULT 100 COMMENT '生命值',
    `max_hp` INT NOT NULL DEFAULT 100 COMMENT '最大生命值',
    `mp` INT NOT NULL DEFAULT 100 COMMENT '内力值',
    `max_mp` INT NOT NULL DEFAULT 100 COMMENT '最大内力值',
    `stamina` INT NOT NULL DEFAULT 100 COMMENT '体力值',
    `max_stamina` INT NOT NULL DEFAULT 100 COMMENT '最大体力值',
    
    -- 战斗属性
    `attack` INT NOT NULL DEFAULT 10 COMMENT '攻击',
    `defense` INT NOT NULL DEFAULT 5 COMMENT '防御',
    `speed` INT NOT NULL DEFAULT 10 COMMENT '速度',
    `hit` INT NOT NULL DEFAULT 50 COMMENT '命中',
    `dodge` INT NOT NULL DEFAULT 10 COMMENT '闪避',
    `crit` INT NOT NULL DEFAULT 5 COMMENT '暴击率',
    `crit_damage` INT NOT NULL DEFAULT 150 COMMENT '暴击伤害',
    
    -- 位置信息
    `map_id` INT NOT NULL DEFAULT 1 COMMENT '当前地图ID',
    `map_x` INT NOT NULL DEFAULT 100 COMMENT '地图X坐标',
    `map_y` INT NOT NULL DEFAULT 100 COMMENT '地图Y坐标',
    
    -- PK相关
    `pk_mode` TINYINT NOT NULL DEFAULT 0 COMMENT 'PK模式: 0=和平, 1=全体, 2=门派, 3=组队',
    `pk_value` INT NOT NULL DEFAULT 0 COMMENT 'PK值(善恶值,越低越红)',
    `kill_count` INT NOT NULL DEFAULT 0 COMMENT '杀人数量',
    `death_count` INT NOT NULL DEFAULT 0 COMMENT '死亡次数',
    
    -- 状态
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0=正常, 1=打坐, 2=死亡, 3=在线',
    `hp_regen` INT NOT NULL DEFAULT 1 COMMENT '生命回复',
    `mp_regen` INT NOT NULL DEFAULT 1 COMMENT '内力回复',
    
    -- 时间戳
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `last_save_time` DATETIME DEFAULT NULL COMMENT '最后保存时间',
    `online_time` BIGINT NOT NULL DEFAULT 0 COMMENT '累计在线时间(秒)',
    `logout_time` DATETIME DEFAULT NULL COMMENT '最后下线时间',
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`),
    KEY `idx_account_id` (`account_id`),
    KEY `idx_map_id` (`map_id`),
    KEY `idx_level` (`level`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- ==========================================
-- 3. 武学基础表 (skill_base) - 千年完整武学系统
-- ==========================================
DROP TABLE IF EXISTS `skill_base`;
CREATE TABLE `skill_base` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '武学ID',
    `name` VARCHAR(50) NOT NULL COMMENT '武学名称',
    `type` TINYINT NOT NULL COMMENT '类型: 1=内功, 2=外功, 3=身法, 4=护体, 5=拳法, 6=剑法, 7=刀法, 8=枪法, 9=斧法',
    `sub_type` TINYINT DEFAULT NULL COMMENT '子类型: 1=无名武功(初级), 2=进阶武功, 3=高级武功',
    `level` INT NOT NULL DEFAULT 1 COMMENT '武学等级要求',
    `max_level` INT NOT NULL DEFAULT 100 COMMENT '武学最高等级',
    `exp_factor` INT NOT NULL DEFAULT 100 COMMENT '经验系数(每级所需经验)',
    `description` VARCHAR(255) DEFAULT NULL COMMENT '武学描述',
    
    -- 属性加成
    `hp_bonus` INT NOT NULL DEFAULT 0 COMMENT '生命加成/级',
    `mp_bonus` INT NOT NULL DEFAULT 0 COMMENT '内力加成/级',
    `attack_bonus` INT NOT NULL DEFAULT 0 COMMENT '攻击加成/级',
    `defense_bonus` INT NOT NULL DEFAULT 0 COMMENT '防御加成/级',
    `speed_bonus` INT NOT NULL DEFAULT 0 COMMENT '速度加成/级',
    `hit_bonus` INT NOT NULL DEFAULT 0 COMMENT '命中加成/级',
    `dodge_bonus` INT NOT NULL DEFAULT 0 COMMENT '闪避加成/级',
    `crit_bonus` INT NOT NULL DEFAULT 0 COMMENT '暴击加成/级',
    
    -- 特效
    `buff_id` INT DEFAULT NULL COMMENT '被动BUFF ID',
    `skill_effect` VARCHAR(100) DEFAULT NULL COMMENT '技能特效',
    
    `is_active` TINYINT NOT NULL DEFAULT 1 COMMENT '是否激活',
    PRIMARY KEY (`id`),
    KEY `idx_type` (`type`),
    KEY `idx_level` (`level`),
    KEY `idx_sub_type` (`sub_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='武学基础表';

-- ==========================================
-- 4. 角色武学表 (role_skill) - 角色已学习的武学
-- ==========================================
DROP TABLE IF EXISTS `role_skill`;
CREATE TABLE `role_skill` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    `skill_id` INT UNSIGNED NOT NULL COMMENT '武学ID',
    `level` INT NOT NULL DEFAULT 1 COMMENT '当前等级',
    `exp` BIGINT NOT NULL DEFAULT 0 COMMENT '当前熟练度',
    `is_equip` TINYINT NOT NULL DEFAULT 0 COMMENT '是否装备: 0=未装备, 1=已装备',
    `learn_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '学习时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_skill` (`role_id`, `skill_id`),
    KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色武学表';

-- ==========================================
-- 5. 道具基础表 (item_base)
-- ==========================================
DROP TABLE IF EXISTS `item_base`;
CREATE TABLE `item_base` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '道具ID',
    `name` VARCHAR(50) NOT NULL COMMENT '道具名称',
    `type` TINYINT NOT NULL COMMENT '类型: 1=药品, 2=装备, 3=材料, 4=任务, 5=秘籍, 6=时装, 7=货币',
    `sub_type` TINYINT DEFAULT NULL COMMENT '子类型',
    `quality` TINYINT NOT NULL DEFAULT 1 COMMENT '品质: 1=白, 2=绿, 3=蓝, 4=紫, 5=橙',
    `level_req` INT NOT NULL DEFAULT 0 COMMENT '等级需求',
    `stack_max` INT NOT NULL DEFAULT 99 COMMENT '最大堆叠数量',
    `price` INT NOT NULL DEFAULT 0 COMMENT '售价',
    `description` VARCHAR(255) DEFAULT NULL COMMENT '描述',
    
    -- 装备属性
    `equip_type` TINYINT DEFAULT NULL COMMENT '装备位置: 1=武器, 2=衣服, 3=头盔, 4=护腕, 5=腰带, 6=鞋子, 7=戒指, 8=项链',
    `hp_bonus` INT DEFAULT 0 COMMENT '生命加成',
    `mp_bonus` INT DEFAULT 0 COMMENT '内力加成',
    `attack_bonus` INT DEFAULT 0 COMMENT '攻击加成',
    `defense_bonus` INT DEFAULT 0 COMMENT '防御加成',
    `speed_bonus` INT DEFAULT 0 COMMENT '速度加成',
    
    -- 药品效果
    `hp_restore` INT DEFAULT 0 COMMENT '恢复生命',
    `mp_restore` INT DEFAULT 0 COMMENT '恢复内力',
    `buff_id` INT DEFAULT NULL COMMENT '使用后获得BUFF',
    
    `icon` VARCHAR(100) DEFAULT NULL COMMENT '图标资源路径',
    `model` VARCHAR(100) DEFAULT NULL COMMENT '模型资源路径',
    `is_dropable` TINYINT NOT NULL DEFAULT 1 COMMENT '是否可丢弃',
    `is_sellable` TINYINT NOT NULL DEFAULT 1 COMMENT '是否可出售',
    `is_destroyable` TINYINT NOT NULL DEFAULT 1 COMMENT '是否可销毁',
    `is_bind` TINYINT NOT NULL DEFAULT 0 COMMENT '是否绑定',
    
    PRIMARY KEY (`id`),
    KEY `idx_type` (`type`),
    KEY `idx_quality` (`quality`),
    KEY `idx_level_req` (`level_req`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='道具基础表';

-- ==========================================
-- 6. 角色背包表 (role_bag)
-- ==========================================
DROP TABLE IF EXISTS `role_bag`;
CREATE TABLE `role_bag` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    `grid_index` INT NOT NULL COMMENT '背包格子索引(0-79,共80格)',
    `item_id` INT UNSIGNED NOT NULL COMMENT '道具ID',
    `count` INT NOT NULL DEFAULT 1 COMMENT '数量',
    `is_bind` TINYINT NOT NULL DEFAULT 0 COMMENT '是否绑定: 0=未绑定, 1=绑定',
    `enhance_level` INT NOT NULL DEFAULT 0 COMMENT '强化等级',
    `dur_max` INT DEFAULT NULL COMMENT '最大耐久(装备)',
    `dur_current` INT DEFAULT NULL COMMENT '当前耐久(装备)',
    `get_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '获得时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_grid` (`role_id`, `grid_index`),
    KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色背包表';

-- ==========================================
-- 7. 角色装备表 (role_equipment)
-- ==========================================
DROP TABLE IF EXISTS `role_equipment`;
CREATE TABLE `role_equipment` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    `equip_type` TINYINT NOT NULL COMMENT '装备位置: 1=武器, 2=衣服, 3=头盔, 4=护腕, 5=腰带, 6=鞋子, 7=戒指, 8=项链',
    `bag_item_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '背包物品ID(NULL表示空位)',
    `equip_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '穿戴时间',
    UNIQUE KEY `uk_role_equip` (`role_id`, `equip_type`),
    KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色装备表';

-- ==========================================
-- 8. 地图基础表 (map_base)
-- ==========================================
DROP TABLE IF EXISTS `map_base`;
CREATE TABLE `map_base` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '地图ID',
    `name` VARCHAR(50) NOT NULL COMMENT '地图名称',
    `width` INT NOT NULL COMMENT '地图宽度(像素)',
    `height` INT NOT NULL COMMENT '地图高度(像素)',
    `tile_width` INT NOT NULL DEFAULT 48 COMMENT '瓦片宽度',
    `tile_height` INT NOT NULL DEFAULT 48 COMMENT '瓦片高度',
    `map_file` VARCHAR(100) NOT NULL COMMENT '地图文件路径(.map)',
    `music` VARCHAR(100) DEFAULT NULL COMMENT '背景音乐',
    `pk_allowed` TINYINT NOT NULL DEFAULT 1 COMMENT '是否允许PK',
    `revive_map_id` INT DEFAULT NULL COMMENT '复活地图ID',
    `revive_x` INT DEFAULT NULL COMMENT '复活X坐标',
    `revive_y` INT DEFAULT NULL COMMENT '复活Y坐标',
    `level_req` INT NOT NULL DEFAULT 0 COMMENT '进入等级要求',
    `minimap_file` VARCHAR(100) DEFAULT NULL COMMENT '小地图文件',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='地图基础表';

-- ==========================================
-- 9. 门派基础表 (gang_base)
-- ==========================================
DROP TABLE IF EXISTS `gang_base`;
CREATE TABLE `gang_base` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '门派ID',
    `name` VARCHAR(20) NOT NULL COMMENT '门派名称',
    `leader_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '掌门角色ID',
    `level` INT NOT NULL DEFAULT 1 COMMENT '门派等级',
    `exp` BIGINT NOT NULL DEFAULT 0 COMMENT '门派经验',
    `money` BIGINT NOT NULL DEFAULT 0 COMMENT '门派资金',
    `announcement` VARCHAR(255) DEFAULT NULL COMMENT '门派公告',
    `description` VARCHAR(255) DEFAULT NULL COMMENT '门派描述',
    `icon` VARCHAR(100) DEFAULT NULL COMMENT '门派图标',
    `skill_1` INT UNSIGNED DEFAULT NULL COMMENT '门派技能1',
    `skill_2` INT UNSIGNED DEFAULT NULL COMMENT '门派技能2',
    `skill_3` INT UNSIGNED DEFAULT NULL COMMENT '门派技能3',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_level` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='门派基础表';

-- ==========================================
-- 10-15. 其他表结构 (简化版)
-- ==========================================
DROP TABLE IF EXISTS `role_gang`;
CREATE TABLE `role_gang` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    `gang_id` INT UNSIGNED NOT NULL COMMENT '门派ID',
    `position` TINYINT NOT NULL DEFAULT 1 COMMENT '职位: 1=帮众, 2=长老, 3=副帮主, 4=帮主',
    `contribution` INT NOT NULL DEFAULT 0 COMMENT '帮贡',
    `join_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_id` (`role_id`),
    KEY `idx_gang_id` (`gang_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色门派关系表';

DROP TABLE IF EXISTS `team`;
CREATE TABLE `team` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '队伍ID',
    `leader_id` BIGINT UNSIGNED NOT NULL COMMENT '队长ID',
    `max_member` TINYINT NOT NULL DEFAULT 5 COMMENT '最大成员数',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_leader_id` (`leader_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='队伍表';

DROP TABLE IF EXISTS `team_member`;
CREATE TABLE `team_member` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `team_id` BIGINT UNSIGNED NOT NULL COMMENT '队伍ID',
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    `join_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_team_role` (`team_id`, `role_id`),
    KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='队伍成员表';

DROP TABLE IF EXISTS `friend`;
CREATE TABLE `friend` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    `friend_id` BIGINT UNSIGNED NOT NULL COMMENT '好友ID',
    `type` TINYINT NOT NULL DEFAULT 0 COMMENT '关系: 0=好友, 1=黑名单',
    `remark` VARCHAR(50) DEFAULT NULL COMMENT '备注',
    `add_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '添加时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_friend` (`role_id`, `friend_id`),
    KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友关系表';

DROP TABLE IF EXISTS `mail`;
CREATE TABLE `mail` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '邮件ID',
    `receiver_id` BIGINT UNSIGNED NOT NULL COMMENT '接收者ID',
    `sender_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '发送者ID(NULL=系统邮件)',
    `sender_name` VARCHAR(20) NOT NULL COMMENT '发送者名称',
    `title` VARCHAR(50) NOT NULL COMMENT '邮件标题',
    `content` TEXT COMMENT '邮件内容',
    `attachment_type` TINYINT NOT NULL DEFAULT 0 COMMENT '附件类型: 0=无, 1=金币, 2=道具',
    `attachment_gold` BIGINT DEFAULT 0 COMMENT '附件金币',
    `attachment_item_id` INT UNSIGNED DEFAULT NULL COMMENT '附件道具ID',
    `attachment_item_count` INT DEFAULT 0 COMMENT '附件道具数量',
    `is_read` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读',
    `is_received` TINYINT NOT NULL DEFAULT 0 COMMENT '是否已领取附件',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间',
    `expire_time` DATETIME NOT NULL COMMENT '过期时间',
    PRIMARY KEY (`id`),
    KEY `idx_receiver_id` (`receiver_id`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邮件表';

DROP TABLE IF EXISTS `npc_base`;
CREATE TABLE `npc_base` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'NPC ID',
    `name` VARCHAR(50) NOT NULL COMMENT 'NPC名称',
    `type` TINYINT NOT NULL COMMENT '类型: 1=普通NPC, 2=商店NPC, 3=任务NPC, 4=仓库NPC, 5=传送NPC',
    `map_id` INT UNSIGNED NOT NULL COMMENT '所属地图ID',
    `x` INT NOT NULL COMMENT 'X坐标',
    `y` INT NOT NULL COMMENT 'Y坐标',
    `face` TINYINT NOT NULL DEFAULT 0 COMMENT '朝向: 0=下, 1=左, 2=右, 3=上',
    `sprite_id` INT DEFAULT NULL COMMENT '精灵ID',
    `dialog_text` TEXT COMMENT '对话文本',
    `shop_id` INT UNSIGNED DEFAULT NULL COMMENT '商店ID',
    PRIMARY KEY (`id`),
    KEY `idx_map_id` (`map_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='NPC基础表';

DROP TABLE IF EXISTS `monster_base`;
CREATE TABLE `monster_base` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '怪物ID',
    `name` VARCHAR(50) NOT NULL COMMENT '怪物名称',
    `level` INT NOT NULL DEFAULT 1 COMMENT '等级',
    `type` TINYINT NOT NULL DEFAULT 0 COMMENT '类型: 0=普通, 1=精英, 2=BOSS',
    `map_id` INT UNSIGNED NOT NULL COMMENT '所属地图ID',
    `hp` INT NOT NULL COMMENT '生命值',
    `attack` INT NOT NULL COMMENT '攻击力',
    `defense` INT NOT NULL COMMENT '防御力',
    `speed` INT NOT NULL COMMENT '速度',
    `hit` INT NOT NULL DEFAULT 50 COMMENT '命中率',
    `dodge` INT NOT NULL DEFAULT 10 COMMENT '闪避率',
    `crit` INT NOT NULL DEFAULT 5 COMMENT '暴击率',
    `ai_type` TINYINT NOT NULL DEFAULT 0 COMMENT 'AI类型: 0=被动, 1=主动攻击, 2=巡逻',
    `attack_range` INT NOT NULL DEFAULT 3 COMMENT '警戒范围(格)',
    `chase_range` INT NOT NULL DEFAULT 5 COMMENT '追击范围(格)',
    `gold_min` INT NOT NULL DEFAULT 0 COMMENT '金币掉落下限',
    `gold_max` INT NOT NULL DEFAULT 0 COMMENT '金币掉落上限',
    `exp` INT NOT NULL DEFAULT 0 COMMENT '经验掉落',
    `drop_group_id` INT UNSIGNED DEFAULT NULL COMMENT '掉落组ID',
    `sprite_id` INT DEFAULT NULL COMMENT '精灵ID',
    `respawn_time` INT NOT NULL DEFAULT 60 COMMENT '复活时间(秒)',
    PRIMARY KEY (`id`),
    KEY `idx_map_id` (`map_id`),
    KEY `idx_level` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='怪物基础表';

DROP TABLE IF EXISTS `drop_group`;
CREATE TABLE `drop_group` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '掉落组ID',
    `monster_id` INT UNSIGNED NOT NULL COMMENT '怪物ID',
    `item_id` INT UNSIGNED NOT NULL COMMENT '道具ID',
    `drop_rate` INT NOT NULL DEFAULT 100 COMMENT '掉落概率(万分比)',
    `drop_min` INT NOT NULL DEFAULT 1 COMMENT '掉落数量下限',
    `drop_max` INT NOT NULL DEFAULT 1 COMMENT '掉落数量上限',
    PRIMARY KEY (`id`),
    KEY `idx_monster_id` (`monster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='掉落组表';

DROP TABLE IF EXISTS `shop`;
CREATE TABLE `shop` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '商店ID',
    `name` VARCHAR(50) NOT NULL COMMENT '商店名称',
    `type` TINYINT NOT NULL COMMENT '类型: 1=普通商店, 2=拍卖行, 3=门派商店',
    `npc_id` INT UNSIGNED DEFAULT NULL COMMENT '关联NPC ID',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商店表';

DROP TABLE IF EXISTS `shop_goods`;
CREATE TABLE `shop_goods` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '商品ID',
    `shop_id` INT UNSIGNED NOT NULL COMMENT '商店ID',
    `item_id` INT UNSIGNED NOT NULL COMMENT '道具ID',
    `price` BIGINT NOT NULL COMMENT '价格',
    `price_type` TINYINT NOT NULL DEFAULT 1 COMMENT '货币类型: 1=金币, 2=绑定金币, 3=元宝',
    `stock` INT DEFAULT -1 COMMENT '库存(-1=无限)',
    `level_req` INT NOT NULL DEFAULT 0 COMMENT '等级要求',
    `is_available` TINYINT NOT NULL DEFAULT 1 COMMENT '是否上架',
    PRIMARY KEY (`id`),
    KEY `idx_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商店商品表';

DROP TABLE IF EXISTS `quest_base`;
CREATE TABLE `quest_base` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '任务ID',
    `name` VARCHAR(50) NOT NULL COMMENT '任务名称',
    `type` TINYINT NOT NULL COMMENT '类型: 1=主线, 2=支线, 3=日常, 4=门派, 5=活动',
    `level_req` INT NOT NULL DEFAULT 1 COMMENT '等级要求',
    `repeatable` TINYINT NOT NULL DEFAULT 0 COMMENT '是否可重复',
    `target_type` TINYINT NOT NULL COMMENT '目标类型: 1=杀怪, 2=收集, 3=对话, 4=送达',
    `target_id` INT NOT NULL COMMENT '目标ID',
    `target_count` INT NOT NULL DEFAULT 1 COMMENT '目标数量',
    `reward_exp` BIGINT NOT NULL DEFAULT 0 COMMENT '经验奖励',
    `reward_gold` BIGINT NOT NULL DEFAULT 0 COMMENT '金币奖励',
    `reward_item_id` INT UNSIGNED DEFAULT NULL COMMENT '道具奖励ID',
    `reward_item_count` INT DEFAULT NULL COMMENT '道具奖励数量',
    `description` TEXT COMMENT '任务描述',
    `npc_accept_id` INT UNSIGNED DEFAULT NULL COMMENT '接任务NPC ID',
    `npc_complete_id` INT UNSIGNED DEFAULT NULL COMMENT '交任务NPC ID',
    PRIMARY KEY (`id`),
    KEY `idx_type` (`type`),
    KEY `idx_level_req` (`level_req`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务基础表';

DROP TABLE IF EXISTS `role_quest`;
CREATE TABLE `role_quest` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    `quest_id` INT UNSIGNED NOT NULL COMMENT '任务ID',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0=进行中, 1=已完成, 2=已提交',
    `progress` INT NOT NULL DEFAULT 0 COMMENT '当前进度',
    `accept_time` DATETIME DEFAULT NULL COMMENT '接受时间',
    `complete_time` DATETIME DEFAULT NULL COMMENT '完成时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_quest` (`role_id`, `quest_id`),
    KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色任务表';

DROP TABLE IF EXISTS `buff_base`;
CREATE TABLE `buff_base` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'BUFF ID',
    `name` VARCHAR(50) NOT NULL COMMENT 'BUFF名称',
    `type` TINYINT NOT NULL COMMENT '类型: 1=增益, 2=减益, 3=控制',
    `duration` INT NOT NULL DEFAULT 0 COMMENT '持续时间(秒,0=永久)',
    `hp_change` INT NOT NULL DEFAULT 0 COMMENT '生命变化(/秒)',
    `mp_change` INT NOT NULL DEFAULT 0 COMMENT '内力变化(/秒)',
    `attack_change` INT NOT NULL DEFAULT 0 COMMENT '攻击变化',
    `defense_change` INT NOT NULL DEFAULT 0 COMMENT '防御变化',
    `speed_change` INT NOT NULL DEFAULT 0 COMMENT '速度变化',
    `hit_change` INT NOT NULL DEFAULT 0 COMMENT '命中变化',
    `dodge_change` INT NOT NULL DEFAULT 0 COMMENT '闪避变化',
    `crit_change` INT NOT NULL DEFAULT 0 COMMENT '暴击变化',
    `can_cancel` TINYINT NOT NULL DEFAULT 1 COMMENT '是否可主动取消',
    `stack_max` INT NOT NULL DEFAULT 1 COMMENT '最大叠加层数',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='BUFF效果表';

DROP TABLE IF EXISTS `role_buff`;
CREATE TABLE `role_buff` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    `buff_id` INT UNSIGNED NOT NULL COMMENT 'BUFF ID',
    `stack_count` INT NOT NULL DEFAULT 1 COMMENT '当前层数',
    `start_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '开始时间',
    `end_time` DATETIME DEFAULT NULL COMMENT '结束时间(NULL=永久)',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_buff` (`role_id`, `buff_id`),
    KEY `idx_role_id` (`role_id`),
    KEY `idx_end_time` (`end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色BUFF表';

DROP TABLE IF EXISTS `operation_log`;
CREATE TABLE `operation_log` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
    `operator_type` TINYINT NOT NULL COMMENT '操作者类型: 1=玩家, 2=GM, 3=系统',
    `operator_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '操作者ID',
    `operator_name` VARCHAR(50) DEFAULT NULL COMMENT '操作者名称',
    `action` VARCHAR(100) NOT NULL COMMENT '操作动作',
    `target_type` VARCHAR(50) DEFAULT NULL COMMENT '目标类型',
    `target_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '目标ID',
    `detail` TEXT COMMENT '详细信息(JSON)',
    `ip` VARCHAR(50) DEFAULT NULL COMMENT 'IP地址',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
    PRIMARY KEY (`id`),
    KEY `idx_operator_id` (`operator_id`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';

DROP TABLE IF EXISTS `server_config`;
CREATE TABLE `server_config` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '配置ID',
    `key` VARCHAR(50) NOT NULL COMMENT '配置键',
    `value` TEXT COMMENT '配置值',
    `type` VARCHAR(20) NOT NULL DEFAULT 'string' COMMENT '值类型',
    `description` VARCHAR(255) DEFAULT NULL COMMENT '配置说明',
    `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务器配置表';

DROP TABLE IF EXISTS `announcement`;
CREATE TABLE `announcement` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '公告ID',
    `title` VARCHAR(100) NOT NULL COMMENT '公告标题',
    `content` TEXT NOT NULL COMMENT '公告内容',
    `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型: 1=滚动, 2=弹窗, 3=系统',
    `priority` TINYINT NOT NULL DEFAULT 0 COMMENT '优先级',
    `start_time` DATETIME NOT NULL COMMENT '开始时间',
    `end_time` DATETIME DEFAULT NULL COMMENT '结束时间(NULL=永久)',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_time` (`start_time`, `end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='公告表';

-- ==========================================
-- 初始化默认数据
-- ==========================================

-- 插入服务器配置
INSERT INTO `server_config` (`key`, `value`, `type`, `description`) VALUES
('exp_rate', '1', 'float', '经验倍率'),
('gold_rate', '1', 'float', '金币倍率'),
('drop_rate', '1', 'float', '掉落倍率'),
('max_level', '100', 'int', '最高等级'),
('max_online', '1000', 'int', '最大在线人数'),
('auto_save_interval', '60', 'int', '自动保存间隔(秒)'),
('hp_regen_rate', '1', 'float', '生命回复倍率'),
('mp_regen_rate', '1', 'float', '内力回复倍率');

-- 插入初始地图
INSERT INTO `map_base` (`id`, `name`, `width`, `height`, `tile_width`, `tile_height`, `map_file`, `pk_allowed`, `revive_map_id`, `revive_x`, `revive_y`, `level_req`) VALUES
(1, '新手村', 1000, 1000, 48, 48, 'map/village.map', 0, 1, 500, 500, 0),
(2, '江湖野外', 2000, 2000, 48, 48, 'map/field.map', 1, 1, 500, 500, 5),
(3, '主城', 1500, 1500, 48, 48, 'map/city.map', 0, 1, 750, 750, 1);

-- ==========================================
-- 千年完整武学系统
-- type: 1=内功, 2=外功, 3=身法, 4=护体, 5=拳法, 6=剑法, 7=刀法, 8=枪法, 9=斧法
-- sub_type: 1=无名武功(初级), 2=进阶武功, 3=高级武功
-- ==========================================

-- 【内功类】(type=1)
INSERT INTO `skill_base` (`id`, `name`, `type`, `sub_type`, `level`, `max_level`, `exp_factor`, `description`, `hp_bonus`, `mp_bonus`, `attack_bonus`, `defense_bonus`, `speed_bonus`, `hit_bonus`, `dodge_bonus`, `crit_bonus`) VALUES
(101, '吐纳法', 1, 1, 1, 100, 100, '基础内功心法,提升生命和内力上限', 5, 5, 0, 0, 0, 0, 0, 0),
(102, '小无相功', 1, 2, 10, 100, 150, '进阶内功,增强内力恢复', 8, 10, 0, 1, 0, 0, 0, 0),
(103, '九阳神功', 1, 3, 20, 100, 200, '高级内功,大幅提升生命内力', 15, 15, 0, 2, 0, 0, 0, 0),
(104, '九阴真经', 1, 3, 30, 100, 250, '绝世内功,阴阳并济', 20, 20, 2, 3, 0, 2, 0, 1),
(105, '北冥神功', 1, 3, 40, 100, 300, '逍遥派绝学,吸取他人内力', 25, 25, 0, 2, 0, 0, 0, 2),
(106, '易筋经', 1, 3, 50, 100, 350, '少林绝学,强筋健骨', 30, 10, 0, 5, 0, 0, 0, 0);

-- 【外功类】(type=2)
INSERT INTO `skill_base` (`id`, `name`, `type`, `sub_type`, `level`, `max_level`, `exp_factor`, `description`, `hp_bonus`, `mp_bonus`, `attack_bonus`, `defense_bonus`, `speed_bonus`, `hit_bonus`, `dodge_bonus`, `crit_bonus`) VALUES
(201, '罗汉拳', 2, 1, 1, 100, 100, '基础外功,提升攻击', 0, 0, 1, 0, 0, 0, 0, 0),
(202, '太祖长拳', 2, 2, 10, 100, 150, '进阶外功,拳法刚猛', 0, 0, 3, 1, 0, 1, 0, 0),
(203, '黯然销魂掌', 2, 3, 20, 100, 200, '高级外功,威力惊人', 0, 0, 5, 2, 0, 2, 0, 1),
(204, '降龙十八掌', 2, 3, 30, 100, 250, '丐帮绝学,天下第一掌法', 5, 0, 8, 3, 0, 3, 0, 2),
(205, '打狗棒法', 2, 3, 35, 100, 250, '丐帮绝学,招式精妙', 0, 0, 6, 2, 2, 4, 1, 1);

-- 【身法类】(type=3)
INSERT INTO `skill_base` (`id`, `name`, `type`, `sub_type`, `level`, `max_level`, `exp_factor`, `description`, `hp_bonus`, `mp_bonus`, `attack_bonus`, `defense_bonus`, `speed_bonus`, `hit_bonus`, `dodge_bonus`, `crit_bonus`) VALUES
(301, '轻功术', 3, 1, 1, 100, 100, '基础身法,提升速度和闪避', 0, 0, 0, 0, 1, 0, 1, 0),
(302, '凌波微步', 3, 2, 10, 100, 150, '进阶身法,步法灵动', 0, 0, 0, 0, 3, 1, 2, 0),
(303, '神行百变', 3, 3, 20, 100, 200, '高级身法,来去如风', 0, 0, 0, 1, 5, 2, 3, 0),
(304, '梯云纵', 3, 3, 30, 100, 250, '武当绝学,纵跃如飞', 0, 0, 0, 0, 7, 3, 5, 0),
(305, '踏雪无痕', 3, 3, 40, 100, 300, '逍遥派绝学,踏雪无痕', 0, 0, 0, 0, 10, 4, 7, 1);

-- 【护体类】(type=4)
INSERT INTO `skill_base` (`id`, `name`, `type`, `sub_type`, `level`, `max_level`, `exp_factor`, `description`, `hp_bonus`, `mp_bonus`, `attack_bonus`, `defense_bonus`, `speed_bonus`, `hit_bonus`, `dodge_bonus`, `crit_bonus`) VALUES
(401, '铁布衫', 4, 1, 1, 100, 100, '基础护体,提升防御和生命', 3, 0, 0, 1, 0, 0, 0, 0),
(402, '金钟罩', 4, 2, 10, 100, 150, '进阶护体,刀枪不入', 5, 0, 0, 3, 0, 0, 0, 0),
(403, '金刚不坏', 4, 3, 20, 100, 200, '高级护体,百毒不侵', 8, 0, 0, 5, 0, 0, 0, 0),
(404, '先天功', 4, 3, 30, 100, 250, '全真教绝学,返璞归真', 10, 5, 0, 7, 0, 0, 0, 1),
(405, '蛤蟆功', 4, 3, 40, 100, 300, '西毒绝学,以守代攻', 15, 0, 0, 8, -2, 0, 0, 2);

-- 【拳法类】(type=5)
INSERT INTO `skill_base` (`id`, `name`, `type`, `sub_type`, `level`, `max_level`, `exp_factor`, `description`, `hp_bonus`, `mp_bonus`, `attack_bonus`, `defense_bonus`, `speed_bonus`, `hit_bonus`, `dodge_bonus`, `crit_bonus`) VALUES
(501, '无名拳法', 5, 1, 1, 100, 100, '初学拳法,奠定基础', 0, 0, 2, 0, 0, 1, 0, 0),
(502, '长拳', 5, 1, 5, 100, 120, '基础拳法,攻防兼备', 0, 0, 3, 1, 0, 2, 0, 0),
(503, '太极拳', 5, 2, 15, 100, 160, '武当绝学,以柔克刚', 3, 3, 4, 3, 1, 3, 2, 0),
(504, '七伤拳', 5, 2, 25, 100, 180, '崆峒绝学,先伤己后伤人', 0, 0, 7, 0, 0, 2, 0, 2),
(505, '金刚拳', 5, 3, 35, 100, 220, '少林绝学,刚猛无匹', 5, 0, 9, 5, 0, 3, 0, 1),
(506, '空明拳', 5, 3, 45, 100, 250, '周伯通创拳,空明无迹', 0, 5, 6, 2, 3, 5, 3, 1),
(507, '灵蛇拳', 5, 3, 50, 100, 280, '欧阳锋绝学,诡异难防', 0, 0, 10, 1, 2, 4, 1, 3);

-- 【剑法类】(type=6)
INSERT INTO `skill_base` (`id`, `name`, `type`, `sub_type`, `level`, `max_level`, `exp_factor`, `description`, `hp_bonus`, `mp_bonus`, `attack_bonus`, `defense_bonus`, `speed_bonus`, `hit_bonus`, `dodge_bonus`, `crit_bonus`) VALUES
(601, '无名剑法', 6, 1, 1, 100, 100, '初学剑法,剑意初成', 0, 0, 2, 0, 1, 1, 0, 0),
(602, '华山剑法', 6, 1, 5, 100, 120, '五岳剑派基础剑法', 0, 0, 3, 1, 1, 2, 0, 0),
(603, '太极剑', 6, 2, 15, 100, 160, '武当绝学,剑如太极', 2, 2, 4, 2, 2, 4, 1, 0),
(604, '独孤九剑', 6, 3, 30, 100, 220, '独孤求败绝学,无招胜有招', 0, 0, 10, 0, 3, 6, 2, 2),
(605, '辟邪剑法', 6, 3, 35, 100, 240, '葵花宝典分支,诡异狠辣', 0, 0, 12, 0, 5, 5, 1, 3),
(606, '六脉神剑', 6, 3, 40, 100, 260, '大理段氏绝学,剑气伤人', 0, 5, 14, 0, 2, 7, 0, 2),
(607, '越女剑法', 6, 3, 45, 100, 280, '越女阿青所创,轻灵飘逸', 0, 0, 8, 1, 5, 6, 3, 1);

-- 【刀法类】(type=7)
INSERT INTO `skill_base` (`id`, `name`, `type`, `sub_type`, `level`, `max_level`, `exp_factor`, `description`, `hp_bonus`, `mp_bonus`, `attack_bonus`, `defense_bonus`, `speed_bonus`, `hit_bonus`, `dodge_bonus`, `crit_bonus`) VALUES
(701, '无名刀法', 7, 1, 1, 100, 100, '初学刀法,刀意初生', 0, 0, 3, 0, 0, 1, 0, 0),
(702, '五虎断门刀', 7, 1, 5, 100, 120, '基础刀法,刀势威猛', 0, 0, 4, 1, 0, 2, 0, 0),
(703, '血刀大法', 7, 2, 15, 100, 160, '血刀门绝学,刀刀见血', 0, 0, 6, 0, 1, 3, 0, 2),
(704, '燃木刀法', 7, 3, 30, 100, 220, '少林绝学,刀如火焰', 3, 0, 9, 3, 0, 4, 0, 1),
(705, '圣火令', 7, 3, 35, 100, 240, '明教至宝,诡异难测', 0, 0, 11, 1, 2, 5, 1, 2),
(706, '反两仪刀法', 7, 3, 40, 100, 260, '华山绝学,两仪化四象', 0, 0, 10, 4, 1, 6, 0, 1),
(707, '金毛狮王刀', 7, 3, 50, 100, 300, '谢逊绝学,刀光如狮吼', 5, 0, 13, 2, 0, 5, 0, 3);

-- 【枪法类】(type=8)
INSERT INTO `skill_base` (`id`, `name`, `type`, `sub_type`, `level`, `max_level`, `exp_factor`, `description`, `hp_bonus`, `mp_bonus`, `attack_bonus`, `defense_bonus`, `speed_bonus`, `hit_bonus`, `dodge_bonus`, `crit_bonus`) VALUES
(801, '无名枪法', 8, 1, 1, 100, 100, '初学枪法,枪意初现', 0, 0, 3, 0, 0, 2, 0, 0),
(802, '杨家枪法', 8, 1, 5, 100, 120, '杨家将传承枪法', 0, 0, 4, 1, 0, 3, 0, 0),
(803, '沥泉枪法', 8, 2, 15, 100, 160, '岳飞传承枪法,气势如虹', 0, 0, 6, 2, 0, 4, 0, 1),
(804, '霸王枪', 8, 3, 30, 100, 220, '西楚霸王绝学,霸气无双', 5, 0, 10, 3, 0, 5, 0, 1),
(805, '百鸟朝凤枪', 8, 3, 35, 100, 240, '枪法绝学,如凤飞舞', 0, 0, 8, 1, 2, 7, 1, 0),
(806, '雷电枪法', 8, 3, 45, 100, 280, '雷部绝学,枪如雷电', 0, 0, 12, 0, 1, 6, 0, 2);

-- 【斧法类】(type=9)
INSERT INTO `skill_base` (`id`, `name`, `type`, `sub_type`, `level`, `max_level`, `exp_factor`, `description`, `hp_bonus`, `mp_bonus`, `attack_bonus`, `defense_bonus`, `speed_bonus`, `hit_bonus`, `dodge_bonus`, `crit_bonus`) VALUES
(901, '无名斧法', 9, 1, 1, 100, 100, '初学斧法,力沉势猛', 2, 0, 3, 1, 0, 1, 0, 0),
(902, '宣花斧', 9, 1, 5, 100, 120, '基础斧法,力大势沉', 3, 0, 5, 2, 0, 2, 0, 0),
(903, '开山斧法', 9, 2, 15, 100, 160, '进阶斧法,开山裂石', 5, 0, 7, 3, 0, 3, 0, 1),
(904, '盘古开天', 9, 3, 35, 100, 250, '上古绝学,开天辟地', 10, 0, 12, 5, -1, 4, 0, 2),
(905, '刑天舞戚', 9, 3, 50, 100, 300, '战神绝学,战意无双', 8, 0, 14, 4, 0, 5, 0, 3);

-- 插入道具数据
INSERT INTO `item_base` (`id`, `name`, `type`, `quality`, `level_req`, `price`, `description`, `hp_restore`, `mp_restore`) VALUES
(1, '金创药', 1, 1, 0, 50, '恢复100点生命值', 100, 0),
(2, '回蓝丹', 1, 1, 0, 50, '恢复100点内力值', 0, 100),
(3, '生命精华', 1, 3, 10, 500, '恢复500点生命值', 500, 0),
(4, '内力精华', 1, 3, 10, 500, '恢复500点内力值', 0, 500),
(5, '铁剑', 2, 1, 1, 100, '普通铁剑', 0, 0, 5, 0, 0),
(6, '钢剑', 2, 2, 5, 500, '精钢打造', 0, 0, 15, 0, 0),
(7, '布衣', 2, 1, 1, 100, '普通布衣', 0, 0, 0, 3, 0),
(8, '皮甲', 2, 2, 5, 500, '皮革护甲', 0, 0, 0, 8, 0),
(9, '铁矿石', 3, 1, 0, 10, '锻造材料'),
(10, '千年雪莲', 3, 4, 0, 1000, '稀有药材');

-- 插入公告
INSERT INTO `announcement` (`title`, `content`, `type`, `priority`, `start_time`) VALUES
('欢迎来到千年江湖', '欢迎各位侠士入驻千年江湖,祝大家江湖愉快!', 3, 100, NOW());

-- 插入BUFF数据
INSERT INTO `buff_base` (`id`, `name`, `type`, `duration`, `hp_change`, `description`) VALUES
(1, '打坐', 1, 0, 10, '打坐中,持续恢复生命'),
(2, '中毒', 2, 30, -5, '中毒状态,持续损失生命'),
(3, '虚弱', 2, 60, 0, '攻击和防御下降');

SELECT '千年江湖数据库初始化完成!' AS message;
SELECT CONCAT('武学总数: ', COUNT(*)) AS skill_count FROM skill_base;
