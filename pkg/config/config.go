package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr          string
	GRPCAddr          string
	RedisAddr         string
	MinIOEndpoint     string
	MinIOAccessKey    string
	MinIOSecretKey    string
	MinIOUseSSL       bool
	SubmissionsBucket string
	ResultsBucket     string
	TestsBucket       string
	QueueName         string
	EventsChannel     string
	StatePrefix       string
	PollInterval      time.Duration
	BootstrapTestsDir string
}

func Load() Config {
	return Config{
		HTTPAddr:          env("APP_HTTP_ADDR", ":8080"),
		GRPCAddr:          env("APP_GRPC_ADDR", ":9090"),
		RedisAddr:         env("APP_REDIS_ADDR", "localhost:6379"),
		MinIOEndpoint:     env("APP_MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:    env("APP_MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:    env("APP_MINIO_SECRET_KEY", "minioadmin"),
		MinIOUseSSL:       envBool("APP_MINIO_USE_SSL", false),
		SubmissionsBucket: env("APP_SUBMISSIONS_BUCKET", "submissions"),
		ResultsBucket:     env("APP_RESULTS_BUCKET", "results"),
		TestsBucket:       env("APP_TESTS_BUCKET", "tests"),
		QueueName:         env("APP_QUEUE_NAME", "runs:queue"),
		EventsChannel:     env("APP_EVENTS_CHANNEL", "runs:events"),
		StatePrefix:       env("APP_STATE_PREFIX", "runs:state:"),
		PollInterval:      time.Duration(envInt("APP_POLL_INTERVAL_MS", 250)) * time.Millisecond,
		BootstrapTestsDir: env("APP_BOOTSTRAP_TESTS_DIR", "testdata/minio/tests"),
	}
}

func (c Config) ValidateGateway() error {
	if c.RedisAddr == "" || c.MinIOEndpoint == "" {
		return fmt.Errorf("redis/minio config is required")
	}
	return nil
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
