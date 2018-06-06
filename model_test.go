package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateAndUpdateArm(t *testing.T) {
	assert := assert.New(t)

	m, err := NewEpsilonModel(3)
	assert.Nil(err)

	arm := m.SelectArm(0.1)
	assert.Equal(false, arm.IsActionTaken, "should create an arm without action taken")
	assert.Equal(false, arm.IsCompleted, "should create an arm that is not yet completed")
	assert.Equal(0.0, arm.Reward, "should create an arm without reward")

	counts, rewards := m.Info()
	assert.Equal(make([]int, 3), counts, "counts should be initialized to zero values")
	assert.Equal(make([]float64, 3), rewards, "rewards should be initialized to zero values")

	err = m.Create(*arm)
	assert.Nil(err)

	arm.Reward = 1.0
	err = m.Update(*arm)
	assert.Nil(err)

	counts, rewards = m.Info()
	assert.Equal(1, counts[arm.Arm], "counts should be correct for arm %d", arm.Arm)
	assert.Equal(1.0, rewards[arm.Arm], "rewards should be correct for arm %d", arm.Arm)

	arms := m.GetArms()
	assert.Equal(true, arms[0].IsActionTaken, "should flag arm as action taken")
	assert.Equal(true, arms[0].IsCompleted, "should flag arm as completed")
}

func TestCreateAndSweepArm_SweepComplete(t *testing.T) {
	assert := assert.New(t)

	m, err := NewEpsilonModel(3)
	assert.Nil(err)

	// Select an arm and set the creation time to 1 minute ago - to test for expiry
	arm := m.SelectArm(0.1)
	arm.CreatedAt = time.Now().Add(time.Duration(-1 * time.Minute))

	// Create the arm, and perform a sweep
	err = m.Create(*arm)
	assert.Nil(err)

	err = m.Sweep(time.Duration(1 * time.Minute))
	assert.Nil(err)

	// Compare results
	counts, rewards := m.Info()
	assert.Equal(1, counts[arm.Arm], "counts should be correct for arm %d", arm.Arm)
	assert.Equal(0.0, rewards[arm.Arm], "rewards should be correct for arm %d", arm.Arm)

	arms := m.GetArms()
	assert.Equal(true, arms[0].IsActionTaken, "should flag arm as action taken")
	assert.Equal(true, arms[0].IsCompleted, "should flag arm as completed")
}

func TestCreateAndSweepArm_NoSweep(t *testing.T) {
	assert := assert.New(t)

	m, err := NewEpsilonModel(3)
	assert.Nil(err)

	// Create the arm, and perform a sweep
	arm := m.SelectArm(0.1)
	err = m.Create(*arm)
	assert.Nil(err)

	err = m.Sweep(time.Duration(1 * time.Minute))
	assert.Nil(err)

	// Compare results
	counts, rewards := m.Info()
	assert.Equal(0, counts[arm.Arm], "counts should be correct for arm %d", arm.Arm)
	assert.Equal(0.0, rewards[arm.Arm], "rewards should be correct for arm %d", arm.Arm)

	arms := m.GetArms()
	assert.Equal(false, arms[0].IsActionTaken, "should flag arm as action taken")
	assert.Equal(false, arms[0].IsCompleted, "should flag arm as completed")
}
