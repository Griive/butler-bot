# butler-bot
authorization


project structure

    telegram-auth-bot/
    ├── cmd/
    │   └── bot/
    │       └── main.go
    ├── internal/
    │   ├── bot/
    │   │   ├── handler.go
    │   │   └── middleware.go
    │   ├── auth/
    │   │   ├── service.go
    │   │   └── sms.go
    │   ├── storage/
    │   │   ├── repository.go
    │   │   └── redis.go
    │   └── models/
    │       └── user.go
    ├── configs/
    │   └── config.yaml
    ├── deployments/
    │   └── docker-compose.yml
    ├── go.mod
    ├── go.sum
    └── README.md

    Настройка и запуск
Создай бота в Telegram через @BotFather

Настрой конфигурацию в configs/config.yaml

Запусти Redis: docker-compose up redis -d

Собери и запусти бота: go run cmd/bot/main.go

Бот поддерживает:

Регистрацию пользователей

Отправку SMS с кодом подтверждения

Верификацию номера телефона

Повторную отправку кода

Защиту от brute-force атак

Graceful shutdown

