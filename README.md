# Telegram Bot Integration MVP для Posiflora

## Описание
Минимальная реализация интеграции Telegram-бота для уведомлений о новых заказах. Backend на Go + PostgreSQL, без frontend (используйте curl/Swagger).

## Быстрый старт

### 1. Запуск с Docker
```bash
# Клонировать проект
git clone https://github.com/goshva/posiflora.git
cd posiflora

# Запустить все сервисы
docker-compose up --build
```

### 2. Проверка работы
- **API**: http://localhost:8080
- **Swagger UI**: http://localhost:8080/swagger
- **База данных**: PostgreSQL на порту 5432

### 3. Заполнение тестовых данных
```bash
# Создать тестовый магазин (ID: 1)
curl -X POST http://localhost:8080/shops/seed

# Создать тестовые заказы
curl -X POST http://localhost:8080/orders/seed
```

## Использование API

### 1. Подключить Telegram-бота
```bash
curl -X POST http://localhost:8080/shops/1/telegram/connect \
  -H "Content-Type: application/json" \
  -d '{
    "bot_token": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
    "chat_id": "-1234567890",
    "enabled": true
  }'
```

### 2. Создать заказ (отправит уведомление если интеграция включена)
```bash
curl -X POST http://localhost:8080/shops/1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "number": "A-2005",
    "total": 2490,
    "customer_name": "Анна"
  }'
```

### 3. Проверить статус интеграции
```bash
curl http://localhost:8080/shops/1/telegram/status
```

## Режимы работы Telegram

```bash
docker-compose up --build
```

### Реальный режим
Для реальной отправки:
1. Создайте `.env` файл:
```bash
cp .env.example .env
# Редактируйте .env если нужно
```

2. Установите `jq` для тестов:
```bash
apt install jq
```

3. Запустите с реальным токеном:
```bash
TELEGRAM_MOCK_MODE=false docker-compose up --build
```

## Запуск тестов

### 1. Подготовка
```bash
cd backend

# Установить jq для парсинга JSON в тестах
apt install jq

# Скопировать конфиг
cp .env.example .env
```

### 2. Запуск тестов
```bash
apt instal jq
bash test.sh
```

### 3. Что тестируется
- `TestCreateOrderWithIntegration` - отправка уведомлений
- `TestCreateOrderIdempotency` - идемпотентность (без дублей)
- `TestCreateOrderTelegramError` - обработка ошибок Telegram

## Структура проекта
```
backend/
├── cmd/app/main.go          # Точка входа
├── internal/
│   ├── handler/             # HTTP обработчики
│   ├── service/             # Бизнес-логика
│   ├── repository/          # Работа с БД
│   └── telegram/client.go   # Клиент Telegram
├── migrations/              # Миграции БД
└── test.sh                  # Тесты
```

## Допущения
1. **Безопасность**: Токены в БД (для MVP). В production - шифрование.
2. **Устойчивость**: Нет retry/очередей. Ошибки Telegram не ломают заказы.
3. **Идемпотентность**: Уникальный индекс (shop_id, order_id) предотвращает дубли.
4. **Мок по умолчанию**: Реальная отправка отключена, нужно явно включать.

## Для собеседования
Обсудим:
- Идемпотентность через индекс БД
- Переход на очереди (RabbitMQ)
- Шифрование токенов
- Мониторинг и алертинг
- Горизонтальное масштабирование

## Полезные команды
```bash
# Остановка
docker-compose down

# Полная очистка (включая данные БД)
docker-compose down -v
```

## API Endpoints
- `POST /shops/{id}/telegram/connect` - подключение Telegram
- `POST /shops/{id}/orders` - создание заказа
- `GET /shops/{id}/telegram/status` - статус интеграции
- `GET /shops/{id}` - информация о магазине
- `POST /shops/seed` - тестовые данные магазина
- `POST /orders/seed` - тестовые данные заказов