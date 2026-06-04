package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"log/slog"
	"rec-service-test/internal/model"
	"rec-service-test/internal/service"
)

type SubscriptionHandler struct {
	svc *service.SubscriptionService
}

func NewSubscriptionHandler(svc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

// Create godoc
// @Summary      Создать подписку
// @Description  Добавляет новую подписку в систему
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        request body model.CreateSubscriptionRequest true "Данные подписки"
// @Success      201  {object}  model.Subscription
// @Failure      400  {string}  string "Неверный запрос"
// @Failure      500  {string}  string "Внутренняя ошибка"
// @Router       /subscriptions [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("create: decode error", "error", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	sub, err := h.svc.Create(&req)
	if err != nil {
		slog.Warn("create: service error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendJSON(w, http.StatusCreated, sub)
}

// GetByID godoc
// @Summary      Получить подписку по ID
// @Description  Возвращает одну подписку
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      string  true  "UUID подписки"
// @Success      200  {object}  model.Subscription
// @Failure      404  {string}  string "Подписка не найдена"
// @Failure      400  {string}  string "Неверный ID"
// @Router       /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/subscriptions/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	sub, err := h.svc.GetByID(id)
	if err != nil {
		if err.Error() == "subscription not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	sendJSON(w, http.StatusOK, sub)
}

// Update godoc
// @Summary      Обновить подписку
// @Description  Полностью заменяет данные подписки
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id      path      string                       true  "UUID подписки"
// @Param        request body      model.UpdateSubscriptionRequest true  "Новые данные"
// @Success      200     {object}  model.Subscription
// @Failure      404     {string}  string "Подписка не найдена"
// @Failure      400     {string}  string "Ошибка в данных"
// @Router       /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/subscriptions/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	var req model.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	sub, err := h.svc.Update(id, &req)
	if err != nil {
		if err.Error() == "subscription not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	sendJSON(w, http.StatusOK, sub)
}

// Delete godoc
// @Summary      Удалить подписку
// @Description  Удаляет подписку по ID
// @Tags         subscriptions
// @Param        id   path      string  true  "UUID подписки"
// @Success      204  "No Content"
// @Failure      404  {string}  string "Подписка не найдена"
// @Router       /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/subscriptions/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	if err := h.svc.Delete(id); err != nil {
		if err.Error() == "subscription not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// List godoc
// @Summary      Список подписок
// @Description  Возвращает список подписок с фильтрацией и пагинацией
// @Tags         subscriptions
// @Produce      json
// @Param        service_name  query     string  false  "Название сервиса"
// @Param        user_id       query     string  false  "ID пользователя (UUID)"
// @Param        limit         query     int     false  "Количество записей (по умолчанию 20)"
// @Param        offset        query     int     false  "Сдвиг для пагинации"
// @Success      200  {array}   model.Subscription
// @Failure      500  {string}  string "Внутренняя ошибка"
// @Router       /subscriptions [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var serviceName, userID *string
	if sn := query.Get("service_name"); sn != "" {
		serviceName = &sn
	}
	if uid := query.Get("user_id"); uid != "" {
		userID = &uid
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	subs, err := h.svc.List(serviceName, userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendJSON(w, http.StatusOK, subs)
}

// GetTotalCost godoc
// @Summary      Суммарная стоимость подписок за период
// @Description  Вычисляет сумму цен подписок, попадающих в указанный интервал
// @Tags         subscriptions
// @Produce      json
// @Param        start_period  query     string  true  "Начало периода (MM-YYYY)"
// @Param        end_period    query     string  true  "Конец периода (MM-YYYY)"
// @Param        user_id       query     string  false "Фильтр по ID пользователя"
// @Param        service_name  query     string  false "Фильтр по названию сервиса"
// @Success      200  {object}  map[string]int
// @Failure      400  {string}  string "Неверный формат даты или параметры"
// @Router       /subscriptions/total-cost [get]
func (h *SubscriptionHandler) GetTotalCost(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	startPeriod := query.Get("start_period")
	endPeriod := query.Get("end_period")
	if startPeriod == "" || endPeriod == "" {
		http.Error(w, "start_period and end_period are required (MM-YYYY)", http.StatusBadRequest)
		return
	}
	var userID, serviceName *string
	if uid := query.Get("user_id"); uid != "" {
		userID = &uid
	}
	if sn := query.Get("service_name"); sn != "" {
		serviceName = &sn
	}
	total, err := h.svc.GetTotalCost(startPeriod, endPeriod, userID, serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendJSON(w, http.StatusOK, map[string]int{"total_cost": total})
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
