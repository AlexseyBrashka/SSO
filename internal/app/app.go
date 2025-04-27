package app

import (
	"SSO/internal/domain/models"
	"SSO/internal/storage/postgresql"
	"context"
	"log/slog"
	"time"

	grpcapp "SSO/internal/app/grpc"
	"SSO/internal/services/auth"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	authApp models.AuthApp,
	casher *models.RedisCasher,
	accTokenTTL time.Duration,
	refTokenTTL time.Duration,
	migrationPath string,
	dbName string,
	limiters *models.Limiters,
	log *slog.Logger,
	grpcPort int,
	connStr string,
) *App {
	storage, err := postgresql.New(context.Background(), migrationPath, connStr, dbName)
	if err != nil {
		panic(err)
	}

	authService := auth.New(authApp, casher, accTokenTTL, refTokenTTL, storage, limiters.RegLimiter, limiters.LoginLimiter, log)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
