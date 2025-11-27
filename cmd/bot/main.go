package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"telegram-auth-bot/internal/auth"
	"telegram-auth-bot/internal/bot"
	"telegram-auth-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
)

func main() {
	// Загрузка конфигурации
	viper.SetConfigFile("configs/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	// Инициализация бота
	botToken := viper.GetString("bot.token")
	if botToken == "YOUR_BOT_TOKEN" {
		log.Fatal("Please set your bot token in config.yaml")
	}

	telegramBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	log.Printf("Authorized on account %s", telegramBot.Self.UserName)

	// Инициализация хранилищ
	redisStorage := storage.NewRedisStorage(
		viper.GetString("redis.host"),
		viper.GetString("redis.port"),
		viper.GetString("redis.password"),
		viper.GetInt("redis.db"),
	)

	userStorage := storage.NewMemoryUserStorage()

	// Инициализация SMS сервиса
	smsService := auth.NewSMSService(viper.GetString("auth.sms_gateway_url"))

	// Инициализация сервиса авторизации
	authService := auth.NewAuthService(
		userStorage,
		redisStorage,
		smsService,
		viper.GetInt("auth.code_length"),
		viper.GetInt("auth.max_attempts"),
	)

	// Инициализация обработчика
	botHandler := bot.NewBotHandler(telegramBot, authService)

	// Настройка обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = viper.GetInt("bot.timeout")

	updates := telegramBot.GetUpdatesChan(u)

	// Обработка сигналов для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Обработка сообщений
	log.Println("Bot started...")
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			go botHandler.HandleUpdate(update)
		}
	}
}
