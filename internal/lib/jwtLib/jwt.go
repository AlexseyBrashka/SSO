package jwtLib

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"

	"SSO/internal/domain/models"
)

func createAccessToken(ctx context.Context, User models.User, tokenUUID uuid.UUID, authApp models.AuthApp, appUUID uuid.UUID, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["uuid"] = tokenUUID
	claims["email"] = User.Email
	claims["app"] = appUUID
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["permissions"] = User.Permissions
	tokenString, err := token.SignedString([]byte(authApp.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func createRefreshToken(ctx context.Context, User models.User, tokenUUID uuid.UUID, authApp models.AuthApp, appUUID uuid.UUID, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["uuid"] = tokenUUID
	claims["email"] = User.Email
	claims["app"] = appUUID
	claims["exp"] = time.Now().Add(duration).Unix()
	tokenString, err := token.SignedString([]byte(authApp.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// TODO пророписать логику обновления токенов при изменении прав

func CreateTokenPair(ctx context.Context, User models.User, authApp models.AuthApp, app uuid.UUID, accessTTL, refTTL time.Duration) (models.Tokens, error) {
	tokensUUID, err := uuid.NewRandom()
	if err != nil {
		return models.Tokens{}, err
	}

	accessToken, err := createAccessToken(ctx, User, tokensUUID, authApp, app, accessTTL)
	if err != nil {
		return models.Tokens{}, err
	}

	refreshToken, err := createRefreshToken(ctx, User, tokensUUID, authApp, app, refTTL)
	if err != nil {
		return models.Tokens{}, err
	}

	return models.Tokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
