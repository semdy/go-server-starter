package app

import (
	"context"
	"fmt"
	"go-server-starter/internal/config"
	"go-server-starter/internal/enum"
	"go-server-starter/internal/database/migration"
	"go-server-starter/internal/handler"
	"go-server-starter/internal/middleware"
	"go-server-starter/internal/repo"
	"go-server-starter/internal/router"
	"go-server-starter/internal/seed"
	"go-server-starter/internal/service"
	"go-server-starter/pkg/auth"
	"go-server-starter/pkg/database"
	"go-server-starter/pkg/jwt"
	"go-server-starter/pkg/logger"
	"go-server-starter/pkg/redis"
	"go-server-starter/pkg/snowflake"
	"go-server-starter/pkg/taskq"
	"go-server-starter/pkg/translator"
	"go-server-starter/pkg/validator"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type App struct {
	config     *config.Config
	engine     *gin.Engine
	server     *http.Server
	db         *database.DB
	redis      *redis.Client
	logger     *zap.Logger
	snowflake  *snowflake.Snowflake
	translator *translator.Translator
	ratelimit  *middleware.RateLimit
	jwt        *jwt.JWT
	auth       auth.Auth
	handler    handler.Handler
	repo       repo.Repo
	service    service.Service
	seed        seed.Seed
	taskqClient *taskq.Client
	taskqServer *taskq.Server
}

func NewApp(config *config.Config, logger *zap.Logger) *App {
	app := &App{config: config, logger: logger}
	return app
}

func (a *App) Start() error {
	var serverConfig = a.config.Server
	if a.config.Mode != enum.ServerModeDev {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	var isDev = a.config.Mode == enum.ServerModeDev
	// 初始化gin引擎
	a.engine = gin.New()
	a.engine.Use(gin.Recovery())
	a.engine.Use(middleware.ZapLogger(a.logger.Named("GIN")))
	a.engine.Use(middleware.ZapRecovery(a.logger.Named("GIN-RECOVERY"), isDev))

	if isDev {
		a.engine.Use(middleware.CORS())
	}

	// 初始化validator
	if err := validator.Init(); err != nil {
		return err
	}

	// 初始化翻译器
	trans, err := translator.NewTranslator()
	if err != nil {
		return err
	}
	a.translator = trans
	a.engine.Use(middleware.Translations(trans))

	// 初始化http服务器
	a.server = &http.Server{
		Addr:           fmt.Sprintf(":%d", serverConfig.Port),
		Handler:        a.engine,
		ReadTimeout:    serverConfig.ReadTimeout,
		WriteTimeout:   serverConfig.WriteTimeout,
		MaxHeaderBytes: serverConfig.MaxHeaderKB * 1024,
	}

	// 初始化数据库日志
	gormLogger, err := logger.NewGormLogger(a.logger.Named("GORM"), a.config.GormLogger)
	if err != nil {
		return err
	}

	// 连接数据库
	db, err := database.NewDB(a.config.Database, gormLogger, nil)
	if err != nil {
		return err
	}
	a.db = db

	// run database migrations (goose)
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	if err := migration.Run(sqlDB); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	// 初始化redis
	redis, err := redis.NewClient(a.config.Redis, a.logger.Named("redis"), context.Background())
	if err != nil {
		return err
	}
	a.redis = redis

	// 初始化任务队列
	taskqClient, err := taskq.NewClient(a.config.AsynQ, a.logger.Named("TASKQ-CLIENT"))
	if err != nil {
		return fmt.Errorf("init taskq client: %w", err)
	}
	a.taskqClient = taskqClient

	taskqServer := taskq.NewServer(a.config.AsynQ, taskq.ServerConfig{
		Concurrency: a.config.AsynQ.Concurrency,
		Queues: map[string]int{
			"default": 3,
			"low":     1,
		},
	}, a.logger, nil) // nil alerter = default no-op

	// 注册任务处理器（示例：欢迎邮件）
	taskqServer.HandleFunc(taskq.TaskEmailWelcome, taskq.HandleEmailWelcome)

	// 启动后台 worker
	go func() {
		if err := taskqServer.Start(); err != nil {
			a.logger.Error("taskq worker stopped with error", zap.Error(err))
		}
	}()
	a.taskqServer = taskqServer

	// snowflake
	snowflake, err := snowflake.NewSnowflake(a.config.Server.SnowflakeNode)
	if err != nil {
		return err
	}
	a.snowflake = snowflake

	// 初始化repo
	a.repo = repo.NewRepo(
		db.DB,
		a.logger.Named("REPO"),
	)

	// 初始化seed
	a.seed = seed.NewSeed(
		a.repo,
		a.logger.Named("SEED"),
	)

	// 运行seed
	if err := a.seed.Run(); err != nil {
		return err
	}

	// 初始化jwt
	a.jwt = jwt.NewJWT(
		&a.config.JWT,
		a.logger.Named("JWT"),
	)

	// 初始化service
	a.service = service.NewService(
		a.db.DB,
		a.config,
		a.jwt,
		a.redis,
		a.snowflake,
		a.repo,
		a.taskqClient,
		a.logger.Named("SERVICE"),
	)
	// 初始化auth
	a.auth = auth.NewAuth(a.service, a.logger.Named("AUTH"))
	// 初始化handler
	a.handler = handler.NewHandler(
		a.service,
		a.logger.Named("HANDLER"),
	)
	// 初始化ratelimit
	a.ratelimit = middleware.NewRateLimit(a.redis, a.logger.Named("RATELIMIT"))
	// 每分钟限流 100 次
	a.engine.Use(a.ratelimit.RateLimit(100))
	// 初始化router
	router := router.NewRouter(
		a.handler,
		a.engine.Group(serverConfig.APIPrefix),
		a.jwt,
		a.auth,
		a.ratelimit,
	)
	router.SetupRoutes()

	go func() {
		a.logger.Info(fmt.Sprintf("Server starting on port %d...", serverConfig.Port))
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("Server error", zap.Error(err))
		}
	}()

	return nil
}

func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	a.logger.Info("Gracefully shutting down...")

	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			a.logger.Error("Failed to shutdown server", zap.Error(err))
			if err := a.server.Close(); err != nil {
				a.logger.Error("Failed to force close server", zap.Error(err))
				return err
			}
		}
	}

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.logger.Error("Failed to close database", zap.Error(err))
		}
	}

	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			a.logger.Error("Failed to close redis", zap.Error(err))
		}
	}

	if a.taskqServer != nil {
		a.taskqServer.Shutdown()
	}
	if a.taskqClient != nil {
		if err := a.taskqClient.Close(); err != nil {
			a.logger.Error("Failed to close taskq client", zap.Error(err))
		}
	}

	a.logger.Info("Server shutdown successfully")
	return nil
}
