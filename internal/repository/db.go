package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type DBStorage struct {
	db *sql.DB
	logger *zap.Logger
}

func (d *DBStorage) SetGauge(name string, value float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := d.db.ExecContext(ctx, `
		INSERT INTO metrics (name, metric_type, gauge_value)
		VALUES ($1, 'gauge', $2)
		ON CONFLICT (name, metric_type) 
		DO UPDATE 
		SET gauge_value = EXCLUDED.gauge_value, counter_value = NULL`, 
		name, value)
	if err != nil {
		return fmt.Errorf("failed to set gauge %q: %w", name, err)
	}
	
	return nil
}

func (d *DBStorage) GetGauge(name string) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var value float64

	gauge := d.db.QueryRowContext(ctx, `
		SELECT gauge_value FROM metrics
		WHERE name = $1
	`, name) 

	err := gauge.Scan(&value)

	if err == sql.ErrNoRows {
		d.logger.Sugar().Fatalf("metric %q not found", name)
		return 0, false
	}
	if err != nil {
		d.logger.Sugar().Fatalf("failed to get metric %q: %w", name, err)
		return 0, false
	}

	return value, true
}

func (d *DBStorage) GetAllGauges() (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := d.db.QueryContext(ctx, `
		SELECT name, gauge_value FROM metrics 
		WHERE metric_type = 'gauge'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)

	for rows.Next() {
		var name string
		var value float64
		err := rows.Scan(&name, &value)
		if err != nil {
			return nil, err
		}
		result[name] = value
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DBStorage) SetCounter(name string, value int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := d.db.ExecContext(ctx, `
		INSERT INTO metrics (name, metric_type, counter_value)
		VALUES ($1, 'counter', $2)
		ON CONFLICT (name, metric_type) 
		DO UPDATE 
		SET counter_value = metrics.counter_value + EXCLUDED.counter_value,
		gauge_value = NULL
	`, name, value)
	if err != nil {
		return fmt.Errorf("failed to set counter %q: %w", name, err)
	}
	
	return nil
}

func (d *DBStorage) GetCounter(name string) (int64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var value int64

	counter := d.db.QueryRowContext(ctx, `
		SELECT counter_value FROM metrics
		WHERE name = $1
	`, name) 

	err := counter.Scan(&value)

	if err == sql.ErrNoRows {
		d.logger.Sugar().Fatalf("metric %q not found", name)
		return 0, false
	}
	if err != nil {
		d.logger.Sugar().Fatalf("failed to get metric %q: %w", name, err)
		return 0, false
	}


	return value, true
}

func (d *DBStorage) GetAllCounters() (map[string]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := d.db.QueryContext(ctx, `
		SELECT name, counter_value FROM metrics 
		WHERE metric_type = 'counter'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int64)

	for rows.Next() {
		var name string
		var value int64
		err := rows.Scan(&name, &value)
		if err != nil {
			return nil, err
		}
		result[name] = value
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DBStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := d.db.PingContext(ctx)
	if err != nil {
		return err
	}

	return nil 
}