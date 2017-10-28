package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	bandit "github.com/alextanhongpin/go-bandit"
	"github.com/go-redis/redis"
	"github.com/rs/xid"
)

// Bandit represents the model of the bandit algorithm
type Bandit struct {
	Arm       int64     `json:"arm"`
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

// Job represents the cron job
type Job struct {
	sync.Mutex
	Cache  *redis.Client
	Bandit *bandit.EpsilonGreedy
	Key    string
}

// RecordsSince returns the keys that are available since this period - the seconds
func (j *Job) RecordsSince(seconds int) ([]string, error) {
	var out []string
	start := 0
	stop := time.Now().Add(-time.Duration(seconds)*time.Second).UTC().UnixNano() / 1e6
	cmd := j.Cache.ZRangeByScore(j.Key, redis.ZRangeBy{
		Min: fmt.Sprint(start),
		Max: fmt.Sprint(stop),
	})

	if cmd.Err() != nil {
		return out, cmd.Err()
	}
	out = cmd.Val()
	return out, nil
}

// GetPipeline is a redis pipeline responsible for getting all the values required for processing
func (j *Job) GetPipeline(armKey string) (keys []string, arm int, err error) {
	pipe := j.Cache.TxPipeline()
	kcmd := pipe.HKeys(armKey)
	if kcmd.Err() != nil {
		err = kcmd.Err()
		return
	}
	gcmd := pipe.HGet(armKey, j.Key)
	if gcmd.Err() != nil {
		err = gcmd.Err()
		return
	}
	if _, err = pipe.Exec(); err != nil {
		if err == redis.Nil {
			err = nil
			return
		}
		return
	}
	keys = kcmd.Val()
	armStr := gcmd.Val()

	// Convert arm str to int
	arm, err = strconv.Atoi(armStr)

	if err != nil {
		return
	}
	return
}

// DeletePipeline is a redis pipeline that removes all the stale keys
func (j *Job) DeletePipeline(armKey, key string, keys ...string) error {
	pipe := j.Cache.TxPipeline()
	if len(keys) != 0 {
		if err := pipe.HDel(armKey, keys...).Err(); err != nil {
			return err
		}
	}

	if err := pipe.ZRem("arm", key).Err(); err != nil {
		return err
	}
	if _, err := pipe.Exec(); err != nil && err != redis.Nil {
		return err
	}
	return nil
}
