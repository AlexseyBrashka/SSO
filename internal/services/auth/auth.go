package auth

import (
	_ "SSO/internal/config"
	_ "sync"

	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"SSO/internal/domain/models"
	"SSO/internal/lib/jwt"
	"SSO/internal/lib/logger/sl"
	"SSO/internal/storage"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

type Auth struct {
	log          *slog.Logger
	usrSaver     UserSaver
	usrProvider  UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
	regLimiter   *rate.Limiter
	loginLimiter *rate.Limiter
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
	regLimiter *rate.Limiter,
	loginLimiter *rate.Limiter,

) *Auth {
	return &Auth{
		usrSaver:     userSaver,
		usrProvider:  userProvider,
		log:          log,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
		regLimiter:   regLimiter,
		loginLimiter: loginLimiter,
	}
}

func (a *Auth) RegisterNewUser(context context.Context, email string, password string) (int64, error) {
	const op = "Auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("registering user")

	if !a.regLimiter.Allow() {
		log.Error("too many requests")

		return 0, fmt.Errorf("%s: %w", op, "too many requests")
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(context, email, passHash)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {

	const op = "Auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("username", email))

	log.Info("attempting to login user")

	if !a.regLimiter.Allow() {
		log.Error("too many requests")

		return "", fmt.Errorf("%s: %w", op, "too many requests")
	}

	// Достаем пользователя из БД
	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	// Проверяем корректность полученного пароля
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	// Получаем информацию о приложении
	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	// Создаем токен авторизации
	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, "ErrTooManyRequests")
	}

	return token, nil
}
