package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"telegram-auth-bot/internal/auth"
	"telegram-auth-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	bot         *tgbotapi.BotAPI
	authService *auth.AuthService
	ctx         context.Context
}

func NewBotHandler(bot *tgbotapi.BotAPI, authService *auth.AuthService) *BotHandler {
	return &BotHandler{
		bot:         bot,
		authService: authService,
		ctx:         context.Background(),
	}
}

func (h *BotHandler) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	switch update.Message.Text {
	case "/start":
		h.handleStart(chatID, update.Message.From)
	case "/auth":
		h.handleAuthRequest(chatID, userID)
	default:
		h.handleMessage(chatID, userID, update.Message.Text)
	}
}

func (h *BotHandler) handleStart(chatID int64, user *tgbotapi.User) {
	msg := tgbotapi.NewMessage(chatID,
		`🤖 Добро пожаловать в бот авторизации!

Для начала процесса верификации используйте команду /auth

Бот запросит ваш номер телефона и отправит SMS с кодом подтверждения.`)

	h.bot.Send(msg)
}

func (h *BotHandler) handleAuthRequest(chatID, userID int64) {
	// Проверяем, не верифицирован ли пользователь уже
	user, err := h.authService.UserRepo.GetUserByTelegramID(h.ctx, userID) //.UserRepo.GetUserByTelegramID(h.ctx, userID)
	if err == nil && user.IsVerified {
		msg := tgbotapi.NewMessage(chatID, "✅ Ваш аккаунт уже верифицирован!")
		h.bot.Send(msg)
		return
	}

	// Запрашиваем номер телефона
	msg := tgbotapi.NewMessage(chatID,
		`📱 Для верификации нам нужен ваш номер телефона.

Пожалуйста, отправьте ваш номер телефона в международном формате (например, +79123456789)`)

	msg.ReplyMarkup = h.createPhoneKeyboard()
	h.bot.Send(msg)
}

func (h *BotHandler) handleMessage(chatID, userID int64, text string) {
	// Проверяем, является ли сообщение номером телефона
	if strings.HasPrefix(text, "+") && len(text) >= 10 {
		h.handlePhoneNumber(chatID, userID, text)
		return
	}

	// Проверяем, является ли сообщение кодом подтверждения
	if len(text) == 6 && isNumeric(text) {
		h.handleVerificationCode(chatID, userID, text)
		return
	}

	// Неизвестное сообщение
	msg := tgbotapi.NewMessage(chatID,
		"Не понимаю команду. Используйте /auth для начала верификации.")
	h.bot.Send(msg)
}

func (h *BotHandler) handlePhoneNumber(chatID, userID int64, phone string) {
	// Сохраняем/обновляем пользователя
	user := &models.User{
		TelegramID:  userID,
		PhoneNumber: phone,
		FirstName:   "", // Можно получить из контекста
		LastName:    "", // Можно получить из контекста
		Username:    "", // Можно получить из контекста
		IsVerified:  false,
	}

	if err := h.authService.RegisterOrUpdateUser(h.ctx, user); err != nil {
		log.Printf("Error registering user: %v", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при обработке запроса.")
		h.bot.Send(msg)
		return
	}

	// Начинаем процесс верификации
	if err := h.authService.StartVerification(h.ctx, phone); err != nil {
		log.Printf("Error starting verification: %v", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при отправке SMS.")
		h.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf("📲 SMS с кодом подтверждения отправлено на номер %s\n\nВведите 6-значный код:", phone))
	h.bot.Send(msg)
}

func (h *BotHandler) handleVerificationCode(chatID, userID int64, code string) {
	// Получаем пользователя чтобы узнать номер телефона
	user, err := h.authService.UserRepo.GetUserByTelegramID(h.ctx, userID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Сначала отправьте номер телефона.")
		h.bot.Send(msg)
		return
	}

	// Проверяем код
	verified, err := h.authService.VerifyCode(h.ctx, user.PhoneNumber, code)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		h.bot.Send(msg)
		return
	}

	if verified {
		// Завершаем верификацию
		if err := h.authService.CompleteVerification(h.ctx, userID); err != nil {
			log.Printf("Error completing verification: %v", err)
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при завершении верификации.")
			h.bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(chatID,
			"✅ Ваш аккаунт успешно верифицирован!\n\nТеперь вы можете пользоваться всеми функциями бота.")
		h.bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный код подтверждения.")
		h.bot.Send(msg)
	}
}

func (h *BotHandler) createPhoneKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 Отправить номер телефона"),
		),
	)
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
