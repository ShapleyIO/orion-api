package main

import (
	"net/http"

	"github.com/ShapleyIO/orion-api/api/middleware"
	apiV1 "github.com/ShapleyIO/orion-api/api/v1"
	"github.com/ShapleyIO/orion-api/internal/config"
	"github.com/ShapleyIO/orion-api/internal/connect"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Replace with customized logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	cfg, err := config.NewConfig()
	if err != nil {
		log.Panic().Err(err).Msg("failed to load configuration")
	}

	baseRouter := chi.NewRouter()

	services, err := connect.CreateServices(cfg)
	if err != nil {
		log.Panic().Err(err).Msg("failed to create services")
	}

	// Setup Swagger Validator
	swaggerApi, err := apiV1.GetSwagger()
	if err != nil {
		log.Panic().Err(err).Msg("failed to get swagger for api")
	}

	middlewareOptions := &chi_middleware.Options{
		SilenceServersWarning: true,
	}

	// Setup Middlewares
	baseRouter.Use(chi_middleware.OapiRequestValidatorWithOptions(swaggerApi, middlewareOptions))
	baseRouter.Use(middleware.Logger)

	apiV1.HandlerFromMux(services.Handlers(), baseRouter)

	// router.NotFound()
	// router.MethodNotAllowed()

	log.Info().Int("port", 8080).Msg("starting server")
	err = http.ListenAndServe(":8080", baseRouter)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start http server")
	}
}
