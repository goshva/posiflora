package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"posiflora-mvp/internal/repository"
	"github.com/julienschmidt/httprouter"
)

type StatusHandler struct {
	repo *repository.TelegramRepository
}

func NewStatusHandler(r *repository.TelegramRepository) *StatusHandler {
	return &StatusHandler{repo: r}
}

type TelegramStatusResponse struct {
	Enabled            bool   `json:"enabled"`
	ChatID             string `json:"chatId"`
	LastSentAt         int64  `json:"lastSentAt"` // unix timestamp, 0 если нет отправок
	SentCountLast7Days int    `json:"sentCountLast7Days"`
	FailedCountLast7Days int  `json:"failedCountLast7Days"`
}


func (h *StatusHandler) GetStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	shopID, _ := strconv.ParseInt(ps.ByName("shopId"), 10, 64)

	ti, err := h.repo.Get(r.Context(), shopID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var chatID string
	var enabled bool
	if ti != nil {
		chatID = maskLast4(ti.ChatID)
		enabled = ti.Enabled
	}

	stats, err := h.repo.GetSendStatsLast7Days(r.Context(), shopID)
	if err != nil {
		log.Println("failed to get send stats:", err)
	}

	var lastSentUnix int64
	if stats.LastSentAt != nil {
		lastSentUnix = stats.LastSentAt.Unix()
	} else {
		lastSentUnix = 0
	}

	resp := TelegramStatusResponse{
		Enabled:             enabled,
		ChatID:              chatID,
		LastSentAt:          lastSentUnix,
		SentCountLast7Days:  stats.Sent,
		FailedCountLast7Days: stats.Failed,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
