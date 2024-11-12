package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"
)

// Army struct to hold military details
type Army struct {
    Airforce int      `json:"airforce"`
    Tanks    int      `json:"tanks"`
    Soldiers int      `json:"soldiers"`
    Motos    []string `json:"motos"`
}

// Country struct with additional military details
type Country struct {
    Name        string    `json:"name"`
    Population  int       `json:"population"`
    Size        int       `json:"size"`
    Description string    `json:"description"`
    Year        int       `json:"year"`
    Timestamp   time.Time `json:"timestamp"`
    Army        Army      `json:"army"` // New field for military details
}

func main() {
	// Kafka writer configuration
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{"localhost:9092"}, // Kafka broker addresses
		Topic:        "countries",            // The topic to send messages to
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

	// Array of example countries with attributes and military details
	countries := []Country{
		{
			"France", 67000000, 551695, "Known for its rich culture, cuisine, and the iconic Eiffel Tower.", 843, time.Now(),
			Army{1500, 3000, 100000, []string{"Rafale", "Leclerc", "Mirage"}},
		},
		{
			"Japan", 125800000, 377975, "A country famous for its technological advances and traditional arts.", 660, time.Now(),
			Army{1200, 2900, 250000, []string{"F-15J", "Type 10", "OH-1"}},
		},
		{
			"Brazil", 213000000, 8515767, "The largest country in South America, known for its Amazon Rainforest and Carnival.", 1822, time.Now(),
			Army{800, 1500, 235000, []string{"AMX", "Leopard 1", "Super Tucano"}},
		},
		{
			"Australia", 25690000, 7692024, "Known for its unique wildlife and the Great Barrier Reef.", 1901, time.Now(),
			Army{700, 1200, 59000, []string{"F/A-18", "M1 Abrams", "Bushmaster"}},
		},
		{
			"Canada", 38000000, 9984670, "Famous for its natural beauty, polite people, and maple syrup.", 1867, time.Now(),
			Army{450, 800, 68000, []string{"CF-18", "Leopard 2", "LAV III"}},
		},
		{
			"India", 1400000000, 3287263, "A country with a deep history, known for its diversity and landmarks like the Taj Mahal.", 1947, time.Now(),
			Army{1700, 4000, 1400000, []string{"Sukhoi Su-30MKI", "Arjun", "INSAS"}},
		},
		{
			"Germany", 83000000, 357022, "Renowned for its engineering, beer, and historical sites.", 1871, time.Now(),
			Army{900, 1600, 184000, []string{"Eurofighter", "Leopard 2", "Boxer"}},
		},
		{
			"Kenya", 54000000, 580367, "Known for its stunning savannas and rich wildlife.", 1963, time.Now(),
			Army{150, 300, 24000, []string{"F-5E", "T-72", "AML-90"}},
		},
		{
			"Italy", 60000000, 301340, "Home of pasta, pizza, and world-class art and architecture.", 1861, time.Now(),
			Army{850, 1300, 175000, []string{"AMX", "Ariete", "Mangusta"}},
		},
		{
			"South Korea", 52000000, 100210, "Known for K-pop, high-tech cities, and traditional palaces.", 1948, time.Now(),
			Army{1000, 2000, 600000, []string{"KF-21", "K2 Black Panther", "K9 Thunder"}},
		},
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

	fmt.Println("Finished publishing detailed country messages.")
}
