package main

import (
	_ "chatapp/cmd/server/docs"
	"chatapp/cmd/server/handlers"
	"chatapp/cmd/server/middlewares"
	"chatapp/internal/clients"
	"chatapp/internal/config"
	eventlisteners "chatapp/internal/eventListeners"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"chatapp/internal/services"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

// @title           Chat Application API
// @version         1.0
// @description     API documentation for the Chat Application.
// @termsOfService  http://swagger.io/terms/

// @securityDefinitions.apikey UserTokenAuth
// @in header
// @name X-User-Token
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath  /api/v1
func main() {
	ctx := context.Background()

	app := fiber.New()
	logger := logger.NewLogger()

	//parse config
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal(ctx, fmt.Errorf("failed to load config: %w", err))
	}
	//run clients
	cls, err := clients.NewClients(ctx, cfg, logger)
	if err != nil {
		logger.Fatal(ctx, err)
	}

	defer func() {
		if err := cls.Postgres.Close(); err != nil {
			logger.Error(ctx, fmt.Errorf("failed to close Database connection: %w", err))
		}
	}()

	//init dependencies
	rps := repositories.NewRepositories(cls)
	evls := eventlisteners.NewEventListeners()
	srvs := services.NewServices(cfg, cls, logger, rps, evls)
	middlewares := middlewares.NewMiddlewares(srvs, rps, logger)
	h := handlers.NewHandlers(srvs, middlewares, rps, evls, logger)

	h.RegisterRoutes(app)

	var wg sync.WaitGroup
	wg.Add(1)
	//run server
	go func() {
		defer wg.Done()
		if err := app.Listen(":" + cfg.APPPort); err != nil {
			logger.Panic(ctx, fmt.Errorf("server listening failed: %v", err))
		}
	}()
	logger.Info(ctx, "server started")

	//handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	err = app.Shutdown()
	if err != nil {
		logger.Fatal(ctx, fmt.Errorf("server shutdown failed: %v", err))
	}
	logger.Info(ctx, "Server shutdown gracefully")

	wg.Wait()

	logger.Info(ctx, "Application exited successfully")

}
