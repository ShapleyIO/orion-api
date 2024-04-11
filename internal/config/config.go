package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	Database *DatabaseConfig
	Redis    *RedisConfig
}

type DatabaseConfig struct {
	Host string
	Port int
	User string
}

type RedisConfig struct {
	Host string
	Port int
	DB   int
}

func NewConfig() (config *Config, err error) {
	viper.AutomaticEnv()

	// Map Viper Variables to ENV Aliases
	viper.RegisterAlias("Database.Host", "IAM_DATABASE_HOST")
	viper.RegisterAlias("Database.Port", "IAM_DATABASE_PORT")
	viper.RegisterAlias("Database.User", "IAM_DATABASE_USER")

	viper.RegisterAlias("Redis.Host", "IAM_REDIS_HOST")
	viper.RegisterAlias("Redis.Port", "IAM_REDIS_PORT")
	viper.RegisterAlias("Redis.DB", "IAM_REDIS_DB")

	// Set Defaults to Local Env
	viper.SetDefault("Database.Host", "localhost")
	viper.SetDefault("Database.Port", "5432")
	viper.SetDefault("Database.User", "postgres")

	viper.SetDefault("Redis.Host", "localhost")
	viper.SetDefault("Redis.Port", "6379")
	viper.SetDefault("Redis.DB", 0)

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	log.Info().Interface("config", config).Msg("loaded config")

	return config, nil
}
