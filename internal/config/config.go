package config

import "os"

type Config struct {
	HTTPPort      string
	MongoURI      string
	MongoDatabase string
	NATSURL       string
}

func Load() Config {

	return Config{
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGO_DATABASE", "Notification"),

		NATSURL: getEnv("NATS_URL", "nats://localhost:4222"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}
