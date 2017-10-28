package main

import (
	"time"
)

// Bandit represents the model of the bandit algorithm
type Bandit struct {
	Arm       int64
	ArmID     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Type      string // Select or Update
	Reward    float64
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
	Ok bool
}

// Stats represents the time series of the data
type Stats struct {
	Stats []Stat
}

// Stat represent the individual stat
type Stat struct {
	Type        string // Select or Update
	ElapsedTime time.Duration
	Reward      float64
	Arm         int64
	ArmID       string
	// Rewards []float64
	// Counts []int64
	// Features []string
	// N int64
	// Epsilon float64

}
