#!/bin/bash

# 数据库结构验证脚本
# 用于验证微信登录相关表是否正确创建

DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"3306"}
DB_NAME=${DB_NAME:-"botgroup_chat"}
DB_USER=${DB_USER:-"root"}
DB_PASS=${DB_PASS:-"123456"}

echo "正在验证数据库结构..."

# 检查数据库连接
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -e "SELECT 1;" > /dev/null 2>&1

if [ $? -ne 0 ]; then
    echo "❌ 数据库连接失败"
    exit 1
fi

echo "✅ 数据库连接成功"

# 检查 users 表
echo "检查 users 表..."
USERS_EXISTS=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SHOW TABLES LIKE 'users';" --silent)
if [ -z "$USERS_EXISTS" ]; then
    echo "❌ users 表不存在"
    exit 1
else
    echo "✅ users 表存在"
fi

# 检查 wechat_users 表
echo "检查 wechat_users 表..."
WECHAT_USERS_EXISTS=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SHOW TABLES LIKE 'wechat_users';" --silent)
if [ -z "$WECHAT_USERS_EXISTS" ]; then
    echo "❌ wechat_users 表不存在"
    echo "请执行: mysql -u$DB_USER -p$DB_PASS $DB_NAME < mysql/migrations/001_create_wechat_tables.sql"
    exit 1
else
    echo "✅ wechat_users 表存在"
fi

# 检查 wechat_users 表结构
echo "检查 wechat_users 表结构..."
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "DESCRIBE wechat_users;" > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "❌ wechat_users 表结构有问题"
    exit 1
else
    echo "✅ wechat_users 表结构正常"
fi

# 检查外键约束
echo "检查外键约束..."
FOREIGN_KEY=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS WHERE CONSTRAINT_TYPE='FOREIGN KEY' AND TABLE_NAME='wechat_users' AND CONSTRAINT_SCHEMA='$DB_NAME';" --silent)
if [ "$FOREIGN_KEY" -eq "0" ]; then
    echo "⚠️  外键约束不存在（可能是开发环境）"
else
    echo "✅ 外键约束正常"
fi

# 检查索引
echo "检查索引..."
INDEX_COUNT=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SHOW INDEX FROM wechat_users;" --silent | wc -l)
if [ "$INDEX_COUNT" -lt "4" ]; then
    echo "⚠️  索引数量可能不足"
else
    echo "✅ 索引配置正常"
fi

echo ""
echo "🎉 数据库结构验证完成！"
echo ""
echo "表结构详情:"
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "DESCRIBE wechat_users;"