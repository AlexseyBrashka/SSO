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
	log *slog.Logger,
	migrationPath string,
	dbName string,
	authApp models.AuthApp,
	Limiters *models.Limiters,
	grpcPort int,
	connStr string,
	tokenTTL time.Duration,
) *App {
	storage, err := postgresql.New(context.Background(), migrationPath, connStr, dbName)
	if err != nil {
		panic(err)
	}

	authService := auth.New(tokenTTL, authApp, storage, Limiters.RegLimiter, Limiters.LoginLimiter, log)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
