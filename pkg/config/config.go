package config

import "os"

type Config struct {
	AWS    AWSConfig
	Server ServerConfig
	JWT    JWTConfig
}

type AWSConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

type ServerConfig struct {
	Port string
}

type JWTConfig struct {
	Secret string
}

func Load() (*Config, error) {
	return &Config{
		AWS: AWSConfig{
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			Region:          getEnv("AWS_REGION", "us-east-1"),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "3001"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "yqKmE7cB7OWpouhuR/x/11HMjx/0Ki5cwwN756K2/dM="),
		},
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
