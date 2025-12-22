package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	models "github.com/makimaki04/go-metrics-agent.git/internal/model"
	"go.uber.org/zap"
)

// DBStorage - struct for the database storage
type DBStorage struct {
	db     *sql.DB
	logger *zap.Logger
}

// Constants for the database storage
const (
	insertGaugeQuery = `
		INSERT INTO metrics (name, metric_type, gauge_value)
		VALUES ($1, 'gauge', $2)
		ON CONFLICT (name, metric_type) 
		DO UPDATE 
		SET gauge_value = EXCLUDED.gauge_value, counter_value = NULL
	`

	getGaugeQuery = `
		SELECT gauge_value FROM metrics
		WHERE name = $1 AND metric_type = 'gauge'
	`

	getAllGaugesQuery = `
		SELECT name, gauge_value FROM metrics 
		WHERE metric_type = 'gauge'
	`

	insertCounterQuery = `
		INSERT INTO metrics (name, metric_type, counter_value)
		VALUES ($1, 'counter', $2)
		ON CONFLICT (name, metric_type) 
		DO UPDATE 
		SET counter_value = metrics.counter_value + EXCLUDED.counter_value,
		    gauge_value = NULL
	`

	getCounterQuery = `
		SELECT counter_value FROM metrics
		WHERE name = $1 AND metric_type = 'counter'
	`

	getAllCountersQuery = `
		SELECT name, counter_value FROM metrics 
		WHERE metric_type = 'counter'
	`
)

// SetGauge - method for setting a gauge
// set the value of the gauge
// if error, return error
// if success, return nil
func (d *DBStorage) SetGauge(name string, value float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := d.db.ExecContext(ctx, insertGaugeQuery,
		name, value)
	if err != nil {
		return fmt.Errorf("failed to set gauge %q: %w", name, err)
	}

	return nil
}

// GetGauge - method for getting a gauge
// get the value of the gauge
// if error, return error
// if success, return the value of the gauge
func (d *DBStorage) GetGauge(name string) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var value float64

	gauge := d.db.QueryRowContext(ctx, getGaugeQuery, name)

	err := gauge.Scan(&value)
	if err != nil {
		d.logger.Info("failed to get metric",
			zap.String("name", name),
			zap.Error(err),
		)
		return 0, false
	}

	return value, true
}

// GetAllGauges - method for getting all gauges
// get all the gauges
// if error, return error
// if success, return the value of the gauges
func (d *DBStorage) GetAllGauges() (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := d.db.QueryContext(ctx, getAllGaugesQuery)
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

// SetCounter - method for setting a counter
// set the value of the counter
// if error, return error
// if success, return nil
func (d *DBStorage) SetCounter(name string, value int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := d.db.ExecContext(ctx, insertCounterQuery, name, value)
	if err != nil {
		return fmt.Errorf("failed to set counter %q: %w", name, err)
	}

	return nil
}

// GetCounter - method for getting a counter
// get the value of the counter
// if error, return false
// if success, return the value of the counter and true
func (d *DBStorage) GetCounter(name string) (int64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var value int64

	counter := d.db.QueryRowContext(ctx, getCounterQuery, name)

	err := counter.Scan(&value)
	if err != nil {
		d.logger.Info("failed to get metric",
			zap.String("name", name),
			zap.Error(err),
		)
		return 0, false
	}

	return value, true
}

// GetAllCounters - method for getting all counters
// get all the counters
// if error, return error
// if success, return the value of the counters
func (d *DBStorage) GetAllCounters() (map[string]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := d.db.QueryContext(ctx, getAllCountersQuery)
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

// SetMetricBatch - method for setting a batch of metrics
// set the value of the metrics
// if error, return error
// if success, return nil
func (d *DBStorage) SetMetricBatch(metrics []models.Metrics) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmtGauge, err := tx.PrepareContext(ctx, insertGaugeQuery)
	if err != nil {
		return err
	}
	defer stmtGauge.Close()

	stmtCounter, err := tx.PrepareContext(ctx, insertCounterQuery)
	if err != nil {
		return err
	}
	defer stmtCounter.Close()

	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			if m.Value == nil {
				return fmt.Errorf("gauge %s has no value", m.ID)
			}
			if _, err := stmtGauge.ExecContext(ctx, m.ID, m.Value); err != nil {
				return fmt.Errorf("failed to insert gauge %s: %w", m.ID, err)
			}
		case "counter":
			if m.Delta == nil {
				return fmt.Errorf("counter %s has no delta", m.ID)
			}
			if _, err := stmtCounter.ExecContext(ctx, m.ID, m.Delta); err != nil {
				return fmt.Errorf("failed to insert counter %s: %w", m.ID, err)
			}
		}
	}

	return tx.Commit()
}

// Ping - method for pinging the database
func (d *DBStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := d.db.PingContext(ctx)
	if err != nil {
		return err
	}

	return nil
}
