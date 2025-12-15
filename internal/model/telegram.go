package model

import "time"

type TelegramIntegration struct {
	ID        int64
	ShopID    int64
	BotToken  string
	ChatID    string
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
