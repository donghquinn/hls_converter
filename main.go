package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/donghquinn/hls_converter/configs"
	summary "github.com/donghquinn/hls_converter/kafka"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	configs.SetGlobalConfiguration()

	configs.SetKafkaConfig()

	// Create directories if they don't exist
	createDirectories()

	// Create Kafka consumer
	kafkaInstance, err := summary.NewKafkaInstance()
	if err != nil {
		log.Fatalf("Failed to create Kafka instance: %v", err)
	}
	defer kafkaInstance.Close()

	// Create a context that will be canceled on SIGINT or SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %s, shutting down...", sig)
		cancel()
	}()

	// Start Kafka consumer
	log.Println("Starting HLS converter Kafka consumer")
	kafkaInstance.Consume(ctx)

	// Wait for graceful shutdown
	<-ctx.Done()
	log.Println("Shutdown complete")
}

func createDirectories() {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(configs.GlobalConfiguration.UploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(configs.GlobalConfiguration.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
}
