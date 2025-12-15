package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"posiflora-mvp/internal/model"
	"posiflora-mvp/internal/repository"
	"posiflora-mvp/internal/telegram"
	"strings"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}
type OrderService struct {
	orderRepo      *repository.OrderRepository
	telegramRepo   *repository.TelegramRepository
	telegramClient telegram.Client
}

func NewOrderService(
	orderRepo *repository.OrderRepository,
	telegramRepo *repository.TelegramRepository,
	telegramClient telegram.Client,
) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		telegramRepo:   telegramRepo,
		telegramClient: telegramClient,
	}
}

func (s *OrderService) CreateOrder(
	ctx context.Context,
	shopID int64,
	payload model.OrderPayload,
) (*model.Order, string, error) {

	order, _, err := s.orderRepo.Upsert(ctx, shopID, payload)
	if err != nil {
		return nil, "", err
	}

	// 1. Идемпотентность
	exists, err := s.orderRepo.HasSendLog(ctx, shopID, order.ID)
	if err != nil {
		return nil, "", err
	}
	if exists {
		return order, "skipped", nil
	}

	// 2. Проверяем интеграцию
	ti, err := s.telegramRepo.Get(ctx, shopID)
	if err != nil || ti == nil || !ti.Enabled {
		_ = s.orderRepo.LogSend(
			ctx,
			shopID,
			order.ID,
			"",
			"SKIPPED",
			nil,
		)
		return order, "skipped", nil
	}

	// 3. Отправка
	msg := fmt.Sprintf(
		"Новый заказ %s на сумму %.2f ₽, клиент %s",
		order.Number,
		order.Total,
		order.CustomerName,
	)

	err = s.telegramClient.SendMessage(ctx, ti.BotToken, ti.ChatID, msg)

	status := "SENT"
	var errText *string
	if err != nil {
		status = "FAILED"
		e := err.Error()
		errText = &e
	}

	_ = s.orderRepo.LogSend(
		ctx,
		shopID,
		order.ID,
		msg,
		status,
		errText,
	)

	return order, strings.ToLower(status), nil
}

