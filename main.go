package main

import (
	"log"
	"net/http"

	"github.com/alextanhongpin/go-bandit"
)

// Server represents the server config
type Server struct {
	Bandit *bandit.EpsilonGreedy
}

// Config
const (
	port    = ":8080"
	nArms   = 3
	epsilon = 0.1
)

func main() {

	eps := bandit.NewEpsilonGreedy(nArms, epsilon)

	server := &Server{
		Bandit: eps,
	}

	// Load from db?
	// eps.SetRewards(exp.Rewards)
	// eps.SetCounts(exp.Counts)
	// eps.Update(int(msg.Arm), float64(msg.Reward))
	// s.Bandit.SelectArm()

	mux := http.NewServeMux()
	mux.Handle("/select-arm", selectArm(server))
	mux.Handle("/update-arm", updateArm(server))
	// Endpoint to get the stats of the current experiment
	// mux.Handle("/stats", nil)

	// Run cron periodically to update those assumptions to false
	log.Printf("listening to port *%s. press ctrl + c to cancel", port)
	http.ListenAndServe(port, mux)
}
