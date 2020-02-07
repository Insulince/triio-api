package configuration

import (
	"os"
	"strings"
	"strconv"
	"log"
	"errors"
	"fmt"
)

type Config struct {
	Port      int         `json:"port"`
	JwtSecret string      `json:"jwtSecret"`
	Cors      CorsConfig  `json:"cors"`
	Mongo     MongoConfig `json:"mongo"`
}

type MongoConfig struct {
	ConnectionString string   `json:"connectionString"`
	DatabaseName     string   `json:"databaseName"`
	CollectionNames  []string `json:"collectionNames"`
}

type CorsConfig struct {
	AllowedOrigins   []string `json:"allowedOrigins"`
	AllowedMethods   []string `json:"allowedMethods"`
	AllowedHeaders   []string `json:"allowedHeaders"`
	AllowCredentials bool     `json:"allowCredentials"`
}

func LoadConfig() (config *Config, err error) {
	log.Printf("Loading config from environment variables...\n")
	config = &Config{}

	if value, present := os.LookupEnv("PORT"); present {
		port, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Environment variable \"PORT\" is invalid! Must be an integer, but got \"%v\".\n", value))
		}
		config.Port = int(port)
	} else {
		return nil, createEnvironmentVariableNotPresentError("PORT")
	}

	if value, present := os.LookupEnv("JWT_SECRET"); present {
		config.JwtSecret = value
	} else {
		return nil, createEnvironmentVariableNotPresentError("JWT_SECRET")
	}

	if value, present := os.LookupEnv("MONGO_CONNECTION_STRING"); present {
		config.Mongo.ConnectionString = value
	} else {
		return nil, createEnvironmentVariableNotPresentError("MONGO_CONNECTION_STRING")
	}
	if value, present := os.LookupEnv("MONGO_DATABASE_NAME"); present {
		config.Mongo.DatabaseName = value
	} else {
		return nil, createEnvironmentVariableNotPresentError("MONGO_DATABASE_NAME")
	}

	if value, present := os.LookupEnv("CORS_ALLOWED_ORIGINS"); present {
		config.Cors.AllowedOrigins = strings.Split(value, ",")
	} else {
		return nil, createEnvironmentVariableNotPresentError("CORS_ALLOWED_ORIGINS")
	}
	if value, present := os.LookupEnv("CORS_ALLOWED_METHODS"); present {
		config.Cors.AllowedMethods = strings.Split(value, ",")
	} else {
		return nil, createEnvironmentVariableNotPresentError("CORS_ALLOWED_METHODS")
	}
	if value, present := os.LookupEnv("CORS_ALLOWED_HEADERS"); present {
		config.Cors.AllowedHeaders = strings.Split(value, ",")
	} else {
		return nil, createEnvironmentVariableNotPresentError("CORS_ALLOWED_HEADERS")
	}
	if value, present := os.LookupEnv("CORS_ALLOW_CREDENTIALS"); present {
		allowCredentials, err := strconv.ParseBool(value)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Environment variable \"CORS_ALLOW_CREDENTIALS\" is invalid! Must be one of [\"true\", \"false\"], but got \"%v\".\n", value))
		}
		config.Cors.AllowCredentials = allowCredentials
	} else {
		return nil, createEnvironmentVariableNotPresentError("CORS_ALLOW_CREDENTIALS")
	}

	log.Printf("Successfully loaded config.\n")
	return config, nil
}

func createEnvironmentVariableNotPresentError(environmentVariableName string) (err error) {
	return errors.New(fmt.Sprintf("Environment variable \"%v\" not provided!\n", environmentVariableName))
}
