package config

import (
	"go-server-starter/internal/enum"
	"time"
)

type ServerConfig struct {
	Port          int           `koanf:"port"`
	ReadTimeout   time.Duration `koanf:"readTimeout"`   // 读取超时时间
	WriteTimeout  time.Duration `koanf:"writeTimeout"`  // 写入超时时间
	MaxHeaderKB   int           `koanf:"maxHeaderKB"`   // 最大头KB数
	SnowflakeNode int64         `koanf:"snowflakeNode"` // 雪花算法节点
	APIPrefix     string        `koanf:"apiPrefix"`     // API前缀
}

type RedisConfig struct {
	Host     string `koanf:"host"`     // 主机
	Port     int    `koanf:"port"`     // 端口
	Password string `koanf:"password"` // 密码
	DB       int    `koanf:"db"`       // 数据库
}

type JWTConfig struct {
	// 签发者
	Issuer string `koanf:"issuer"` // 签发者
	// token
	TokenSecret  string                 `koanf:"tokenSecret"`  // 令牌密钥
	TokenExpires MultiTokenExpireConfig `koanf:"tokenExpires"` // 令牌过期时间
}

type MultiTokenExpireConfig struct {
	Web             time.Duration `koanf:"web"`             // 网页过期时间
	Desktop         time.Duration `koanf:"desktop"`         // 桌面软件过期时间
	Mobile          time.Duration `koanf:"mobile"`          // 移动端APP过期时间
	ChromeExtension time.Duration `koanf:"chromeExtension"` // Chrome扩展过期时间
	API             time.Duration `koanf:"api"`             // API过期时间
	Default         time.Duration `koanf:"default"`         // 默认过期时间
}

func (m *MultiTokenExpireConfig) Get(deviceType enum.DeviceType) time.Duration {
	switch deviceType {
	case enum.DeviceTypeWeb:
		return m.Web
	case enum.DeviceTypeDesktop:
		return m.Desktop
	case enum.DeviceTypeMobile:
		return m.Mobile
	case enum.DeviceTypeChromeExtension:
		return m.ChromeExtension
	case enum.DeviceTypeApi:
		return m.API
	default:
		return m.Default
	}
}

type LoggerConfig struct {
	Level         string `koanf:"level"`         // 日志级别
	FileDir       string `koanf:"fileDir"`       // 日志文件目录
	MaxSize       int    `koanf:"maxSize"`       // 日志文件最大大小
	MaxAge        int    `koanf:"maxAge"`        // 日志文件最大保存时间(天)
	MaxBackups    int    `koanf:"maxBackups"`    // 日志文件最大保存数量
	Compress      bool   `koanf:"compress"`      // 日志文件是否压缩
	ConsoleOutput bool   `koanf:"consoleOutput"` // 是否输出到控制台
}

type DatabaseConfig struct {
	Host            string        `koanf:"host"`
	Port            int           `koanf:"port"`
	Username        string        `koanf:"username"`
	Password        string        `koanf:"password"`
	Name            string        `koanf:"name"`
	MaxIdleConns    int           `koanf:"maxIdleConns"`
	MaxOpenConns    int           `koanf:"maxOpenConns"`
	ConnMaxLifetime time.Duration `koanf:"connMaxLifetime"`
	Timezone        string        `koanf:"timezone"`  // timezone configuration
	Charset         string        `koanf:"charset"`   // character set (primarily for MySQL)
	ParseTime       bool          `koanf:"parseTime"` // parse time (for MySQL)
}

type GormLoggerConfig struct {
	Level                     string        `koanf:"level"`                     // 日志级别
	SlowThreshold             time.Duration `koanf:"slowThreshold"`             // 慢查询阈值
	SkipCallerLookup          bool          `koanf:"skipCallerLookup"`          // 是否跳过调用者查找
	IgnoreRecordNotFoundError bool          `koanf:"ignoreRecordNotFoundError"` // 是否忽略记录未找到错误
}

type AsynQConfig struct {
	RedisConfig RedisConfig `koanf:"redisConfig"` // Redis配置
	Concurrency int         `koanf:"concurrency"` // 并发数
}
