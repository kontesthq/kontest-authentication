package service

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"kontest-authentication/utils/kafka_utils"
	"log"
	"time"
)

var producer *kafka.Producer

func InitKafka(broker string) {
	var err error

	producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
	})

	if err != nil {
		log.Fatalf("Failed to create kafka_utils producer: %s", err)
	}
}

func PublishMessage(topic string, message string) error {
	// Create a kafka_utils message
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}

	// Produce the message
	err := producer.Produce(msg, nil)
	if err != nil {
		return err
	}

	// Wait for the delivery report
	e := <-producer.Events()
	if ev := e.(*kafka.Message); ev.TopicPartition.Error != nil {
		return ev.TopicPartition.Error
	}

	return nil
}

// PublishRegistrationMessage publishes a registration message to kafka_utils
func PublishRegistrationMessage(email string) error {
	// Create JSON message
	jsonMessage := map[string]interface{}{
		"email":            email,
		"registrationDate": time.Now(),
	}

	// Convert the map to JSON string
	jsonString, err := json.Marshal(jsonMessage)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Get the topic from environment variable
	topic := kafka_utils.UserRegistrationEventTopic.DefaultValue
	if topic == "" {
		return fmt.Errorf("KafkaUserRegistrationEventTopic is not set in environment variables")
	}

	// Publish the message to kafka_utils using PublishMessage function
	if err := PublishMessage(topic, string(jsonString)); err != nil {
		return fmt.Errorf("failed to publish registration message: %v", err)
	}

	return nil
}
