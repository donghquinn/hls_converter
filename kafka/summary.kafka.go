package summary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/donghquinn/hls_converter/biz/converter"
	"github.com/donghquinn/hls_converter/configs"
	"github.com/segmentio/kafka-go"
)

// FileMessage represents the message structure expected from Kafka
type FileMessage struct {
	RequestID  string `json:"requestId"`
	FilePath   string `json:"filePath"`
	OutputPath string `json:"outputPath,omitempty"`
}

// CompletionMessage represents the message to be sent after conversion
type CompletionMessage struct {
	RequestID    string    `json:"requestId"`
	Status       string    `json:"status"`
	InputFile    string    `json:"inputFile"`
	OutputFile   string    `json:"outputFile"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
	CompletedAt  time.Time `json:"completedAt"`
}

type KafkaInterface struct {
	ConsumerConn *kafka.Reader
	ProducerConn *kafka.Writer
	Broker       string
	InputTopic   string
	OutputTopic  string
	GroupID      string
	ConsumerID   string
	OutputDir    string
}

func NewKafkaInstance() (*KafkaInterface, error) {
	kafkaConfig := configs.KafkaConfig

	log.Printf("[DEBUGGING] Kafka Broker: %s", kafkaConfig.Broker)

	// Validate broker connection
	conn, err := kafka.Dial("tcp", kafkaConfig.Broker)
	if err != nil {
		return nil, fmt.Errorf("failed to dial Kafka broker: %v", err)
	}
	defer conn.Close()

	log.Printf("[DEBUGGING] Kafka Input Topic: %s", kafkaConfig.InputTopic)

	// Ensure topic exists
	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             kafkaConfig.InputTopic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil {
		// If topic already exists, log but continue
		log.Printf("Note: could not create topic (may already exist): %v", err)
	}

	log.Printf("[DEBUGGING] Kafka OUPUT Topic: %s", kafkaConfig.OutputTopic)

	// Create output topic if provided
	if kafkaConfig.OutputTopic != "" {
		err = conn.CreateTopics(kafka.TopicConfig{
			Topic:             kafkaConfig.OutputTopic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
		if err != nil {
			log.Printf("Note: could not create output topic (may already exist): %v", err)
		}
	}

	log.Printf("[DEBUGGING] Kafka GroupID: %s", kafkaConfig.GroupId)

	// Create consumer
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaConfig.Broker},
		GroupID:        kafkaConfig.GroupId,
		Topic:          kafkaConfig.InputTopic,
		MaxBytes:       10e6, // 10MB
		CommitInterval: 0,    // Disable auto-commit
	})

	// Create producer for completion messages if output topic is provided
	var producer *kafka.Writer
	if kafkaConfig.OutputTopic != "" {
		producer = &kafka.Writer{
			Addr:     kafka.TCP(kafkaConfig.Broker),
			Topic:    kafkaConfig.OutputTopic,
			Balancer: &kafka.LeastBytes{},
		}
	}

	instance := &KafkaInterface{
		ConsumerConn: consumer,
		ProducerConn: producer,
		Broker:       kafkaConfig.Broker,
		InputTopic:   kafkaConfig.InputTopic,
		OutputTopic:  kafkaConfig.OutputTopic,
		GroupID:      kafkaConfig.GroupId,
		ConsumerID:   kafkaConfig.ConsumerId,
		OutputDir:    configs.GlobalConfiguration.OutputDir,
	}

	return instance, nil
}

func (k *KafkaInterface) Consume(ctx context.Context) {
	if k.ConsumerConn == nil {
		log.Println("[KAFKA] Consumer connection is nil")
		return
	}

	defer func() {
		if err := k.ConsumerConn.Close(); err != nil {
			log.Printf("[KAFKA] Failed to close consumer: %v", err)
		}
	}()

	log.Printf("[KAFKA] Starting consumer for topic: %s", k.InputTopic)

	for {
		// Check if context is done
		select {
		case <-ctx.Done():
			log.Println("[KAFKA] Context done, stopping consumer")
			return
		default:
		}

		// Read message with timeout
		readCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		m, err := k.ConsumerConn.FetchMessage(readCtx)
		cancel()

		if err != nil {
			if err == context.DeadlineExceeded {
				// This is just a timeout, continue
				continue
			}
			log.Printf("[KAFKA] Error reading message: %v", err)
			time.Sleep(1 * time.Second) // Avoid tight loop in case of persistent errors
			continue
		}

		log.Printf("[KAFKA] Received message from %s/%d/%d", m.Topic, m.Partition, m.Offset)

		// Process the message
		if err := k.processMessage(ctx, m); err != nil {
			log.Printf("[KAFKA] Error processing message: %v", err)
			// Decide whether to commit the message or not based on error type
			if shouldCommitOnError(err) {
				if commitErr := k.ConsumerConn.CommitMessages(ctx, m); commitErr != nil {
					log.Printf("[KAFKA] Failed to commit failed message: %v", commitErr)
				}
			}
			continue
		}

		// Commit the message after successful processing
		if err := k.ConsumerConn.CommitMessages(ctx, m); err != nil {
			log.Printf("[KAFKA] Failed to commit message: %v", err)
		}
	}
}

func (k *KafkaInterface) processMessage(ctx context.Context, m kafka.Message) error {
	// Parse the message
	var fileMsg FileMessage
	if err := json.Unmarshal(m.Value, &fileMsg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	if fileMsg.RequestID == "" || fileMsg.FilePath == "" {
		return fmt.Errorf("invalid message format: missing requestId or filePath")
	}

	// Validate that the file exists
	if _, err := os.Stat(fileMsg.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", fileMsg.FilePath)
	}

	// Create output directory
	outputDir := filepath.Join(k.OutputDir, fileMsg.RequestID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create a conversion job
	job := &converter.ConversionJob{
		ID:        fileMsg.RequestID,
		InputFile: fileMsg.FilePath,
		OutputDir: outputDir,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	log.Printf("[KAFKA] Starting HLS conversion for request %s: %s -> %s",
		fileMsg.RequestID, fileMsg.FilePath, outputDir)

	// Perform the conversion
	err := converter.ConvertToHLS(job)
	completionMsg := CompletionMessage{
		RequestID:   fileMsg.RequestID,
		InputFile:   fileMsg.FilePath,
		OutputFile:  filepath.Join(outputDir, "playlist.m3u8"),
		Status:      job.Status,
		CompletedAt: time.Now(),
	}

	if err != nil {
		completionMsg.Status = "failed"
		completionMsg.ErrorMessage = err.Error()
		log.Printf("[KAFKA] Conversion failed for request %s: %v", fileMsg.RequestID, err)
	} else {
		log.Printf("[KAFKA] Conversion completed for request %s", fileMsg.RequestID)
	}

	// Send completion message if output topic is configured
	if k.ProducerConn != nil && k.OutputTopic != "" {
		if err := k.sendCompletionMessage(ctx, completionMsg); err != nil {
			log.Printf("[KAFKA] Failed to send completion message: %v", err)
			// Don't return error here, we still want to commit the input message
		}
	}

	return nil
}

func (k *KafkaInterface) sendCompletionMessage(ctx context.Context, msg CompletionMessage) error {
	if k.ProducerConn == nil {
		return fmt.Errorf("producer connection is nil")
	}

	value, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal completion message: %v", err)
	}

	return k.ProducerConn.WriteMessages(ctx, kafka.Message{
		Key:   []byte(msg.RequestID),
		Value: value,
	})
}

// shouldCommitOnError determines if we should commit a message that resulted in error
// This helps prevent endless reprocessing of bad messages
func shouldCommitOnError(err error) bool {
	// Add logic to determine if error is permanent (true) or temporary (false)
	// For example, file not found or invalid message format are permanent errors
	return true
}

// Close gracefully shuts down the Kafka connections
func (k *KafkaInterface) Close() {
	if k.ConsumerConn != nil {
		k.ConsumerConn.Close()
	}
	if k.ProducerConn != nil {
		k.ProducerConn.Close()
	}
}
