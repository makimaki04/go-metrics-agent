package repository

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

type Repository interface {
	SetGauge(name string, value float64)
	SetCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
}

func NewStorage() Repository {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemStorage) SetGauge(name string, value float64) {
	m.gauges[name] = value
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	value, ok := m.gauges[name]
	return value, ok
}

func (m *MemStorage) SetCounter(name string, value int64) {
	m.counters[name] += value
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	value, ok := m.counters[name]
	return value, ok
}

