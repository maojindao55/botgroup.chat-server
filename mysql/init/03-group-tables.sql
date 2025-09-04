-- 群组功能相关表创建脚本
-- 这个脚本会在 MySQL 容器首次启动时自动执行

-- 使用数据库
USE botgroup_chat;

-- 创建群组表
CREATE TABLE llm_groups (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '群组名称',
    description TEXT COMMENT '群组描述',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    -- 索引
    INDEX idx_name (name),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组信息表';

-- 创建群组角色表
CREATE TABLE group_characters (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    gid BIGINT NOT NULL COMMENT '群组ID，关联groups表的id字段',
    name VARCHAR(100) NOT NULL COMMENT '角色名称',
    personality VARCHAR(100) NOT NULL DEFAULT '' COMMENT '角色性格描述',
    model VARCHAR(50) COMMENT 'AI模型名称',
    avatar TEXT COMMENT '角色头像URL',
    custom_prompt TEXT COMMENT '自定义提示词',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 索引
    INDEX idx_gid (gid),
    INDEX idx_name (name),
    INDEX idx_model (model),
    INDEX idx_created_at (created_at),
    
    -- 外键约束
    FOREIGN KEY (gid) REFERENCES llm_groups(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组角色信息表';

-- 插入测试数据 (可选)
-- INSERT INTO groups (name, description) VALUES
-- ('测试群组', '这是一个测试群组，用于演示群组功能');

-- INSERT INTO group_characters (gid, name, personality, model, avatar, custom_prompt) VALUES
-- (1, '小助手', '友善、乐于助人的AI助手', 'gpt-3.5-turbo', 'https://example.com/avatar1.jpg', '你是一个友善的AI助手，总是乐于帮助用户解决问题。');
