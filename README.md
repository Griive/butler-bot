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