# ---------- Builder ----------
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем модули для кэширования
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник из папки cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/app

# ---------- Runtime ----------
FROM alpine:latest

RUN apk --no-cache add ca-certificates bash

WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/server .

# Копируем .env в контейнер
COPY .env .

EXPOSE 8080

# Используем bash для загрузки .env перед запуском сервера
CMD ["sh", "-c", "export $(grep -v '^#' .env | xargs) && ./server"]
