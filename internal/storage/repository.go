package storage

import (
	"context"
	"telegram-auth-bot/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*models.User, error)
	UpdateUserVerification(ctx context.Context, telegramID int64, verified bool) error
}

type VerificationRepository interface {
	CreateVerificationCode(ctx context.Context, phone, code string) error
	GetVerificationCode(ctx context.Context, phone string) (*models.VerificationCode, error)
	IncrementAttempts(ctx context.Context, phone string) error
	DeleteVerificationCode(ctx context.Context, phone string) error
}
