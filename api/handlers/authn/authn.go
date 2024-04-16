package authn

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ShapleyIO/shapley.io-api/api/handlers/identity"
	"github.com/ShapleyIO/shapley.io-api/api/middleware"
	v1 "github.com/ShapleyIO/shapley.io-api/api/v1"
	"github.com/ShapleyIO/shapley.io-api/internal/passwordhasher"
	"github.com/redis/go-redis/v9"
)

type ServiceAuthN struct {
	redisClient *redis.Client
	hasher      passwordhasher.PasswordHasher
}

func NewServiceAuthN(redisClient *redis.Client, hasher passwordhasher.PasswordHasher) *ServiceAuthN {
	return &ServiceAuthN{
		redisClient: redisClient,
		hasher:      hasher,
	}
}

// Login a User
// (POST /v1/login)
func (s *ServiceAuthN) Login(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())

	defer r.Body.Close()

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read request body")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Unmarshal the request body
	var login v1.Login
	if err := json.Unmarshal(body, &login); err != nil {
		logger.Error().Err(err).Msg("failed to unmarshal request body")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get the user from Redis
	userJsonBytes, err := s.redisClient.Get(r.Context(), string(login.Email)).Result()
	if err != nil {
		if err == redis.Nil {
			logger.Error().Str("email", string(login.Email)).Msg("user not found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		logger.Error().Err(err).Msg("failed to get user from Redis")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Unmarshal the user
	var user identity.UserWithPassword
	if err := json.Unmarshal([]byte(userJsonBytes), &user); err != nil {
		logger.Error().Err(err).Msg("failed to unmarshal user")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Compare the password
	valid, err := s.hasher.Compare(login.Password, user.Password)
	if err != nil {
		logger.Error().Err(err).Msg("failed to compare passwords")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !valid {
		logger.Error().Msg("invalid password")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Redirect to SSO
	http.Redirect(w, r, "https://sso.shapley.io/confirm", http.StatusFound)
}

// Logout a User
// (POST /v1/logout)
func (s *ServiceAuthN) Logout(w http.ResponseWriter, r *http.Request) {
}
