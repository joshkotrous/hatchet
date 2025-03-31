package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/hatchet-dev/hatchet/pkg/client"
	"github.com/hatchet-dev/hatchet/pkg/cmdutils"
	"github.com/hatchet-dev/hatchet/pkg/worker"
)

type sampleEvent struct{}

type timeoutInput struct{}

func main() {
	// Set up logging
	logger := log.New(os.Stderr, "", log.LstdFlags)

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		logger.Fatalf("Failed to load environment variables: %v", err)
	}

	// Initialize client
	client, err := client.New(
		client.InitWorkflows(),
	)
	if err != nil {
		logger.Fatalf("Failed to initialize client: %v", err)
	}

	// Initialize worker
	worker, err := worker.NewWorker(
		worker.WithClient(
			client,
		),
	)
	if err != nil {
		logger.Fatalf("Failed to initialize worker: %v", err)
	}

	err = worker.RegisterAction("timeout:timeout", func(ctx context.Context, input *timeoutInput) (result any, err error) {
		// wait for context done signal
		timeStart := time.Now().UTC()
		<-ctx.Done()
		fmt.Println("context cancelled in ", time.Since(timeStart).Seconds(), " seconds")

		return map[string]interface{}{}, nil
	})
	if err != nil {
		logger.Fatalf("Failed to register action: %v", err)
	}

	interruptCtx, cancel := cmdutils.InterruptContextFromChan(cmdutils.InterruptChan())
	defer cancel()

	cleanup, err := worker.Start()
	if err != nil {
		logger.Fatalf("Error starting worker: %v", err)
	}

	event := sampleEvent{}

	// push an event
	err = client.Event().Push(
		context.Background(),
		"user:create",
		event,
	)
	if err != nil {
		logger.Printf("Error pushing event: %v", err)
		// Non-critical error, continue execution
	}

	for {
		select {
		case <-interruptCtx.Done():
			if err := cleanup(); err != nil {
				logger.Printf("Error cleaning up: %v", err)
				os.Exit(1)
			}
			return
		default:
			time.Sleep(time.Second)
		}
	}
}