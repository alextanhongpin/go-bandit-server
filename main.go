package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/alextanhongpin/go-bandit"
	"github.com/go-redis/redis"
	"github.com/robfig/cron"
)

// Server represents the server config
type Server struct {
	sync.Mutex
	Bandit *bandit.EpsilonGreedy
	Cache  *redis.Client
}

// Config
const (
	port     = ":8080"
	nArms    = 3
	epsilon  = 0.1
	scoreMin = 0.0
	scoreMax = 1.0
	schedule = "*/5 * * * *"
	redisKey = "arm"
)

var features = [...]string{"feat-1", "feat-2", "feat-3"}

func main() {
	rand.Seed(time.Now().UnixNano())

	if len(features) != nArms {
		log.Fatal("Number of features to be tested is incorrect")
	}
	cache := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	epsBandit := bandit.NewEpsilonGreedy(nArms, epsilon)
	server := &Server{
		Bandit: epsBandit,
		Cache:  cache,
	}

	mux := http.NewServeMux()
	mux.Handle("/select-arm", selectArm(server))
	mux.Handle("/update-arm", updateArm(server))
	mux.Handle("/stats", getStats(server))

	// Run cron periodically to update those rewards to 0
	j := Job{
		Key:   redisKey,
		Cache: cache,
	}
	c := cron.New()
	c.AddFunc(schedule, func() {
		j.Lock()
		keys, err := j.RecordsSince(10) // Seconds
		j.Unlock()
		if err != nil {
			log.Println("error pulling redis record:", err.Error())
			return
		}

		for i := 0; i < len(keys); i++ {
			func(index int) {
				key := keys[index]
				armKey := fmt.Sprintf("%s:%s", redisKey, key)
				armKeys, arm, err := j.GetPipeline(armKey)
				if err != nil {
					log.Println("getPipelineError:", err.Error())
					return
				}

				if err := j.DeletePipeline(armKey, key, armKeys...); err != nil {
					log.Println("deletePipelineError:", err.Error())
					return
				}
				j.Lock()
				epsBandit.Update(arm, 0)
				j.Unlock()
			}(i)
		}
	})
	c.Start()

	log.Printf("listening to port *%s. press ctrl + c to cancel", port)
	http.ListenAndServe(port, mux)
}
