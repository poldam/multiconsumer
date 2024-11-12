package main

import (
	"context"
	"github.com/segmentio/kafka-go"
)
// FullConfig represents the entire application configuration
type FullConfig struct {
    Redis          EnvConfig[RedisConfig]   `json:"redis"`
    Mongo          EnvConfig[MongoConfig]   `json:"mongo"`
    MySQL          EnvConfig[MySQLConfig]   `json:"mysql"`
    KafkaConsumers []ConsumerConfig         `json:"kafkaConsumers"`
}

// EnvConfig represents a generic configuration for multiple environments
type EnvConfig[T any] struct {
    Production  T `json:"production"`
    Development T `json:"development"`
    Staging     T `json:"staging"`
}

// RedisConfig represents Redis connection details
type RedisConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

// MongoConfig represents MongoDB connection details
type MongoConfig struct {
    Server   string `json:"server"`
    Port     int    `json:"port"`
    Username string `json:"username"`
    Password string `json:"password"`
    Database string `json:"database"`
}

// MySQLConfig represents MySQL connection details
type MySQLConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    User     string `json:"user"`
    Password string `json:"password"`
    Database string `json:"database"`
}

// ConsumerConfig represents the configuration for a Kafka consumer
type ConsumerConfig struct {
    Brokers    []string               `json:"brokers"`
    Topic      string                 `json:"topic"`
    GroupID    string                 `json:"group_id"`
    LogFile    string                 `json:"log_file"`
    LogPrefix  string                 `json:"log_prefix"`
    DebugMode  bool                   `json:"debug_mode"`
    Settings   map[string]interface{} `json:"settings"` 
    HandlerName string                `json:"handler_name"` 
}

// KafkaConsumer represents a Kafka consumer with logging and consumption logic
type KafkaConsumer struct {
    reader         *kafka.Reader
    logger         func(level string, msg string, args ...interface{})
    ctx            context.Context
    cancel         context.CancelFunc
    handleMessage  func(message kafka.Message, config ConsumerConfig, logFunc func(level string, msg string, args ...interface{}))
    consumerConfig ConsumerConfig 
}
