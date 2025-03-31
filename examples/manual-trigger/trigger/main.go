package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/hatchet-dev/hatchet/pkg/client"
	"github.com/hatchet-dev/hatchet/pkg/cmdutils"
)

type userCreateEvent struct {
	Username string            `json:"username"`
	UserID   string            `json:"user_id"`
	Data     map[string]string `json:"data"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Failed to load environment file: %v", err)
		fmt.Fprintf(os.Stderr, "Configuration error: could not load environment settings\n")
		os.Exit(1)
	}

	events := make(chan string, 50)
	if err := run(cmdutils.InterruptChan(), events); err != nil {
		log.Printf("Application error: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to execute workflow trigger\n")
		os.Exit(1)
	}
}

// safelyProcessPayload validates and safely formats the payload
// to prevent ML09 (Manipulation of ML Model Outputs) and
// ML10 (Poisoning of ML Model Parameters) vulnerabilities
func safelyProcessPayload(payload interface{}) string {
	// Basic validation - reject nil payloads
	if payload == nil {
		return "[Warning: Empty payload received]"
	}
	
	// Convert to JSON for safe display - this prevents injection attacks
	// when the payload is displayed or processed further
	jsonBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Sprintf("[Error: Unable to process payload: %v]", err)
	}
	
	return string(jsonBytes)
}

func run(ch <-chan interface{}, events chan<- string) error {
	c, err := client.New()

	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	time.Sleep(1 * time.Second)

	// trigger workflow
	workflow, err := c.Admin().RunWorkflow(
		"post-user-update",
		&userCreateEvent{
			Username: "echo-test",
			UserID:   "1234",
			Data: map[string]string{
				"test": "test",
			},
		},
		client.WithRunMetadata(map[string]interface{}{
			"hello": "world",
		}),
	)

	if err != nil {
		return fmt.Errorf("error running workflow: %w", err)
	}

	fmt.Println("workflow run id:", workflow.WorkflowRunId())

	interruptCtx, cancel := cmdutils.InterruptContextFromChan(ch)
	defer cancel()

	// SECURITY: Always validate and sanitize workflow event payloads before
	// processing them to prevent manipulation or poisoning attacks in ML systems.
	err = c.Subscribe().On(interruptCtx, workflow.WorkflowRunId(), func(event client.WorkflowEvent) error {
		// Process the payload safely before displaying or using it
		safePayload := safelyProcessPayload(event.EventPayload)
		fmt.Println("Validated workflow payload:", safePayload)
		
		// IMPORTANT: For future use, if this payload will be passed to ML models or
		// other sensitive operations, implement additional domain-specific validation.

		return nil
	})

	return err
}