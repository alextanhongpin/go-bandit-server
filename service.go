package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/go-redis/redis"

	bandit "github.com/alextanhongpin/go-bandit"
)

type BanditService interface {
	SetCache(SetCacheRequest) error
	ClearCache(ClearCacheRequest) error
	Write(WriteRequest) error
	Read(ReadRequest) (*ReadResponse, error)
	ReadAll(ReadAllRequest) (*ReadAllResponse, error)
	ReadAndCompare(ReadAndCompareRequest) (*ReadAndCompareResponse, error)
	ExpireCache(ExpireCacheRequest) (*ExpireCacheResponse, error)
}

var (
	errInvalidID        = errors.New("The id does not exist or has been deleted")
	errInvalidArm       = errors.New("The value of the arm does not match")
	errRewardOutOfRange = errors.New("The value of the reward is out of range")
)

type banditService struct {
	Bandit *bandit.EpsilonGreedy
	Cache  *redis.Client
	DB     *badger.DB
}

func NewBanditService(bdt *bandit.EpsilonGreedy, cache *redis.Client, db *badger.DB) *banditService {
	return &banditService{
		Bandit: bdt,
		Cache:  cache,
		DB:     db,
	}
}

func (b *banditService) SetCache(req SetCacheRequest) error {
	cmd := b.Cache.ZAdd(req.Key, redis.Z{
		Score:  float64(req.Timestamp),
		Member: req.ID,
	})
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (b *banditService) ClearCache(req ClearCacheRequest) error {
	cmd := b.Cache.ZScan(req.Key, 0, req.ID, 1)
	if cmd.Err() != nil && cmd.Err() != redis.Nil {
		return cmd.Err()
	}

	vals, _ := cmd.Val()
	if len(vals) == 0 {
		return errInvalidID
	}
	val := vals[0]
	if err := b.Cache.ZRem(req.Key, val).Err(); err != nil && err != redis.Nil {
		return err
	}
	return nil
}

func (b *banditService) Write(req WriteRequest) error {
	err := b.DB.Update(func(txn *badger.Txn) error {
		val, err := json.Marshal(req.Val)
		if err != nil {
			return err
		}

		err = txn.Set([]byte(req.Key), []byte(val), 0)
		return err
	})
	return err
}

func (b *banditService) Read(req ReadRequest) (*ReadResponse, error) {
	var bandit Bandit
	err := b.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(req.Key))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}
		err = json.Unmarshal(val, &bandit)
		return err
	})

	if err != nil {
		return nil, err
	}
	return &ReadResponse{
		Bandit: bandit,
	}, nil
}

func (b *banditService) ReadAll(req ReadAllRequest) (*ReadAllResponse, error) {
	err := b.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.Value()
			if err != nil {
				return err
			}
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})
	return nil, err
}

func (b *banditService) ReadAndCompare(req ReadAndCompareRequest) (*ReadAndCompareResponse, error) {
	var bandit Bandit
	err := b.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(req.Key))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) == 0 {
			return errInvalidID
		}
		if err := json.Unmarshal(val, &bandit); err != nil {
			return err
		}

		if bandit.Arm != req.Val.Arm {
			return errInvalidArm
		}

		if req.Val.Reward > scoreMax || req.Val.Reward < scoreMin {
			return errRewardOutOfRange
		}
		bandit.UpdatedAt = time.Now()
		bandit.Type = "UPDATE"
		bandit.Reward = req.Val.Reward
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ReadAndCompareResponse{
		Bandit: bandit,
	}, nil
}

func (b *banditService) ExpireCache(req ExpireCacheRequest) (*ExpireCacheResponse, error) {
	var out []redis.Z

	start := 0
	stop := time.Now().Add(-time.Duration(req.Seconds)*time.Second).UTC().UnixNano() / 1e6

	cmd := b.Cache.ZRangeByScoreWithScores(req.Key, redis.ZRangeBy{
		Min: fmt.Sprint(start),
		Max: fmt.Sprint(stop),
	})

	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	out = cmd.Val()
	return &ExpireCacheResponse{
		Keys: out,
	}, nil
}

type SetCacheRequest struct {
	ID        string
	Key       string
	Timestamp int64
}

type ClearCacheRequest struct {
	Key string
	ID  string
}

type WriteRequest struct {
	Key string
	Val Bandit
}

type ReadRequest struct {
	Key string
}
type ReadResponse struct {
	Bandit Bandit
}

type ReadAllRequest struct{}
type ReadAllResponse struct{}

type ReadAndCompareRequest struct {
	Key string
	Val Bandit
}

type ReadAndCompareResponse struct {
	Bandit Bandit
}

type ExpireCacheRequest struct {
	Key     string
	Seconds int
}
type ExpireCacheResponse struct {
	Keys []redis.Z
}
