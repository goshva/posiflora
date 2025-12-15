package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"posiflora-mvp/internal/model"
	"posiflora-mvp/internal/repository"

	"github.com/julienschmidt/httprouter"
)

type TelegramHandler struct {
	repo *repository.TelegramRepository
}

func NewTelegramHandler(r *repository.TelegramRepository) *TelegramHandler {
	return &TelegramHandler{repo: r}
}

func (h *TelegramHandler) ConnectTelegram(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	shopID, _ := strconv.ParseInt(ps.ByName("shopId"), 10, 64)

	var payload struct {
		BotToken string `json:"botToken"`
		ChatID   string `json:"chatId"`
		Enabled  bool   `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	ti, err := h.repo.Get(r.Context(), shopID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ti == nil {
		ti = &model.TelegramIntegration{ShopID: shopID}
	}

	if payload.BotToken != "" {
		ti.BotToken = payload.BotToken
	}
	if payload.ChatID != "" && !(len(payload.ChatID) >= 4 && payload.ChatID[:4] == "****") {
		ti.ChatID = payload.ChatID
	}
	ti.Enabled = payload.Enabled

	if ti.Enabled && (ti.BotToken == "" || ti.ChatID == "") {
		http.Error(w, "botToken and chatId are required when enabled", http.StatusBadRequest)
		return
	}



	if err := h.repo.Upsert(r.Context(), ti); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем замаскированные поля клиенту
	resp := struct {
		ID        int64  `json:"ID"`
		ShopID    int64  `json:"ShopID"`
		BotToken  string `json:"BotToken"`
		ChatID    string `json:"ChatID"`
		Enabled   bool   `json:"Enabled"`
		CreatedAt string `json:"CreatedAt"`
		UpdatedAt string `json:"UpdatedAt"`
	}{
		ID:        ti.ID,
		ShopID:    ti.ShopID,
		BotToken:  "****",
		ChatID:    maskLast4(ti.ChatID),
		Enabled:   ti.Enabled,
		CreatedAt: ti.CreatedAt.Format("2006-01-02T15:04:05.000000-07:00"),
		UpdatedAt: ti.UpdatedAt.Format("2006-01-02T15:04:05.000000-07:00"),
	}

	json.NewEncoder(w).Encode(resp)
}
