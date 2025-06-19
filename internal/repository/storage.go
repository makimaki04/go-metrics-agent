package repository

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

type Storage interface {
	SetGauge(name string, value float64)
	SetCounter(name string, value int64)
}

func NewStorage() Storage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemStorage) SetGauge(name string, value float64) {
	m.gauges[name] = value
}

func (m *MemStorage) SetCounter(name string, value int64) {
	m.counters[name] += value
}
