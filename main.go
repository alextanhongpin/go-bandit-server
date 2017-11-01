package main

import (
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/alextanhongpin/go-bandit"
	"github.com/dgraph-io/badger"
	"github.com/go-redis/redis"
	"github.com/robfig/cron"
)

// Config
const (
	nArms    = 3
	epsilon  = 0.1
	scoreMin = 0.0
	scoreMax = 1.0

	schedule = "*/5 * * * *"
	port     = ":8080"

	redisKey  = "arm"
	redisAddr = "localhost:6379"
	redisPass = ""
	redisDB   = 0

	dbDir = "./tmp/badger"
)

var features = [...]string{"feat-1", "feat-2", "feat-3"}

func main() {
	rand.Seed(time.Now().UnixNano())

	if len(features) != nArms {
		log.Fatal("Number of features to be tested is incorrect")
	}
	cache := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       redisDB,
	})
	// Badger
	db := newDB(dbDir)
	defer db.Close()
	// TODO: allow selection of different bandit algorithm
	// e.g. if "greedy" then (...) else if "ucb1" then (...)
	epsBandit := bandit.NewEpsilonGreedy(nArms, epsilon)

	banditsvc := NewBanditService(epsBandit, cache, db)
	endpoint := Endpoint{
		Service: banditsvc,
	}

	mux := http.NewServeMux()
	mux.Handle("/select-arm", endpoint.SelectArm())
	mux.Handle("/update-arm", endpoint.UpdateArm())
	mux.Handle("/stats", endpoint.GetStats())
	mux.Handle("/logs", endpoint.GetLogs())

	c := cron.New()
	c.AddFunc(schedule, func() {
		resp, err := banditsvc.ExpireCache(ExpireCacheRequest{
			Key:     redisKey,
			Seconds: 10,
		})
		if err != nil {
			log.Println("error pulling redis record:", err.Error())
			return
		}
		keys := resp.Keys
		messages := make(chan int)
		go func() {
			defer close(messages)
			for i := 0; i < len(keys); i++ {
				key := keys[i]
				if err := banditsvc.ClearCache(ClearCacheRequest{
					Key: redisKey,
					ID:  key.Member.(string),
				}); err != nil {
					log.Println(err)
					return
				}

				resp, err := banditsvc.Read(ReadRequest{
					Key: key.Member.(string),
				})
				if err != nil {
					log.Println(err)
					return
				}

				messages <- int(resp.Bandit.Arm)
			}
		}()

		for msg := range messages {
			epsBandit.Update(msg, 0)
		}
	})
	c.Start()

	go func() {
		for {
			log.Println("Backends ", runtime.NumGoroutine())
			time.Sleep(time.Second)
		}
	}()
	log.Printf("listening to port *%s. press ctrl + c to cancel", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func newDB(dir string) *badger.DB {
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
