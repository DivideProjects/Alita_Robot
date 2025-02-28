package cache

import (
	"context"
	"github.com/divideprojects/Alita_Robot/alita/config"
	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	redisstore "github.com/eko/gocache/store/redis/v4"
	ristrettostore "github.com/eko/gocache/store/ristretto/v4"
	"github.com/redis/go-redis/v9"
)

var (
	Context = context.Background()
	Marshal *marshaler.Marshaler
)

type AdminCache struct {
	ChatId   int64
	UserInfo []gotgbot.MergedChatMember
	Cached   bool
}

// InitCache initializes the cache.
func InitCache() {
	opt, err := redis.ParseURL(config.RedisURI)
	if err != nil {
		log.Fatalf("failed to parse redis url: %v", err)
	}

	redisClient := redis.NewClient(opt)
	if err = redisClient.Ping(Context).Err(); err != nil {
		log.Fatalf("failed to ping redis: %v", err)
	}

	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{NumCounters: 1000, MaxCost: 100, BufferItems: 64})
	if err != nil {
		log.Fatalf("failed to create ristretto cache: %v", err)
	}

	cacheManager := cache.NewChain[any](
		cache.New[any](ristrettostore.NewRistretto(ristrettoCache)),
		cache.New[any](redisstore.NewRedis(redisClient)),
	)

	// Initializes marshaler
	Marshal = marshaler.New(cacheManager)
}
