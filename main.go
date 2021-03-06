package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/robfig/cron"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	log.Println("app version:", os.Getenv("VCS_REF"))
	experimentName := "colors"
	features := []string{"red", "green", "blue"}
	everyMinute := "0 * * * * *"
	sweepDuration := time.Duration(time.Minute * 1)

	m, err := NewEpsilonModel(len(features), features)
	if err != nil {
		log.Fatal(err)
	}

	c := cron.New()
	c.AddFunc(everyMinute, func() {
		log.Println("running cron")
		if err := m.Sweep(sweepDuration); err != nil {
			log.Println(err)
		}
	})
	c.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("/arms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			arm := m.SelectArm(rand.Float64())
			if err := m.Create(*arm); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ResponseError{
					Error: err.Error(),
					Code:  http.StatusBadRequest,
				})
				return
			}

			json.NewEncoder(w).Encode(arm)
			return
		}
		if r.Method == http.MethodPost {
			var arm Arm
			if err := json.NewDecoder(r.Body).Decode(&arm); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ResponseError{
					Error: err.Error(),
					Code:  http.StatusBadRequest,
				})
				return
			}
			if err := m.Update(arm); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ResponseError{
					Error: err.Error(),
					Code:  http.StatusBadRequest,
				})
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			fmt.Fprint(w, `{"ok": true}`)
		}
	})

	mux.HandleFunc("/arms/stats", func(w http.ResponseWriter, r *http.Request) {
		counts, rewards := m.Info()
		json.NewEncoder(w).Encode(GetStatsResponse{
			Counts:         counts,
			Rewards:        rewards,
			Features:       features,
			ExperimentName: experimentName,
		})
	})

	srv := &http.Server{
		Addr:         ":9090",
		Handler:      mux,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	log.Printf("listening to port *:9090. press ctrl + c to cancel.")
	log.Fatal(srv.ListenAndServe())
}
