package health

import (
	"net/http"

	"github.com/ShapleyIO/orion-api/api/middleware"
	"github.com/redis/go-redis/v9"
)

type ServiceHealth struct {
	redisClient *redis.Client
}

func NewServiceHealth(redisClient *redis.Client) *ServiceHealth {
	return &ServiceHealth{
		redisClient: redisClient,
	}
}

// HealthAlive
// (GET /v1/health/alive)
func (s *ServiceHealth) HealthAlive(w http.ResponseWriter, r *http.Request) {
	// Check if the Redis client is connected
	if _, err := s.redisClient.Ping(r.Context()).Result(); err != nil {
		logger := middleware.GetLogger(r.Context())
		logger.Debug().Err(err).Msg("failed to ping Redis")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.Write([]byte("{\"status\":\"alive\"}"))
}

// HealthReady
// (GET /v1/health/ready)
func (s *ServiceHealth) HealthReady(w http.ResponseWriter, r *http.Request) {
	// Check if the Redis client is connected
	if _, err := s.redisClient.Ping(r.Context()).Result(); err != nil {
		logger := middleware.GetLogger(r.Context())
		logger.Debug().Err(err).Msg("failed to ping Redis")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.Write([]byte("{\"status\":\"ready\"}"))
}
