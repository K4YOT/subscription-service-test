package service

import (
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"rec-service-test/internal/model"
	"rec-service-test/internal/repository"
	"time"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Create(req *model.CreateSubscriptionRequest) (*model.Subscription, error) {
	// Валидация
	if req.ServiceName == "" {
		return nil, errors.New("service_name is required")
	}
	if req.Price < 0 {
		return nil, errors.New("price must be non-negative")
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user_id (UUID expected)")
	}
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		return nil, errors.New("start_date must be in MM-YYYY format")
	}
	var endDate *time.Time
	if req.EndDate != "" {
		t, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			return nil, errors.New("end_date must be in MM-YYYY format")
		}
		if t.Before(startDate) {
			return nil, errors.New("end_date cannot be before start_date")
		}
		endDate = &t
	}

	sub := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}
	if err := s.repo.Create(sub); err != nil {
		slog.Error("service: create repo error", "error", err)
		return nil, errors.New("internal server error")
	}
	return sub, nil
}

func (s *SubscriptionService) GetByID(idStr string) (*model.Subscription, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, errors.New("invalid subscription id")
	}
	sub, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, errors.New("subscription not found")
	}
	return sub, nil
}

func (s *SubscriptionService) Update(idStr string, req *model.UpdateSubscriptionRequest) (*model.Subscription, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, errors.New("invalid subscription id")
	}
	// Сначала проверим, существует ли
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("subscription not found")
	}
	// Парсинг новых данных
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, errors.New("invalid user_id")
	}
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		return nil, errors.New("start_date must be in MM-YYYY format")
	}
	var endDate *time.Time
	if req.EndDate != "" {
		t, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			return nil, errors.New("end_date must be in MM-YYYY format")
		}
		if t.Before(startDate) {
			return nil, errors.New("end_date cannot be before start_date")
		}
		endDate = &t
	}
	updated := &model.Subscription{
		ID:          id,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}
	if err := s.repo.Update(updated); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *SubscriptionService) Delete(idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return errors.New("invalid subscription id")
	}
	return s.repo.Delete(id)
}

func (s *SubscriptionService) List(serviceName, userID *string, limit, offset int) ([]model.Subscription, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(serviceName, userID, limit, offset)
}

func (s *SubscriptionService) GetTotalCost(startPeriod, endPeriod string, userID, serviceName *string) (int, error) {
	start, err := time.Parse("01-2006", startPeriod)
	if err != nil {
		return 0, errors.New("start_period must be MM-YYYY")
	}
	end, err := time.Parse("01-2006", endPeriod)
	if err != nil {
		return 0, errors.New("end_period must be MM-YYYY")
	}
	if end.Before(start) {
		return 0, errors.New("end_period cannot be before start_period")
	}
	return s.repo.GetTotalCost(start, end, userID, serviceName)
}
