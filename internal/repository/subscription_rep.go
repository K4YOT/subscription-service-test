package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	// ЗАМЕНИТЕ ЭТОТ ПУТЬ на ваш module path из go.mod
	"rec-service-test/internal/model"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create – добавляет новую подписку в БД
func (r *SubscriptionRepository) Create(sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	)
	if err != nil {
		slog.Error("failed to create subscription in DB", "error", err, "user_id", sub.UserID)
		return err
	}
	slog.Info("subscription saved to DB", "id", sub.ID)
	return nil
}

// GetByID – получает одну подписку по её UUID
func (r *SubscriptionRepository) GetByID(id uuid.UUID) (*model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
		WHERE id = $1
	`
	row := r.db.QueryRow(query, id)

	var sub model.Subscription
	var endDate sql.NullTime

	err := row.Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&endDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("subscription not found", "id", id)
			return nil, nil // не найдено – возвращаем nil без ошибки
		}
		slog.Error("failed to get subscription by ID", "error", err, "id", id)
		return nil, err
	}

	if endDate.Valid {
		sub.EndDate = &endDate.Time
	} else {
		sub.EndDate = nil
	}
	return &sub, nil
}

// Update – полностью обновляет существующую подписку
func (r *SubscriptionRepository) Update(sub *model.Subscription) error {
	query := `
		UPDATE subscriptions
		SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5
		WHERE id = $6
	`
	result, err := r.db.Exec(query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		sub.ID,
	)
	if err != nil {
		slog.Error("failed to update subscription", "error", err, "id", sub.ID)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		slog.Warn("update: subscription not found", "id", sub.ID)
		return errors.New("subscription not found")
	}
	slog.Info("subscription updated", "id", sub.ID)
	return nil
}

// Delete – удаляет подписку по ID
func (r *SubscriptionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		slog.Error("failed to delete subscription", "error", err, "id", id)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		slog.Warn("delete: subscription not found", "id", id)
		return errors.New("subscription not found")
	}
	slog.Info("subscription deleted", "id", id)
	return nil
}

// List – возвращает список подписок с фильтрацией по service_name и user_id, а также пагинацией
func (r *SubscriptionRepository) List(serviceName, userID *string, limit, offset int) ([]model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
		WHERE ( $1 IS NULL OR service_name = $1 )
		  AND ( $2 IS NULL OR user_id = $2 )
		ORDER BY start_date DESC
		LIMIT $3 OFFSET $4
	`

	var rows *sql.Rows
	var err error

	// Передаём параметры: если filter пустой – передаём nil
	rows, err = r.db.Query(query, serviceName, userID, limit, offset)
	if err != nil {
		slog.Error("failed to list subscriptions", "error", err)
		return nil, err
	}
	defer rows.Close()

	var subscriptions []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		var endDate sql.NullTime
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&endDate,
		)
		if err != nil {
			slog.Error("failed to scan row in List", "error", err)
			return nil, err
		}
		if endDate.Valid {
			sub.EndDate = &endDate.Time
		}
		subscriptions = append(subscriptions, sub)
	}

	slog.Info("listed subscriptions", "count", len(subscriptions))
	return subscriptions, nil
}

// GetTotalCost – вычисляет суммарную стоимость подписок за выбранный период с фильтрацией
// Период задаётся началом и концом месяца (включительно).
// Подписка считается, если её период пересекается с запрошенным:
//
//	start_date <= end_period И (end_date IS NULL OR end_date >= start_period)
func (r *SubscriptionRepository) GetTotalCost(startPeriod, endPeriod time.Time, userID, serviceName *string) (int, error) {
	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE start_date <= $1
		  AND (end_date IS NULL OR end_date >= $2)
		  AND ( $3 IS NULL OR user_id = $3 )
		  AND ( $4 IS NULL OR service_name = $4 )
	`

	var total int
	err := r.db.QueryRow(query, endPeriod, startPeriod, userID, serviceName).Scan(&total)
	if err != nil {
		slog.Error("failed to calculate total cost", "error", err)
		return 0, err
	}
	slog.Info("total cost calculated", "total", total)
	return total, nil
}
