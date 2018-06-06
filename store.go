package main

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrNotFound represents the error message when arm is not found
	ErrNotFound = errors.New("arm not found")
)

// Store represents the interface to the storage layer
type Store interface {
	GetArms() []Arm
	List(elapsed time.Duration) ([]Arm, error)
	FindID(id string) (*Arm, error)
	Update(b Arm) error
	Create(b Arm) error
}

type memStore struct {
	sync.RWMutex
	Arms []Arm
}

// NewMemStore returns a new in-memory store
func NewMemStore() Store {
	return &memStore{}
}

func (m *memStore) GetArms() []Arm {
	m.RLock()
	defer m.RUnlock()
	return m.Arms
}

// List will return arms that are not yet completed
func (m *memStore) List(elapsed time.Duration) ([]Arm, error) {
	m.RLock()
	defer m.RUnlock()

	var arms []Arm
	for _, arm := range m.Arms {
		if !arm.IsCompleted && time.Since(arm.CreatedAt) >= elapsed {
			arms = append(arms, arm)
		}
	}
	return arms, nil
}

// FindID will find an existing arm by id
func (m *memStore) FindID(id string) (*Arm, error) {
	m.RLock()
	defer m.RUnlock()

	for _, arm := range m.Arms {
		if arm.ID == id && !arm.IsCompleted {
			return &arm, nil
		}
	}
	return nil, ErrNotFound
}

// Update will update an existing bandit
func (m *memStore) Update(a Arm) error {
	m.Lock()
	defer m.Unlock()

	a.IsCompleted = true
	a.IsActionTaken = true
	a.UpdatedAt = NewUTCDate()

	for i, bandit := range m.Arms {
		if bandit.ID == a.ID {
			m.Arms[i] = a
		}
	}

	return nil
}

// Create will create a new bandit
func (m *memStore) Create(a Arm) error {
	m.Lock()
	defer m.Unlock()

	m.Arms = append(m.Arms, a)
	return nil
}
