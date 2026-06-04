package repository

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"rec-service-test/internal/model"
	"time"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(sub *model.Subscription) error {
	query := `INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate)
	if err != nil {
		slog.Error("create: db error", "error", err)
		return err
	}
	slog.Info("subscription created", "id", sub.ID)
	return nil
}

func (r *SubscriptionRepository) GetByID(id uuid.UUID) (*model.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date
	          FROM subscriptions WHERE id = $1`
	row := r.db.QueryRow(query, id)
	var sub model.Subscription
	var endDate sql.NullTime
	err := row.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &endDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("get by id: scan error", "error", err)
		return nil, err
	}
	if endDate.Valid {
		sub.EndDate = &endDate.Time
	}
	return &sub, nil
}

func (r *SubscriptionRepository) Update(sub *model.Subscription) error {
	query := `UPDATE subscriptions SET service_name=$1, price=$2, user_id=$3, start_date=$4, end_date=$5
	          WHERE id=$6`
	res, err := r.db.Exec(query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, sub.ID)
	if err != nil {
		slog.Error("update: db error", "error", err)
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("subscription not found")
	}
	return nil
}

func (r *SubscriptionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id=$1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		slog.Error("delete: db error", "error", err)
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("subscription not found")
	}
	return nil
}

func (r *SubscriptionRepository) List(serviceName, userID *string, limit, offset int) ([]model.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
		WHERE ($1::TEXT IS NULL OR service_name = $1)
		  AND ($2::TEXT IS NULL OR user_id::TEXT = $2)
		ORDER BY start_date DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(query, serviceName, userID, limit, offset)
	if err != nil {
		slog.Error("list: query error", "error", err)
		return nil, err
	}
	defer rows.Close()

	var subscriptions []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		var endDate sql.NullTime
		err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &endDate)
		if err != nil {
			return nil, err
		}
		if endDate.Valid {
			sub.EndDate = &endDate.Time
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, nil
}

func (r *SubscriptionRepository) GetTotalCost(startPeriod, endPeriod time.Time, userID, serviceName *string) (int, error) {
	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE start_date <= $1
		  AND (end_date IS NULL OR end_date >= $2)
		  AND ($3::TEXT IS NULL OR user_id::TEXT = $3)
		  AND ($4::TEXT IS NULL OR service_name = $4)
	`
	var total int
	err := r.db.QueryRow(query, endPeriod, startPeriod, userID, serviceName).Scan(&total)
	if err != nil {
		slog.Error("total cost: query error", "error", err)
		return 0, err
	}
	return total, nil
}
