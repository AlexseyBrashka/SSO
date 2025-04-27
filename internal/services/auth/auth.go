package auth

import (
	"SSO/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go/token"
	_ "sync"

	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"SSO/internal/domain/models"
	"SSO/internal/lib/jwtLib"
	"SSO/internal/lib/logger/sl"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Storage interface {
	User(ctx context.Context, email string) (models.User, error)
	UserWithPermissions(ctx context.Context, email string, appUUID uuid.UUID) (models.User, error)
	SaveUser(ctx context.Context, uuid uuid.UUID, email string, passHash []byte) error
	SaveApp(ctx context.Context, appUUID uuid.UUID, name string) (uuid.UUID, error)
	DeletePermission(ctx context.Context, permUUID uuid.UUID, appUUID uuid.UUID) error
	SavePermission(ctx context.Context, permUUID uuid.UUID, appUUID uuid.UUID, permission string) (models.Permission, error)
	AddUserPermissions(ctx context.Context, email string, appUUID uuid.UUID, permUUID uuid.UUID) error
	RevokeUserPermissions(ctx context.Context, email string, appUUID uuid.UUID, permUUID uuid.UUID) error
}

type Auth struct {
	authApp      models.AuthApp
	casher       *models.RedisCasher
	accessTTL    time.Duration
	refreshTTL   time.Duration
	storage      Storage
	regLimiter   *rate.Limiter
	loginLimiter *rate.Limiter
	log          *slog.Logger
}

func New(
	AuthApp models.AuthApp,
	Casher *models.RedisCasher,
	AccessTTL time.Duration,
	RefreshTTL time.Duration,
	Storage Storage,
	RegLimiter *rate.Limiter,
	LoginLimiter *rate.Limiter,
	Log *slog.Logger,

) *Auth {
	return &Auth{
		authApp:      AuthApp,
		casher:       Casher,
		accessTTL:    AccessTTL,
		refreshTTL:   RefreshTTL,
		storage:      Storage,
		regLimiter:   RegLimiter,
		loginLimiter: LoginLimiter,
		log:          Log,
	}
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (uuid.UUID, error) {
	const op = "Auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("registering user")

	if !a.regLimiter.Allow() {
		log.Error("too many requests")

		return uuid.Nil, fmt.Errorf("%s: %w", op, "too many requests")
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	userUUID, err := uuid.NewRandom()
	if err != nil {
		log.Error("failed to generate uuid", sl.Err(err))
	}

	err = a.storage.SaveUser(ctx, userUUID, email, passHash)

	if err != nil {
		log.Error("failed to save user", sl.Err(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}
	return userUUID, nil
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appUUID uuid.UUID,
) (string, string, error) {

	const op = "Auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("username", email))

	log.Info("attempting to login user")

	if !a.regLimiter.Allow() {
		log.Error("too many requests")

		return "", "", fmt.Errorf("%s: %w", op, "too many requests")
	}

	user, err := a.storage.UserWithPermissions(ctx, email, appUUID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	tokenPair, err := a.createTokenPair(ctx, user, appUUID)
	if err != nil {
		a.log.Error("failed to create token pair", sl.Err(err))

		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	return tokenPair.AccessToken, tokenPair.RefreshToken, nil
}

func (a *Auth) Logout(ctx context.Context, email string, appUUID uuid.UUID) error {
	op := "Auth.Logout"
	err := a.casher.BlockUserRefresh(ctx, email, appUUID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *Auth) AddPermission(ctx context.Context, appUUID uuid.UUID, permission string) (uuid.UUID, error) {
	op := "Auth.AddPermission"
	permissionUUID, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = a.storage.SavePermission(ctx, permissionUUID, appUUID, permission)
	return permissionUUID, nil
}

func (a *Auth) RemovePermission(ctx context.Context, appUUID uuid.UUID, permissionUUID uuid.UUID) error {
	op := "Auth.RemovePermission"

	err := a.storage.DeletePermission(ctx, appUUID, permissionUUID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *Auth) GrantPermission(ctx context.Context, email string, AppUUID uuid.UUID, permissionUUID uuid.UUID) (accessToken string, refreshToken string, err error) {
	op := "Auth.GrantPermission"

	err = a.storage.AddUserPermissions(ctx, email, permissionUUID, AppUUID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	user, err := a.storage.UserWithPermissions(ctx, email, AppUUID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	tokenPair, err := a.createTokenPair(ctx, user, AppUUID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	return tokenPair.AccessToken, tokenPair.RefreshToken, nil
}

func (a *Auth) RevokePermission(ctx context.Context, email string, AppUUID uuid.UUID, permissionUUID uuid.UUID) (accessToken string, refreshToken string, err error) {
	op := "Auth.RevokePermission"

	err = a.storage.RevokeUserPermissions(ctx, email, permissionUUID, AppUUID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	user, err := a.storage.UserWithPermissions(ctx, email, AppUUID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	tokenPair, err := a.createTokenPair(ctx, user, AppUUID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	return tokenPair.AccessToken, tokenPair.RefreshToken, nil
}
func (a *Auth) RefreshToken(ctx context.Context, RefreshToken string) (accessToken string, refreshToken string, err error) {
	op := "Auth.RefreshToken"
	refToken, err := jwt.Parse(RefreshToken,
		func(refToken *jwt.Token) (interface{}, error) {
			tokenChecked, err := refToken.SignedString(a.authApp.Secret)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			return tokenChecked, nil
		})
	if claims, ok := refToken.Claims.(jwt.MapClaims); ok && refToken.Valid {
		issuedAt := time.Unix(int64(claims["iat"].(float64)), 0)
		if time.Now().After(issuedAt) {
			return "", "", ErrTokenExpired
		} else {
			fmt.Println("Token is valid")
		}
	} else {
		fmt.Println(err)
	}

}
func (a *Auth) createTokenPair(ctx context.Context, user models.User, appUUID uuid.UUID) (TokenPair models.Tokens, err error) {
	const op = "Auth.createTokenPair"
	tokenPair, err := jwtLib.CreateTokenPair(ctx, user, a.authApp, appUUID, a.accessTTL, a.refreshTTL)

	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return models.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	// удаляем старый токен из кеша
	if err := a.casher.BlockUserRefresh(ctx, user.Email, appUUID); err != nil {
		a.log.Error("failed to delete token", sl.Err(err))
		return models.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}
	//записываем новый токен в кеш
	if err := a.casher.SetUserRefresh(ctx, user.Email, appUUID, tokenPair.RefreshToken); err != nil {
		a.log.Error("failed to save token", sl.Err(err))
		return models.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}
	return tokenPair, nil
}
