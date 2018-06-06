package main

import (
	"sync"
	"time"

	bandit "github.com/alextanhongpin/go-bandit"
)

// Model represents the business logic for bandit algorithm
type Model interface {
	GetArms() []Arm
	Sweep(elapsed time.Duration) error
	Update(arm Arm) error
	SelectArm(probability float64) *Arm
	Create(arm Arm) error
	Info() (counts []int, rewards []float64)
}

type model struct {
	sync.RWMutex
	bandit bandit.Bandit
	store  Store
}

// NewModel returns a new model
func NewModel(b bandit.Bandit, s Store) Model {
	return &model{
		bandit: b,
		store:  s,
	}
}

// NewEpsilonModel will return a new model configured with the epsilon greedy algorithm
func NewEpsilonModel(arms int) (Model, error) {
	b, err := NewDefaultEpsilonGreedy()
	if err != nil {
		return nil, err
	}
	if err := b.Init(arms); err != nil {
		return nil, err
	}

	return NewModel(b, NewMemStore()), nil
}

func (m *model) GetArms() []Arm {
	m.RLock()
	defer m.RUnlock()
	return m.store.GetArms()
}

// Sweep will perform cleanup of the arms that are no longer actionable
func (m *model) Sweep(elapsed time.Duration) error {
	m.Lock()
	defer m.Unlock()

	res, err := m.store.List(elapsed)
	if err != nil {
		return err
	}

	for _, r := range res {
		if err := m.store.Update(r); err != nil {
			return err
		}

		if err := m.bandit.Update(r.Arm, r.Reward); err != nil {
			return err
		}
	}
	return nil
}

// Update will update the existing bandit
func (m *model) Update(arm Arm) error {
	m.Lock()
	defer m.Unlock()

	// Check if arm exists - only update arms that was created by us
	if _, err := m.store.FindID(arm.ID); err != nil {
		return err
	}

	// Update the request
	if err := m.store.Update(arm); err != nil {
		return err
	}

	if err := m.bandit.Update(arm.Arm, arm.Reward); err != nil {
		return err
	}

	return nil
}

// SelectArm will return a new arm
func (m *model) SelectArm(probability float64) *Arm {
	m.RLock()
	defer m.RUnlock()

	chosenArm := m.bandit.SelectArm(probability)
	return NewArm(chosenArm)
}

// Create will create a new arm
func (m *model) Create(a Arm) error {
	m.Lock()
	defer m.Unlock()

	return m.store.Create(a)
}

// Info will return the stats of the counts and rewards
func (m *model) Info() (counts []int, rewards []float64) {
	m.RLock()
	defer m.RUnlock()

	return m.bandit.GetCounts(), m.bandit.GetRewards()
}
