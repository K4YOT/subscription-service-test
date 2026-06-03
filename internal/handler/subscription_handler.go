package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"rec-service-test/internal/model"
	"rec-service-test/internal/service"
	"time"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
}

func NewSubscriptionHandler(s *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: s}
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("invalid JSON", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Парсинг даты из "MM-YYYY"
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		http.Error(w, "invalid start_date format (MM-YYYY)", http.StatusBadRequest)
		return
	}

	var endDate *time.Time
	if req.EndDate != "" {
		t, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			http.Error(w, "invalid end_date format", http.StatusBadRequest)
			return
		}
		endDate = &t
	}

	sub := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := h.service.Create(sub); err != nil {
		slog.Error("create failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("subscription created", "id", sub.ID, "user_id", sub.UserID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}
