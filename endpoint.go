package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Endpoint struct {
	sync.RWMutex
	Service *banditService
}

func (e *Endpoint) SelectArm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			arm := e.Service.Bandit.SelectArm()
			bandit := NewBandit(int64(arm))

			if err := e.Service.SetCache(SetCacheRequest{
				ID:        bandit.ArmID,
				Key:       redisKey,
				Timestamp: bandit.CreatedAt.UnixNano() / 1e6,
			}); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := e.Service.Write(WriteRequest{
				Key: bandit.ArmID,
				Val: bandit,
			}); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := json.NewEncoder(w).Encode(SelectArmResponse{bandit}); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "Method not implemented", http.StatusNotImplemented)
		}
	}
}

func (e *Endpoint) UpdateArm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var bandit Bandit
			if err := json.NewDecoder(r.Body).Decode(&bandit); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			resp, err := e.Service.ReadAndCompare(ReadAndCompareRequest{
				Key: bandit.ArmID,
				Val: bandit,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			bandit = resp.Bandit

			log.Printf("got bandit %#v\n", bandit)

			if err := e.Service.ClearCache(ClearCacheRequest{
				Key: redisKey,
				ID:  bandit.ArmID,
			}); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := e.Service.Write(WriteRequest{
				Key: bandit.ArmID,
				Val: bandit,
			}); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			e.Service.Bandit.Update(int(bandit.Arm), bandit.Reward)
			log.Println("bandit afer", e.Service.Bandit)

			if err := json.NewEncoder(w).Encode(UpdateArmResponse{true}); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "Method not implemented", http.StatusNotImplemented)
		}
	}
}

func (e *Endpoint) GetLogs() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		_, err := e.Service.ReadAll(ReadAllRequest{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "done")
	}

}
func (e *Endpoint) GetStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e.RLock()
		bandit := e.Service.Bandit
		e.RUnlock()

		if err := json.NewEncoder(w).Encode(bandit); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
