package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"posiflora-mvp/internal/model"
	"posiflora-mvp/internal/service"

	"github.com/julienschmidt/httprouter"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

func (h *OrderHandler) CreateOrder(
	w http.ResponseWriter,
	r *http.Request,
	ps httprouter.Params,
) {
	shopID, err := strconv.ParseInt(ps.ByName("shopId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid shopId", http.StatusBadRequest)
		return
	}

	var payload model.OrderPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	order, status, err := h.service.CreateOrder(
		context.Background(),
		shopID,
		payload,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"id":         order.ID,
		"number":     order.Number,
		"total":      order.Total,
		"customer":   order.CustomerName,
		"sendStatus": status, // sent | failed | skipped
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
