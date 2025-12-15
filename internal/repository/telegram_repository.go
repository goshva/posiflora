	package repository

	import (
		"context"
		"time"
		"posiflora-mvp/internal/model"

		"github.com/jackc/pgx/v5"
		"github.com/jackc/pgx/v5/pgxpool"
	)

	type TelegramRepository struct {
		pool *pgxpool.Pool
	}

	func NewTelegramRepository(pool *pgxpool.Pool) *TelegramRepository {
		return &TelegramRepository{pool: pool}
	}

	// Get возвращает TelegramIntegration по shopID
	func (r *TelegramRepository) Get(ctx context.Context, shopID int64) (*model.TelegramIntegration, error) {
		var ti model.TelegramIntegration
		err := r.pool.QueryRow(ctx, `
			SELECT id, shop_id, bot_token, chat_id, enabled, created_at, updated_at
			FROM telegram_integrations
			WHERE shop_id=$1
		`, shopID).Scan(
			&ti.ID,
			&ti.ShopID,
			&ti.BotToken,
			&ti.ChatID,
			&ti.Enabled,
			&ti.CreatedAt,
			&ti.UpdatedAt,
		)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		return &ti, nil
	}

	// Upsert вставляет или обновляет запись TelegramIntegration
	func (r *TelegramRepository) Upsert(ctx context.Context, ti *model.TelegramIntegration) error {
		return r.pool.QueryRow(ctx, `
			INSERT INTO telegram_integrations (shop_id, bot_token, chat_id, enabled, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
			ON CONFLICT (shop_id) DO UPDATE SET
				bot_token = EXCLUDED.bot_token,
				chat_id = EXCLUDED.chat_id,
				enabled = EXCLUDED.enabled,
				updated_at = NOW()
			RETURNING id, created_at, updated_at
		`, ti.ShopID, ti.BotToken, ti.ChatID, ti.Enabled).Scan(&ti.ID, &ti.CreatedAt, &ti.UpdatedAt)
	}
	type SendStats struct {
		Sent     int
		Failed   int
		LastSentAt *time.Time
	}

func (r *TelegramRepository) GetSendStatsLast7Days(ctx context.Context, shopID int64) (SendStats, error) {
    var stats SendStats

    // Кол-во успешных отправок за последние 7 дней
    _ = r.pool.QueryRow(ctx, `
        SELECT COUNT(*)
        FROM telegram_send_log
        WHERE shop_id=$1 AND status='SENT' AND sent_at >= NOW() - INTERVAL '7 days'
    `, shopID).Scan(&stats.Sent)

    // Кол-во неудачных отправок за последние 7 дней
    _ = r.pool.QueryRow(ctx, `
        SELECT COUNT(*)
        FROM telegram_send_log
        WHERE shop_id=$1 AND status='FAILED' AND sent_at >= NOW() - INTERVAL '7 days'
    `, shopID).Scan(&stats.Failed)

    // Последняя успешная отправка
    _ = r.pool.QueryRow(ctx, `
        SELECT sent_at
        FROM telegram_send_log
        WHERE shop_id=$1 AND status='SENT'
        ORDER BY sent_at DESC
        LIMIT 1
    `, shopID).Scan(&stats.LastSentAt)

    return stats, nil
}


