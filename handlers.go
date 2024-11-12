package main

import (
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
)

// Define custom handler functions for each consumer
func handler1 (
    message kafka.Message,
    config ConsumerConfig,
    logFunc func(level string, msg string, args ...interface{}),
    redisConfig RedisConfig,
    mongoConfig MongoConfig,
    mysqlConfig MySQLConfig,
) {
    logFunc("INFO", "Handler1 processing message with topic: %s", config.Topic)
    logFunc("DEBUG", "Handler1 processing message with settings: %+v", config.Settings)
    
    var data map[string]interface{}
    err := json.Unmarshal(message.Value, &data)
    if err != nil {
        logFunc("ERROR", "Failed to unmarshal message: %v", err)
        return
    }

    if name, ok := data["name"]; ok && name == "" {
        logFunc("WARNING", "Message has an empty 'name' field")
    }

	if army, ok := data["army"].(map[string]interface{}); ok {
		if motos, ok := army["motos"].([]interface{}); ok {
			// Convert to a slice of strings for easier printing
			var motoStrings []string
			for _, moto := range motos {
				if motoStr, ok := moto.(string); ok {
					motoStrings = append(motoStrings, motoStr)
				}
			}
			fmt.Printf("%v\n", motoStrings) // Print the slice of strings
		}
	}

	if mappings, ok := config.Settings["mappings"].(map[string]interface{}); ok {
        if code, ok := mappings["code"].(string); ok {
            logFunc("INFO", "Using code mapping: %s", code)
        }

        if replaceWith, ok := mappings["replacewith"].(string); ok {
            logFunc("INFO", "Replace code with: %s", replaceWith)
        }
    }

    fmt.Printf("Consumer - %s: %s\n", data["name"], data["description"])
    logFunc("INFO", "Successfully processed message from topic: %s", config.Topic)
}

func handler2 (
    message kafka.Message,
    config ConsumerConfig,
    logFunc func(level string, msg string, args ...interface{}),
    redisConfig RedisConfig,
    mongoConfig MongoConfig,
    mysqlConfig MySQLConfig,
) {
    logFunc("INFO", "Handler2 processing message with topic: %s", config.Topic)
    logFunc("INFO", "Redis Host: %s, Port: %d", redisConfig.Host, redisConfig.Port)
    logFunc("INFO", "Mongo Server: %s, Port: %d", mongoConfig.Server, mongoConfig.Port)
    logFunc("INFO", "MySQL Host: %s, Port: %d", mysqlConfig.Host, mysqlConfig.Port)

    fmt.Printf("Consumer received message: %s\n", string(message.Value))
}
