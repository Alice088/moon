package entity

import (
	"sync"
	"time"
)

type Metrics struct {
	mu        sync.RWMutex
	data      map[string]any
	timestamp time.Time
}

func NewMetrics(name string) *Metrics {
	return &Metrics{
		data:      make(map[string]any),
		timestamp: time.Now(),
	}
}

func (m *Metrics) Set(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *Metrics) Get(key string) any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[key]
}

func (m *Metrics) GetAll() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copy := make(map[string]any)
	for k, v := range m.data {
		copy[k] = v
	}
	return copy
}
