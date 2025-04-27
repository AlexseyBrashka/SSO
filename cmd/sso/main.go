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

	AccessTTL, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_TTL"))
	if err != nil {
		log.Fatalf("Invalid ACCESS_TOKEN_TTL: %v", err)
	}
	RefreshTTL, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_TTL"))
	if err != nil {
		log.Fatalf("Invalid REFRESH_TOKEN_TTL: %v", err)
	}

	authApp := models.NewApp(uuid.New(), os.Getenv("APP_NAME"), os.Getenv("APP_SECRET"))

	loger := setupLogger(os.Getenv("ENV"))

	limiters, err := models.NewLimitersByEnv(os.Getenv("REG_RATE"), os.Getenv("REG_BURST"), os.Getenv("LOGIN_RATE"), os.Getenv("LOGIN_BURST"))
	if err != nil {
		log.Fatalf("Failed to initialize limiters: %v", err)
	}

	casherAddr := os.Getenv("CasherAddress")
	casherPort := os.Getenv("6379")
	casherPath := casherAddr + ":" + casherPort

	casherPassword := os.Getenv("CasherPassword")
	if casherPassword == "" {
		log.Fatalf("empty casher password")
	}

	casherDB := os.Getenv("CasherDBId")
	casherDBID, err := strconv.Atoi(casherDB)
	if err != nil {
		log.Fatalf("error cant use casher db id(%v)", casherDBID)
	}

	casher := models.NewRedisClient(casherPath, casherPassword, casherDBID, RefreshTTL)

	Auth := auth.New(*authApp, casher, AccessTTL, RefreshTTL, storage, limiters.RegLimiter, limiters.LoginLimiter, loger)

	grpcPortStr := os.Getenv("GRPC_PORT")
	grpcPort, err := strconv.Atoi(grpcPortStr)
	if err != nil {
		log.Fatalf("Invalid GRPC_PORT: %v", err)
	}

	app := grpcapp.New(loger, Auth, grpcPort)
	app.MustRun()

	defer app.Stop()
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
