package models

import (
	"context"
	"github.com/google/uuid"
	redisGo "github.com/redis/go-redis/v9"
	"time"
)

type RedisCasher struct {
	*redisGo.Client
	RefreshTTL time.Duration
}

func NewRedisClient(address string, password string, db int, refreshTTL time.Duration) *RedisCasher {

	client := redisGo.NewClient(&redisGo.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
	return &RedisCasher{
		Client:     client,
		RefreshTTL: refreshTTL}
}

// TODO переделать под кеширование только рефреш токенов
func (r *RedisCasher) SetUserRefresh(ctx context.Context, email string, appUUID uuid.UUID, refreshToken string) error {
	set := r.HSet(ctx, email, appUUID.String(), refreshToken)
	if set.Err() != nil {
		return set.Err()
	}
	exp := r.Expire(ctx, email, r.RefreshTTL)
	if exp.Err() != nil {
		return exp.Err()
	}
	return nil
}
func (r *RedisCasher) GetUserRefresh(ctx context.Context, email string, appUUID uuid.UUID) (string, error) {
	var result string

	err := r.HGet(ctx, email, appUUID.String()).Scan(&result)
	if err != nil {
		return "", err
	}
	return result, nil

}
func (r *RedisCasher) BlockUserRefresh(ctx context.Context, email string, appUUID uuid.UUID) error {

	err := r.HDel(ctx, email, appUUID.String()).Err()
	if err != nil {
		return err
	}
	return nil
}
