package main

import (
	"context"
	"encoding/json"
	"fmt"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strings"

	"github.com/segmentio/kafka-go"
)

// ReadConfig reads the full configuration from a JSON file
func ReadConfig(filename string) (*FullConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config FullConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	return &config, nil
}

// customLogger creates a logger function with a configurable log level
func customLogger(prefix string, file *os.File, enableDebug bool) func(level string, msg string, args ...interface{}) {
    logger := log.New(file, "[" + prefix + "] ", log.Ldate|log.Ltime|log.Lshortfile)
    return func(level string, msg string, args ...interface{}) {
        // Filter out specific debug messages if needed
		if strings.Contains(msg, "no messages received from kafka") || strings.Contains(msg, "topic: ") || strings.Contains(msg, "partition "){
			return // Do not log this message
		}

        if level == "DEBUG" && !enableDebug {
            return // Do not log DEBUG messages unless debug mode is on
        }
        
        logger.Printf(fmt.Sprintf("%s: %s", level, msg), args...)
    }
}

func createConsumer(brokers []string, topic, groupID string, logger func(level string, msg string, args ...interface{})) *kafka.Reader {
    readerConfig := kafka.ReaderConfig{
        Brokers:     brokers,
        Topic:       topic,
        GroupID:     groupID,
        MinBytes:    1,
        MaxBytes:    10e6,
        MaxWait:     500 * time.Millisecond,
        Logger:      kafka.LoggerFunc(func(msg string, args ...interface{}) { logger("DEBUG", msg, args...) }),
    }
    return kafka.NewReader(readerConfig)
}

// NewKafkaConsumer creates a new KafkaConsumer with the given configuration and handler function
func NewKafkaConsumer(config ConsumerConfig, handler func(message kafka.Message, config ConsumerConfig, logFunc func(level string, msg string, args ...interface{}))) *KafkaConsumer {
    ctx, cancel := context.WithCancel(context.Background())

    logFile, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatalf("Failed to create log file: %v\n", err)
    }

    unifiedLogger := customLogger(config.LogPrefix, logFile, config.DebugMode)

    consumer := &KafkaConsumer{
        reader:         createConsumer(config.Brokers, config.Topic, config.GroupID, unifiedLogger),
        logger:         unifiedLogger,
        ctx:            ctx,
        cancel:         cancel,
        handleMessage:  handler,
        consumerConfig: config, // Assign the configuration here
    }
    return consumer
}

// Start begins consuming messages from Kafka
func (kc *KafkaConsumer) Start() {
    kc.logger("INFO", "Starting Kafka consumer...")

    go func() {
        var wasDisconnected bool
        for {
            select {
            case <-kc.ctx.Done():
                kc.logger("WARNING", "Consumer shutdown signal received. Stopping...")
                return
            default:
                message, err := kc.reader.FetchMessage(kc.ctx)
                if err != nil {
                    if !wasDisconnected {
                        kc.logger("ERROR", "Lost connection to Kafka. Shutting down consumer and retrying in 5 seconds...")
                        wasDisconnected = true
                        kc.reader.Close()
                    }
                    time.Sleep(5 * time.Second)
                    kc.reader = createConsumer(kc.consumerConfig.Brokers, kc.consumerConfig.Topic, kc.consumerConfig.GroupID, kc.logger)
                    kc.logger("INFO", "Attempting to reconnect to Kafka...")
                    continue
                }

                if wasDisconnected {
                    kc.logger("INFO", "Reconnected to Kafka successfully. Consumer is ready.")
                    wasDisconnected = false
                }

                // Call the custom handler function with the message and logger
                kc.handleMessage(message, kc.consumerConfig, kc.logger)

                if err := kc.reader.CommitMessages(kc.ctx, message); err != nil {
                    kc.logger("ERROR", "Failed to commit message: %v\n", err)
                }
            }
        }
    }()
}

// Stop stops the Kafka consumer gracefully
func (kc *KafkaConsumer) Stop() {
    // Cancel the context to signal the consumer to stop
    kc.cancel()
    // Close the Kafka reader and handle any potential error
    if err := kc.reader.Close(); err != nil {
        kc.logger("ERROR", "Error closing reader: %v\n", err)
    }
    kc.logger("INFO", "Kafka consumer has been stopped.")
}

func GetEnvConfig(env string, config FullConfig) (RedisConfig, MongoConfig, MySQLConfig) {
    switch env {
    case "production":
        return config.Redis.Production, config.Mongo.Production, config.MySQL.Production
    case "staging":
        return config.Redis.Staging, config.Mongo.Staging, config.MySQL.Staging
    case "development":
        return config.Redis.Development, config.Mongo.Development, config.MySQL.Development
    default:
        log.Fatalf("Unknown environment: %s", env)
        return RedisConfig{}, MongoConfig{}, MySQLConfig{} // Unreachable, but required by Go
    }
}

// Command-Line Flag: Use '-env production', -env staging, or -env development to specify
func main() {
    // Parse environment from command-line flag or default to "development"
    env := flag.String("env", "development", "Specify the environment: production, staging, development")
    flag.Parse()

    config, err := ReadConfig("config.json")
    if err != nil {
        log.Fatalf("Failed to load configuration: %v\n", err)
    }

    // Get the environment-specific configurations
    redisConfig, mongoConfig, mysqlConfig := GetEnvConfig(*env, *config) 

    // Map handler names to functions
    handlers := map[string]func(kafka.Message, ConsumerConfig, func(level string, msg string, args ...interface{}), RedisConfig, MongoConfig, MySQLConfig) {
        "handler1": handler1,
        "handler2": handler2,
    }

    // Print configurations for verification
    fmt.Printf("Using environment: %s\n", *env)
    fmt.Printf("Redis Config: Host=%s, Port=%d\n", redisConfig.Host, redisConfig.Port)
    fmt.Printf("Mongo Config: Server=%s, Port=%d, User=%s, Database=%s\n",
        mongoConfig.Server, mongoConfig.Port, mongoConfig.Username, mongoConfig.Database)
    fmt.Printf("MySQL Config: Host=%s, Port=%d, User=%s, Database=%s\n",
        mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.User, mysqlConfig.Database)

	// Create consumers based on the loaded configuration and specified handler from the config
    var consumers []*KafkaConsumer
    for _, consumerConfig := range config.KafkaConsumers {
        handlerFunc, exists := handlers[consumerConfig.HandlerName]
        if !exists {
            log.Fatalf("Handler %s not found\n", consumerConfig.HandlerName)
        }

        consumer := NewKafkaConsumer(consumerConfig, func(message kafka.Message, config ConsumerConfig, logFunc func(level string, msg string, args ...interface{})) {
            handlerFunc(message, config, logFunc, redisConfig, mongoConfig, mysqlConfig)
        })
        consumers = append(consumers, consumer)
        consumer.Start()
    }

	// Listen for termination signals to stop all consumers gracefully
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
	for _, consumer := range consumers {
		consumer.logger("INFO", "Received termination signal. Shutting down...")
		consumer.Stop()
	}
}
