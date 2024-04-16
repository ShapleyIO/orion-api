package identity

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ShapleyIO/shapley.io-api/api/middleware"
	v1 "github.com/ShapleyIO/shapley.io-api/api/v1"
	"github.com/ShapleyIO/shapley.io-api/internal/passwordhasher"
	"github.com/redis/go-redis/v9"
)

type ServiceIdentity struct {
	redisClient *redis.Client
	hasher      passwordhasher.PasswordHasher
}

func NewServiceIdentity(redisClient *redis.Client, hasher passwordhasher.PasswordHasher) *ServiceIdentity {
	return &ServiceIdentity{
		redisClient: redisClient,
		hasher:      hasher,
	}
}

// Create a User
// (POST /v1/user)
func (s *ServiceIdentity) CreateUser(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Unmarshal the request body
	var user v1.User
	if err := json.Unmarshal(body, &user); err != nil {
		logger.Error().Err(err).Msg("failed to unmarshal request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the user already exists
	if s.redisClient.Exists(r.Context(), string(user.Email)).Val() == 1 {
		logger.Error().Str("email", string(user.Email)).Msg("user already exists")
		w.WriteHeader(http.StatusConflict)
	}

	// JSON Marshal the user
	userJson, err := json.Marshal(user)
	if err != nil {
		logger.Error().Err(err).Msg("failed to marshal user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create the user
	if err := s.redisClient.Set(r.Context(), string(user.Email), userJson, 0).Err(); err != nil {
		logger.Error().Err(err).Msg("failed to create user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Update a User's Password
// (PUT /v1/user/password/{user_id})
func (s *ServiceIdentity) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	logger := middleware.GetLogger(r.Context())

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Unmarshal the request body
	var login v1.Login
	if err := json.Unmarshal(body, &login); err != nil {
		logger.Error().Err(err).Msg("failed to unmarshal request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the user
	userJson, err := s.redisClient.Get(r.Context(), string(login.Email)).Result()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get user")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Unmarshal the user
	var user UserWithPassword
	if err := json.Unmarshal([]byte(userJson), &user); err != nil {
		logger.Error().Err(err).Msg("failed to unmarshal user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update the user's password
	user.Password = s.hasher.HashPassword(login.Password)

	// Marshal the user
	userJsonBytes, err := json.Marshal(user)
	if err != nil {
		logger.Error().Err(err).Msg("failed to marshal user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := s.redisClient.Set(r.Context(), string(login.Email), userJsonBytes, 0).Err(); err != nil {
		logger.Error().Err(err).Msg("failed to update user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete a User
// (DELETE /v1/user/{user_id})
func (s *ServiceIdentity) DeleteUser(w http.ResponseWriter, r *http.Request, params v1.DeleteUserParams) {
	logger := middleware.GetLogger(r.Context())

	// Delete user
	if err := s.redisClient.Del(r.Context(), string(params.Email)).Err(); err != nil {
		logger.Error().Err(err).Msg("failed to delete user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Get a User
// (GET /v1/user/{user_id})
func (s *ServiceIdentity) GetUser(w http.ResponseWriter, r *http.Request, params v1.GetUserParams) {
	logger := middleware.GetLogger(r.Context())

	// Get user
	userJson, err := s.redisClient.Get(r.Context(), string(params.Email)).Result()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get user")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Remove the password from the user
	var user UserWithPassword
	if err := json.Unmarshal([]byte(userJson), &user); err != nil {
		logger.Error().Err(err).Msg("failed to unmarshal user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Info().Str("email", string(params.Email)).Str("password", user.Password).Msg("user found")

	user.Password = ""
	userJsonBytes, err := json.Marshal(user)
	if err != nil {
		logger.Error().Err(err).Msg("failed to marshal user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(userJsonBytes)
}

// Update a User
// (PUT /v1/user/{user_id})
func (s *ServiceIdentity) UpdateUser(w http.ResponseWriter, r *http.Request, params v1.UpdateUserParams) {
	logger := middleware.GetLogger(r.Context())

	// Update User
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Unmarshal the request body
	var user v1.User
	if err := json.Unmarshal(body, &user); err != nil {
		logger.Error().Err(err).Msg("failed to unmarshal request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the user
	userJson, err := s.redisClient.Get(r.Context(), string(params.Email)).Result()
	if err != nil {
		logger.Error().Err(err).Msg("failed to get user")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Unmarshal the user
	var userWithPassword UserWithPassword
	if err := json.Unmarshal([]byte(userJson), &userWithPassword); err != nil {
		logger.Error().Err(err).Msg("failed to unmarshal user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update the user
	userWithPassword.FirstName = user.FirstName
	userWithPassword.LastName = user.LastName
	userWithPassword.Email = user.Email
	if err := s.redisClient.Set(r.Context(), string(params.Email), userWithPassword, 0).Err(); err != nil {
		logger.Error().Err(err).Msg("failed to update user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
