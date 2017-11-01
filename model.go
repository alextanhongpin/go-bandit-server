package main

import (
	"time"

	"github.com/rs/xid"
)

// Bandit represents the model of the bandit algorithm
type Bandit struct {
	Arm       int64     `json:"arm"` // Don't omit empty - 0 is a valid arm
	ArmID     string    `json:"arm_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	Type      string    `json:"type,omitempty"`
	Reward    float64   `json:"reward,omitempty"`
	Feature   string    `json:"feature"`
}

// NewBandit returns a new bandit
func NewBandit(arm int64) Bandit {
	guid := xid.New()
	return Bandit{
		Arm:       arm,
		ArmID:     guid.String(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Type:      "SELECT",
		Feature:   features[arm],
	}
}

// SelectArmRequest represents the request of the select arm endpoint
type SelectArmRequest struct {
}

// SelectArmResponse represents the response of the select arm endpoint
type SelectArmResponse struct {
	Bandit
}

// UpdateArmRequest represents the request of the update arm endpoint
type UpdateArmRequest struct {
	Bandit
}

// UpdateArmResponse represents the response of the update arm endpoint
type UpdateArmResponse struct {
	Ok bool `json:"ok"`
}
