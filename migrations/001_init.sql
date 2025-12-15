-- =============================================
-- Полная очистка всех таблиц
-- =============================================

DROP TABLE IF EXISTS telegram_send_log CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS telegram_integrations CASCADE;
DROP TABLE IF EXISTS shops CASCADE;

-- =============================================
-- Создание таблиц (сначала shops!)
-- =============================================

CREATE TABLE shops (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE telegram_integrations (
    id         BIGSERIAL PRIMARY KEY,
    shop_id    BIGINT UNIQUE REFERENCES shops(id) ON DELETE CASCADE,
    bot_token  TEXT NOT NULL,
    chat_id    TEXT NOT NULL,
    enabled    BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE orders (
    id            BIGSERIAL PRIMARY KEY,
    shop_id       BIGINT REFERENCES shops(id) ON DELETE CASCADE,
    number        TEXT NOT NULL,
    total         NUMERIC NOT NULL,
    customer_name TEXT NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(shop_id, number)
);

CREATE TABLE telegram_send_log (
    id       BIGSERIAL PRIMARY KEY,
    shop_id  BIGINT NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    message  TEXT NOT NULL,
    status   TEXT NOT NULL CHECK (status IN ('SENT', 'FAILED')),
    error    TEXT,
    sent_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(shop_id, order_id)
);

-- =============================================
-- Сид данных (только после создания таблиц!)
-- =============================================

-- Магазины
INSERT INTO shops (id, name)
VALUES 
  (1, 'Demo Flowers Shop'),
  (2, 'Second Demo Shop'),
  (3, 'Market Shop')
ON CONFLICT (id) DO NOTHING;

-- Тестовые заказы для магазина 1
INSERT INTO orders (shop_id, number, total, customer_name)
VALUES
  (1, 'A-1001', 1890, 'Мария'),
  (1, 'A-1002', 3200, 'Алексей'),
  (1, 'A-1003', 1250, 'Ольга'),
  (1, 'A-1004', 4500, 'Дмитрий'),
  (1, 'A-1005', 890,  'Екатерина'),
  (1, 'A-1006', 2100, 'Сергей'),
  (1, 'A-1007', 5600, 'Анна'),
  (1, 'A-1008', 1750, 'Иван')
ON CONFLICT (shop_id, number) DO NOTHING;