package config

import (
	"flag"
	"fmt"
	"go-server-starter/internal/enum"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	koanfenv "github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server     ServerConfig     `koanf:"server"`
	JWT        JWTConfig        `koanf:"jwt"`
	Database   DatabaseConfig   `koanf:"database"`
	Redis      RedisConfig      `koanf:"redis"`
	AsynQ        AsynQConfig         `koanf:"asynQ"`
	AlibabaCloud AlibabaCloudConfig  `koanf:"alibabaCloud"`
	Logger       LoggerConfig        `koanf:"logger"`
	GormLogger GormLoggerConfig `koanf:"gormLogger"`
	Mode       enum.ServerMode  `koanf:"-"`
}

// ConfigLoader manages koanf-based configuration loading.
type ConfigLoader struct {
	k      *koanf.Koanf
	config *Config
	mode   *enum.ServerMode
}

// NewConfigLoader creates a new config loader for the given server mode.
func NewConfigLoader(mode *enum.ServerMode) (*ConfigLoader, error) {
	config := &Config{Mode: *mode}
	k := koanf.New(".")

	// Step 1: Load defaults from DefaultConfig struct (lowest priority)
	if err := k.Load(structs.Provider(DefaultConfig, "koanf"), nil); err != nil {
		return nil, fmt.Errorf("failed to load defaults: %w", err)
	}

	cl := &ConfigLoader{k, config, mode}

	// Step 2: Load config files
	if err := cl.loadConfigFiles(); err != nil {
		return nil, err
	}

	// Step 3: Load environment variables (highest priority)
	// Automatic env works for all-lowercase keys (e.g. APP_DATABASE_PASSWORD → database.password).
	// For camelCase keys, explicit binding is required below.
	if err := k.Load(koanfenv.Provider("APP_", ".", envToConfigKey), nil); err != nil {
		fmt.Printf("warning: failed to load env vars: %v\n", err)
	}
	// Explicit bindings for camelCase config keys that automatic env can't match
	bindEnv(k,
		// jwt
		"jwt.tokenSecret",
		// server
		"server.readTimeout", "server.writeTimeout", "server.maxHeaderKB",
		"server.snowflakeNode", "server.apiPrefix",
		// logger
		"logger.fileDir", "logger.maxSize", "logger.maxAge",
		"logger.maxBackups", "logger.consoleOutput",
		// database
		"database.maxIdleConns", "database.maxOpenConns",
		"database.connMaxLifetime", "database.parseTime",
		// gormLogger
		"gormLogger.slowThreshold", "gormLogger.skipCallerLookup",
		"gormLogger.ignoreRecordNotFoundError",
		// alibabaCloud
		"alibabaCloud.accessKeyId", "alibabaCloud.accessKeySecret",
		"alibabaCloud.sms.signName", "alibabaCloud.sms.templateCode",
		"alibabaCloud.email.fromAddress", "alibabaCloud.email.fromName",
	)

	// Step 4: Unmarshal into config struct
	if err := k.UnmarshalWithConf("", config, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: false}); err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}

	return cl, nil
}

// GetConfig returns the parsed configuration.
func (cl *ConfigLoader) GetConfig() *Config {
	return cl.config
}

// loadConfigFiles loads base config and mode-specific config, merging them.
func (cl *ConfigLoader) loadConfigFiles() error {
	tryFiles := []string{"config", "config." + cl.mode.String()}

	for i, name := range tryFiles {
		path, err := cl.findConfigFile(name + ".yml")
		if err != nil {
			// Mode-specific config file is optional
			if i > 0 {
				fmt.Printf("config file %s.yml not found, skip\n", name)
				continue
			}
			return fmt.Errorf("base config file not found: %w", err)
		}

		if err := cl.k.Load(file.Provider(path), yaml.Parser()); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		fmt.Printf("✓ loaded: %s\n", path)
	}
	return nil
}

// findConfigFile searches for a config file in the current directory and ./configs/.
func (cl *ConfigLoader) findConfigFile(filename string) (string, error) {
	searchPaths := []string{filename, "configs/" + filename}
	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("%s not found in . or ./configs", filename)
}

// envToConfigKey transforms an env var name (without prefix) into a koanf config key.
// E.g. "DATABASE_PASSWORD" → "database.password"
func envToConfigKey(envVar string) string {
	return strings.ReplaceAll(strings.ToLower(envVar), "_", ".")
}

// bindEnv explicitly binds camelCase config keys to their APP_ prefixed env vars.
// Example: "alibabaCloud.accessKeyId" → env var APP_ALIBABACLOUD_ACCESSKEYID
func bindEnv(k *koanf.Koanf, keys ...string) {
	for _, key := range keys {
		envVar := "APP_" + strings.ReplaceAll(strings.ToUpper(key), ".", "_")
		if val, ok := os.LookupEnv(envVar); ok && val != "" {
			k.Set(key, val)
		}
	}
}

// ParseMode reads the server mode from APP_MODE env var or -mode flag.
func ParseMode() (*enum.ServerMode, error) {
	mode := "dev"
	if envMode := os.Getenv("APP_MODE"); envMode != "" {
		mode = envMode
	}

	flagMode := flag.String("mode", mode, "set app mode, valid modes: dev, prod, test")
	flag.Parse()

	serverMode, err := enum.ParseServerMode(*flagMode)
	if err != nil {
		return nil, err
	}
	return &serverMode, nil
}
