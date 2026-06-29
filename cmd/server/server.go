package main

// @title           Go Server Starter API
// @version         1.0
// @description     生产就绪的 Go Web 服务脚手架，内置认证、RBAC 权限控制。
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://github.com/your-username/go-server-starter/blob/main/LICENSE

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入 "Bearer {token}"，登录接口返回的 token

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"go-server-starter/internal/app"
	"go-server-starter/internal/config"
	"go-server-starter/pkg/logger"
)

func main() {
	serverMode, err := config.ParseMode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse server mode: %v\n", err)
		os.Exit(1)
	}

	cl, err := config.NewConfigLoader(serverMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := logger.NewLogger(&cl.GetConfig().Logger, *serverMode)
	defer logger.Sync()

	logger.Info(fmt.Sprintf("Starting server in %s mode", serverMode.String()))

	app := app.NewApp(cl.GetConfig(), logger)
	if err := app.Start(); err != nil {
		logger.Fatal(err.Error())
	}

	// Setup signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-quit
	logger.Info(fmt.Sprintf("Received signal: %v", sig))

	// Gracefully shutdown the server
	if err := app.Shutdown(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
