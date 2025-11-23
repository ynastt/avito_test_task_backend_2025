package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/ynastt/avito_test_task_backend_2025/internal/handlers"
	"github.com/ynastt/avito_test_task_backend_2025/internal/repository"
	"github.com/ynastt/avito_test_task_backend_2025/internal/service"
	pr "github.com/ynastt/avito_test_task_backend_2025/internal/service/pullrequest"
	"github.com/ynastt/avito_test_task_backend_2025/internal/service/team"
	"github.com/ynastt/avito_test_task_backend_2025/internal/service/user"
	"github.com/ynastt/avito_test_task_backend_2025/pkg/database"
	"github.com/ynastt/avito_test_task_backend_2025/server"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env file not found: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	db, err := database.NewPostgresDB(database.Config{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		Username: os.Getenv("POSTGRES_USERNAME"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSL"),
	}, logger)
	if err != nil {
		logger.Error("failed to initialize db", "error", err.Error())
		os.Exit(1)
	}

	// Миграция
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("Migration driver error:", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		logger.Error("migrate init error", slog.Any("error", err))
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("migration error", slog.Any("error", err))
		os.Exit(1)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Error error occured on closing database connection", slog.Any("error", err))
		} else {
			logger.Info("Database connection closed gracefully")
		}
	}()

	dbInstance := database.NewDB(db)
	txManager, err := database.NewTransactionManager(db)
	if err != nil {
		logger.Error("error creating transaction manager", slog.Any("error", err))
		os.Exit(1)
	}

	teamRepo := repository.NewTeamRepository(dbInstance)
	userRepo := repository.NewUserRepository(dbInstance)
	prRepo := repository.NewPullRequestRepository(dbInstance)
	statsRepo := repository.NewStatsRepository(dbInstance)

	services := &service.Services{
		TeamService:        team.NewTeamService(teamRepo, userRepo, txManager, logger),
		UserService:        user.NewUserService(userRepo, prRepo, txManager, logger),
		PullRequestService: pr.NewPullRequestService(prRepo, userRepo, txManager, logger),
		StatsService:       service.NewStatsService(statsRepo, logger),
	}

	handlers := handlers.NewHandler(services, logger)

	srv := new(server.Server)
	serverErrors := make(chan error, 1)
	go func() {
		if err := srv.Run(os.Getenv("SERVER_PORT"), handlers.InitRoutes()); err != nil {
			serverErrors <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-quit:
		logger.Info("Gracefully Shutting Down")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("Error occured on server shutting down", slog.Any("error", err))
		}
		// catching ctx.Done(). timeout of 5 seconds.
		<-ctx.Done()
		logger.Info("Timeout of 5 seconds.")

		logger.Info("Server stopped gracefully")
	case err := <-serverErrors:
		logger.Error("Error occured while running server", slog.Any("error", err))
		os.Exit(1)
	}
}
