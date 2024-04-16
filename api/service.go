package api

import (
	"fmt"

	"github.com/ShapleyIO/shapley.io-api/api/handlers/authn"
	"github.com/ShapleyIO/shapley.io-api/api/handlers/health"
	"github.com/ShapleyIO/shapley.io-api/api/handlers/identity"
	v1 "github.com/ShapleyIO/shapley.io-api/api/v1"
	"github.com/ShapleyIO/shapley.io-api/internal/config"
	"github.com/redis/go-redis/v9"
)

type Handlers struct {
	*identity.ServiceIdentity
	*authn.ServiceAuthN
	*health.ServiceHealth
}

var _ v1.ServerInterface = (*Handlers)(nil)

func NewHandlers(cfg *config.Config) (*Handlers, error) {
	handlers := new(Handlers)

	// Create a Redis Client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: "",           // no password set
		DB:       cfg.Redis.DB, // use default DB
	})

	handlers.ServiceIdentity = identity.NewService(redisClient)
	handlers.ServiceHealth = health.NewServiceHealth(redisClient)

	return handlers, nil
}
