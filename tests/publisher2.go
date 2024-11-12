package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"
)

// Country struct to hold detailed information about each country
type Country struct {
	Name         string    `json:"name"`
	Population   int       `json:"population"`
	Size         int       `json:"size"` // in square kilometers
	Description  string    `json:"description"`
	YearCreated  int       `json:"year_created"`
	PublishedAt  time.Time `json:"published_at"`
}

func main() {
	// Kafka writer configuration
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{"localhost:9092"}, // Kafka broker addresses
		Topic:        "cities",            // The topic to send messages to
		Balancer:     &kafka.LeastBytes{},        // Balancer to distribute messages evenly
		RequiredAcks: -1,           // // Wait for all replicas to acknowledge the message (most reliable)
	})

	// Ensure the writer is closed when the main function exits to avoid resource leaks
	defer func() {
		if err := writer.Close(); err != nil {
			fmt.Printf("Error closing writer: %v\n", err)
		}
	}()

	// Random seed initialization for message generation
	rand.Seed(time.Now().UnixNano())

	// Array of example countries with attributes
	countries := []Country{
		{"Paris", 67000000, 551695, "Known for its rich culture, cuisine, and the iconic Eiffel Tower.", 843, time.Now()},
		{"Tokyo", 125800000, 377975, "A country famous for its technological advances and traditional arts.", 660, time.Now()},
		{"Brazilia", 213000000, 8515767, "The largest country in South America, known for its Amazon Rainforest and Carnival.", 1822, time.Now()},
		{"Cambera", 25690000, 7692024, "Known for its unique wildlife and the Great Barrier Reef.", 1901, time.Now()},
		{"Otawa", 38000000, 9984670, "Famous for its natural beauty, polite people, and maple syrup.", 1867, time.Now()},
		{"Delhi", 1400000000, 3287263, "A country with a deep history, known for its diversity and landmarks like the Taj Mahal.", 1947, time.Now()},
		{"Berlin", 83000000, 357022, "Renowned for its engineering, beer, and historical sites.", 1871, time.Now()},
		{"Rome", 60000000, 301340, "Home of pasta, pizza, and world-class art and architecture.", 1861, time.Now()},
		{"Seoul", 52000000, 100210, "Known for K-pop, high-tech cities, and traditional palaces.", 1948, time.Now()},
	}

	fmt.Println("Starting to publish detailed country messages...")

	// Publish 10 random detailed messages to the Kafka topic
	for i := 0; i < 10; i++ {
		// Select a random country
		country := countries[rand.Intn(len(countries))]

		// Convert the country struct to JSON format
		messageValue, err := json.Marshal(country)
		if err != nil {
			fmt.Printf("Failed to marshal country data: %v\n", err)
			continue
		}

		// Create the Kafka message with a key and value
		err = writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(fmt.Sprintf("key-%d", i)), // Assign unique keys to each message
				Value: messageValue,                     // JSON-encoded message content
			},
		)
		if err != nil {
			fmt.Printf("Failed to write message: %v\n", err)
		} else {
			fmt.Printf("Sent message: %s\n", string(messageValue))
		}

		// Delay between message publications for demonstration purposes
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Finished publishing detailed cities messages.")
}
