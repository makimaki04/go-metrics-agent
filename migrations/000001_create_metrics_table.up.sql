-- Создание таблицы метрик
CREATE TABLE metrics (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    metric_type VARCHAR(10) NOT NULL CHECK (metric_type IN ('gauge', 'counter')),
    gauge_value DOUBLE PRECISION NULL,
    counter_value INTEGER NULL,
    timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT value_check CHECK (
        (metric_type = 'gauge' AND gauge_value IS NOT NULL AND counter_value IS NULL) OR
        (metric_type = 'counter' AND xcounter_value IS NOT NULL AND gauge_value IS NULL)
    )
);

CREATE INDEX idx_metrics_name ON metrics(name);