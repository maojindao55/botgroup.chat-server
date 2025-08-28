-- 为用户表添加 openid 字段
-- 执行时间: 2025-03-26

USE botgroup_chat;

-- 添加 openid 字段
ALTER TABLE users 
ADD COLUMN openid VARCHAR(100) DEFAULT NULL COMMENT '微信OpenID' 
AFTER phone;

-- 为 openid 字段创建唯一索引（NULL 值不会冲突）
CREATE UNIQUE INDEX idx_users_openid ON users(openid);

-- 删除 phone 字段的唯一索引
ALTER TABLE users DROP INDEX phone;

-- 更新字段注释
ALTER TABLE users 
MODIFY COLUMN phone VARCHAR(11) DEFAULT '' COMMENT '手机号（非必填）',
MODIFY COLUMN openid VARCHAR(100) DEFAULT '' COMMENT '微信OpenID（唯一标识）';

-- 显示表结构确认
DESCRIBE users;
