package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() {
	var err error

	// 构建数据库连接字符串
	dsn := buildDSN()
	log.Printf("数据库连接字符串: %s", dsn)

	// 配置GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		// 确保正确处理字符集
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 连接数据库
	DB, err = gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		log.Printf("数据库连接参数:")
		log.Printf("  HOST: %s", getEnv("MYSQL_HOST", "localhost"))
		log.Printf("  PORT: %s", getEnv("MYSQL_PORT", "3306"))
		log.Printf("  USER: %s", getEnv("MYSQL_USER", "botgroup"))
		log.Printf("  DATABASE: %s", getEnv("MYSQL_DATABASE", "botgroup_chat"))
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 获取底层的sql.DB来配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取数据库连接失败: %v", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// 强制设置会话字符集（解决乱码问题）
	charsetCommands := []string{
		"SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci",
		"SET character_set_client = utf8mb4",
		"SET character_set_connection = utf8mb4",
		"SET character_set_results = utf8mb4",
		"SET collation_connection = utf8mb4_unicode_ci",
	}

	for _, cmd := range charsetCommands {
		if err := DB.Exec(cmd).Error; err != nil {
			log.Printf("设置字符集失败 [%s]: %v", cmd, err)
		}
	}

	log.Println("字符集设置完成")

	log.Println("数据库连接成功")
}

// buildDSN 构建数据库连接字符串
func buildDSN() string {
	// 从环境变量获取数据库配置
	host := getEnv("MYSQL_HOST", "localhost")
	port := getEnv("MYSQL_PORT", "3306")
	user := getEnv("MYSQL_USER", "botgroup")
	password := getEnv("MYSQL_PASSWORD", "botgroup123")
	dbname := getEnv("MYSQL_DATABASE", "botgroup_chat")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&readTimeout=30s&writeTimeout=30s&timeout=30s",
		user, password, host, port, dbname)
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDB 获取数据库连接实例
func GetDB() *gorm.DB {
	return DB
}
