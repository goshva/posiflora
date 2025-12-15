package service_test

import (
	"context"
	"testing"
	"posiflora-mvp/internal/model"
	"posiflora-mvp/internal/repository"
	"posiflora-mvp/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MockTelegram struct {
	Sent []string
}

func (m *MockTelegram) SendMessage(ctx context.Context, botToken, chatID, text string) error {
	m.Sent = append(m.Sent, text)
	return nil
}

func TestOrderServiceE2E(t *testing.T) {
	dbURL := "postgres://postgres:postgres@localhost:5432/posiflora?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	orderRepo := repository.NewOrderRepository(pool)
	telegramRepo := repository.NewTelegramRepository(pool)
	mockTelegram := &MockTelegram{}

	service := service.NewOrderService(orderRepo, telegramRepo, mockTelegram)

	// Вставляем Telegram integration напрямую через pool
	_, _ = pool.Exec(context.Background(), `
		INSERT INTO telegram_integrations (shop_id, bot_token, chat_id, enabled)
		VALUES (1, 'token', 'chat', true)
		ON CONFLICT DO NOTHING
	`)

	payload := model.OrderPayload{
		Number:       "E2E123",
		Total:        100.5,
		CustomerName: "Tester",
	}

   resp, err := service.CreateOrder(context.Background(), 1, payload)
	if err != nil {
		t.Fatal(err)
	}

	if resp["sendStatus"] != "sending" {
		t.Fatalf("expected sending, got %v", resp["sendStatus"])
	}

	if len(mockTelegram.Sent) != 1 {
		t.Fatalf("expected 1 telegram message, got %d", len(mockTelegram.Sent))
	}
}
