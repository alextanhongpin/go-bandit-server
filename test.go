package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Bandit struct {
	Arm       int64     `json:"arm"`
	ArmID     string    `json:"arm_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	Type      string    `json:"type,omitempty"`
	Reward    float64   `json:"reward,omitempty"`
	Feature   string    `json:"feature"`
}
type MonteCarlo struct {
	Prob float64
}

func (m *MonteCarlo) Pull() int {
	if rand.Float64() > m.Prob {
		return 1
	}
	return 0
}

func main() {
	rand.Seed(time.Now().UnixNano())
	props := [...]float64{0.1, 0.1, 0.1, 0.1, 0.9}

	var wg sync.WaitGroup
	for i := 0; i < len(props); i++ {
		m := MonteCarlo{Prob: props[i]}
		for j := 0; j < 100; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				reward := m.Pull()
				if err := makeCall(reward); err != nil {
					log.Println(err)
				}
			}()
		}
	}
	wg.Wait()
	log.Println("done")
}

func makeCall(reward int) error {
	resp, err := http.Get("http://localhost:8080/select-arm")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var bandit Bandit
	json.NewDecoder(resp.Body).Decode(&bandit)

	bandit.Reward = float64(reward)
	buff, err := json.Marshal(bandit)
	resp, err = http.Post("http://localhost:8080/update-arm", "application/json", bytes.NewBuffer(buff))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println(string(res))
	return nil
}
