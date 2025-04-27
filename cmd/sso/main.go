// `cmd/sso/migrator.go`
package main

import (
	grpcapp "SSO/internal/app/grpc"
	"SSO/internal/domain/models"
	"SSO/internal/services/auth"
	"SSO/internal/storage/postgresql"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbName := os.Getenv("DB_NAME")
	dbConn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		dbName,
		os.Getenv("DB_SSLMODE"),
	)

	storage, err := postgresql.New(context.Background(), os.Getenv("DB_MIGRATIONS"), dbConn, dbName)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Stop()
	TTL, err := time.ParseDuration(os.Getenv("TOKEN_TTL"))
	if err != nil {
		log.Fatalf("Invalid ACCESS_TOKEN_TTL: %v", err)
	}

	authApp := models.NewApp(uuid.New(), os.Getenv("APP_NAME"), os.Getenv("APP_SECRET"))

	loger := setupLogger(os.Getenv("ENV"))

	limiters, err := models.NewLimitersByEnv(os.Getenv("REG_RATE"), os.Getenv("REG_BURST"), os.Getenv("LOGIN_RATE"), os.Getenv("LOGIN_BURST"))
	if err != nil {
		log.Fatalf("Failed to initialize limiters: %v", err)
	}

	Auth := auth.New(*authApp, models.NewRedisClient("localhost:6379", "1", 0, TTL), TTL, TTL, storage, limiters.RegLimiter, limiters.LoginLimiter, loger)

	grpcPortStr := os.Getenv("GRPC_PORT")
	grpcPort, err := strconv.Atoi(grpcPortStr)
	if err != nil {
		log.Fatalf("Invalid GRPC_PORT: %v", err)
	}
	app := grpcapp.New(loger, Auth, grpcPort)
	app.MustRun()

	defer app.Stop()
	//TODO реализовать систему redis кэша токенов, с TTL, возможность перешифроки паролев новым ключем.
	// TODO: спросить у прыепода, как сделать разлогинивание, обновление пользователя ( стоит ли делать пароли для сервисов),
	// что бы из вне фиг обратиться

	//todo проверить защиту от брутфорса
}

func setupLogger(env string) *slog.Logger {
	var loger *slog.Logger

	switch env {
	case envLocal:
		loger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		loger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		loger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		loger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	//TODO: починить логер ибо он не чекает окружение
	return loger
}
