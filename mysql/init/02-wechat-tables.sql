-- 微信登录功能相关表创建脚本
-- 这个脚本会在 MySQL 容器首次启动时自动执行

-- 使用数据库
USE botgroup_chat;

-- 创建微信用户表
CREATE TABLE wechat_users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    uid BIGINT COMMENT '关联用户ID，关联users表的id字段',
    openid VARCHAR(64) UNIQUE NOT NULL COMMENT '微信OpenID',
    nickname VARCHAR(100) COMMENT '微信昵称',
    avatar_url TEXT COMMENT '微信头像URL',
    subscribe_scene VARCHAR(50) COMMENT '关注场景',
    qr_scene VARCHAR(100) COMMENT '二维码场景值',
    subscribe_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '关注时间',
    last_login_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '最后登录时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 索引
    INDEX idx_uid (uid),
    INDEX idx_openid (openid),
    INDEX idx_qr_scene (qr_scene),
    INDEX idx_subscribe_time (subscribe_time),
    
    -- 外键约束
    FOREIGN KEY (uid) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='微信用户信息表';

-- 插入测试数据 (可选)
-- INSERT INTO wechat_users (uid, openid, nickname, avatar_url, subscribe_scene, qr_scene) VALUES
-- (1, 'test_openid_123456', '微信测试用户', 'https://example.com/avatar.jpg', 'qr_scene', 'login_test_scene');