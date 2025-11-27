package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"telegram-auth-bot/internal/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(host, port, password string, db int) *RedisStorage {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	return &RedisStorage{client: client}
}

func (r *RedisStorage) CreateVerificationCode(ctx context.Context, phone, code string) error {
	verification := models.VerificationCode{
		PhoneNumber: phone,
		Code:        code,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		Attempts:    0,
	}

	data, err := json.Marshal(verification)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, "verification:"+phone, data, 5*time.Minute).Err()
}

func (r *RedisStorage) GetVerificationCode(ctx context.Context, phone string) (*models.VerificationCode, error) {
	data, err := r.client.Get(ctx, "verification:"+phone).Bytes()
	if err != nil {
		return nil, err
	}

	var verification models.VerificationCode
	if err := json.Unmarshal(data, &verification); err != nil {
		return nil, err
	}

	return &verification, nil
}

func (r *RedisStorage) IncrementAttempts(ctx context.Context, phone string) error {
	verification, err := r.GetVerificationCode(ctx, phone)
	if err != nil {
		return err
	}

	verification.Attempts++
	data, err := json.Marshal(verification)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, "verification:"+phone, data, time.Until(verification.ExpiresAt)).Err()
}

func (r *RedisStorage) DeleteVerificationCode(ctx context.Context, phone string) error {
	return r.client.Del(ctx, "verification:"+phone).Err()
}

// In-memory user storage для примера (в продакшене используйте PostgreSQL)
type MemoryUserStorage struct {
	users map[int64]*models.User
}

func NewMemoryUserStorage() *MemoryUserStorage {
	return &MemoryUserStorage{
		users: make(map[int64]*models.User),
	}
}

func (m *MemoryUserStorage) CreateUser(ctx context.Context, user *models.User) error {
	user.ID = int64(len(m.users) + 1)
	user.CreatedAt = time.Now()
	m.users[user.TelegramID] = user
	return nil
}

func (m *MemoryUserStorage) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	user, exists := m.users[telegramID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (m *MemoryUserStorage) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	for _, user := range m.users {
		if user.PhoneNumber == phone {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MemoryUserStorage) UpdateUserVerification(ctx context.Context, telegramID int64, verified bool) error {
	user, exists := m.users[telegramID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.IsVerified = verified
	if verified {
		user.VerifiedAt = time.Now()
	}
	return nil
}
