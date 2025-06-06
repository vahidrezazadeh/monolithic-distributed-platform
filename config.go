package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	InstanceID    string
	HTTPPort      int
	GRPCPort      int
	ClusterPort   int
	JoinAddresses []string
	RedisAddress  string
	KafkaAddress  string
	InstanceCount int
	WorkerCount   int
	MaxQueueSize  int
}

func loadConfig() Config {

	cfg := Config{
		InstanceID:    getEnv("INSTANCE_ID", fmt.Sprintf("instance-%d", time.Now().UnixNano())),
		HTTPPort:      getEnvAsInt("HTTP_PORT", 8080),
		GRPCPort:      getEnvAsInt("GRPC_PORT", 50051),
		ClusterPort:   getEnvAsInt("CLUSTER_PORT", 7946),
		RedisAddress:  getEnv("REDIS_ADDRESS", "redis:6379"),
		KafkaAddress:  getEnv("KAFKA_ADDRESS", "kafka:9092"),
		InstanceCount: getEnvAsInt("INSTANCE_COUNT", 3),
		WorkerCount:   getEnvAsInt("WORKER_COUNT", 10),
		MaxQueueSize:  getEnvAsInt("MAX_QUEUE_SIZE", 1000),
	}

	joinAddrs := getEnv("JOIN_ADDRESSES", "")
	if joinAddrs != "" {
		cfg.JoinAddresses = strings.Split(joinAddrs, ",")
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return defaultValue
}
