package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

func selectArm(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.Lock()
			arm := s.Bandit.SelectArm()
			s.Unlock()

			bandit := NewBandit(int64(arm))

			res := SelectArmResponse{
				Bandit: bandit,
			}

			hash := make(map[string]interface{})
			buff, err := json.Marshal(res.Bandit)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := json.Unmarshal(buff, &hash); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			zcmd := s.Cache.ZAdd(redisKey, redis.Z{
				Score:  float64(bandit.CreatedAt.UnixNano() / 1e6),
				Member: bandit.ArmID,
			})

			if zcmd.Err() != nil {
				http.Error(w, zcmd.Err().Error(), http.StatusBadRequest)
				return
			}

			cmd := s.Cache.HMSet(fmt.Sprintf("%s:%s", redisKey, bandit.ArmID), hash)
			if cmd.Err() != nil {
				http.Error(w, cmd.Err().Error(), http.StatusBadRequest)
				return
			}

			// Store a copy to redis and also in memory

			if err := json.NewEncoder(w).Encode(res); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "Method not implemented", http.StatusNotImplemented)
		}
	}
}

func updateArm(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var bandit Bandit
			if err := json.NewDecoder(r.Body).Decode(&bandit); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// Validate that they are the same first
			keyStr := fmt.Sprintf("%s:%s", redisKey, bandit.ArmID)
			cmd := s.Cache.HGetAll(keyStr)
			if cmd.Err() != nil {
				http.Error(w, cmd.Err().Error(), http.StatusBadRequest)
				return
			}
			hash := cmd.Val()
			if len(hash) == 0 {
				http.Error(w, "The field arm_id is missing", http.StatusBadRequest)
				return
			}

			if hash["arm"] != fmt.Sprint(bandit.Arm) {
				http.Error(w, "The field arm does not match", http.StatusBadRequest)
				return
			}

			// Get fields
			keys := make([]string, len(hash))
			for key := range hash {
				keys = append(keys, key)
			}

			// Delete hash
			delCmd := s.Cache.HDel(keyStr, keys...)
			if delCmd.Err() != nil {
				http.Error(w, delCmd.Err().Error(), http.StatusBadRequest)
				return
			}

			createdAt, err := time.Parse(time.RFC3339, hash["created_at"])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			elapsed := time.Since(createdAt)
			log.Println("time elapsed", elapsed)

			if bandit.Reward > scoreMax || bandit.Reward < scoreMin {
				http.Error(w, "The reward provided is out of range", http.StatusBadRequest)
				return
			}

			s.Lock()
			s.Bandit.Update(int(bandit.Arm), bandit.Reward)
			s.Unlock()

			// Update a copy to redis and also in memory
			res := UpdateArmResponse{
				Ok: true,
			}

			if err := json.NewEncoder(w).Encode(res); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "Method not implemented", http.StatusNotImplemented)
		}
	}
}

func getStats(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Lock()
		bandit := s.Bandit
		s.Unlock()

		if err := json.NewEncoder(w).Encode(bandit); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
