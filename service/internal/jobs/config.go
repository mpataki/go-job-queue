package jobs

import "os"

type Config struct {
	redisAddr string
}

func NewConfig() (*Config, error) {
	c := Config{
		redisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
	}

	return &c, nil
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)

	if !exists {
		return defaultValue
	}

	return value
}
