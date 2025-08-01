-- 微信登录功能数据库迁移脚本
-- 执行方式: mysql -u root -p botgroup_chat < 001_create_wechat_tables.sql

-- 检查并创建微信用户表
CREATE TABLE IF NOT EXISTS wechat_users (
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

-- 显示创建结果
SELECT 'wechat_users table created successfully' AS result;