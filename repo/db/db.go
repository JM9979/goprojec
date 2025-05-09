package db

import (
	"fmt"
	"time"

	"ginproject/entity/config"
	"ginproject/middleware/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// DB 全局数据库连接
	DB *gorm.DB
)

// Init 初始化数据库连接
func Init() error {
	var err error

	// 获取数据库配置
	dbConfig := config.GetConfig().GetDBConfig()
	if dbConfig == nil {
		return fmt.Errorf("数据库配置不存在")
	}

	// 构建DSN连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Database,
		dbConfig.Charset)

	// 自定义日志记录器 (使用项目已有的日志系统)
	customLogger := logger.New(
		&gormLogWriter{},
		logger.Config{
			SlowThreshold:             time.Second, // 慢查询阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略记录未找到的错误
			Colorful:                  false,       // 禁用颜色
		},
	)

	// 打开数据库连接
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: customLogger,
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层的SQL DB连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取底层DB连接池失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生存时间

	log.Info("数据库连接初始化成功")
	return nil
}

// GetDB 获取数据库连接实例
func GetDB() *gorm.DB {
	return DB
}

// 实现gorm的日志写入器接口
type gormLogWriter struct{}

// Printf 实现gorm日志写入器接口
func (w *gormLogWriter) Printf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Info("GORM: " + message)
}

// Close 关闭数据库连接
func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Error(fmt.Sprintf("获取DB实例失败: %v", err))
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Error(fmt.Sprintf("关闭数据库连接失败: %v", err))
		} else {
			log.Info("数据库连接已关闭")
		}
	}
}
