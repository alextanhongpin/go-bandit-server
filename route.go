package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func selectArm(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		arm := s.Bandit.SelectArm()

		res := SelectArmResponse{
			Bandit: Bandit{
				Arm:       arm,
				ArmID:     "1",
				CreatedAt: time.Now().UTC().Format(time.RFC3339),
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
				Type:      "SELECT",
			},
		}

		// Store a copy to redis and also in memory

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func updateArm(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bandit := Bandit{}
		if err := json.NewDecoder(r.Body).Decode(&bandit); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.Bandit.Update(bandit.Arm, bandit.Reward)

		// Update a copy to redis and also in memory
		res := UpdateArmResponse{
			Ok: true,
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
