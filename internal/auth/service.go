package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"telegram-auth-bot/internal/models"
	"telegram-auth-bot/internal/storage"
	"time"
)

type AuthService struct {
	UserRepo         storage.UserRepository
	verificationRepo storage.VerificationRepository
	smsService       *SMSService
	codeLength       int
	maxAttempts      int
}

func NewAuthService(UserRepo storage.UserRepository, verificationRepo storage.VerificationRepository, smsService *SMSService, codeLength, maxAttempts int) *AuthService {
	return &AuthService{
		UserRepo:         UserRepo,
		verificationRepo: verificationRepo,
		smsService:       smsService,
		codeLength:       codeLength,
		maxAttempts:      maxAttempts,
	}
}

func (s *AuthService) StartVerification(ctx context.Context, phone string) error {
	// Генерация кода
	code, err := s.generateCode(s.codeLength)
	if err != nil {
		return err
	}

	// Сохранение кода в Redis
	if err := s.verificationRepo.CreateVerificationCode(ctx, phone, code); err != nil {
		return err
	}

	// Отправка SMS
	return s.smsService.SendSMS(phone, fmt.Sprintf("Your verification code: %s", code))
}

func (s *AuthService) VerifyCode(ctx context.Context, phone, code string) (bool, error) {
	verification, err := s.verificationRepo.GetVerificationCode(ctx, phone)
	if err != nil {
		return false, err
	}

	if time.Now().After(verification.ExpiresAt) {
		s.verificationRepo.DeleteVerificationCode(ctx, phone)
		return false, fmt.Errorf("verification code expired")
	}

	if verification.Attempts >= s.maxAttempts {
		s.verificationRepo.DeleteVerificationCode(ctx, phone)
		return false, fmt.Errorf("too many attempts")
	}

	if verification.Code != code {
		s.verificationRepo.IncrementAttempts(ctx, phone)
		return false, fmt.Errorf("invalid code")
	}

	// Код верный
	s.verificationRepo.DeleteVerificationCode(ctx, phone)
	return true, nil
}

func (s *AuthService) generateCode(length int) (string, error) {
	const digits = "0123456789"
	code := make([]byte, length)

	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}

	return string(code), nil
}

func (s *AuthService) RegisterOrUpdateUser(ctx context.Context, user *models.User) error {
	existingUser, err := s.UserRepo.GetUserByTelegramID(ctx, user.TelegramID)
	if err != nil {
		// Пользователь не существует, создаем нового
		return s.UserRepo.CreateUser(ctx, user)
	}

	// Обновляем данные существующего пользователя
	existingUser.PhoneNumber = user.PhoneNumber
	existingUser.FirstName = user.FirstName
	existingUser.LastName = user.LastName
	existingUser.Username = user.Username

	return nil
}

func (s *AuthService) CompleteVerification(ctx context.Context, telegramID int64) error {
	return s.UserRepo.UpdateUserVerification(ctx, telegramID, true)
}
