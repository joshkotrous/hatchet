package main

import (
	"fmt"

	"github.com/joho/godotenv"

	"github.com/hatchet-dev/hatchet/pkg/client"
	"github.com/hatchet-dev/hatchet/pkg/cmdutils"
	"github.com/hatchet-dev/hatchet/pkg/worker"
)

type stepOneOutput struct {
	Message string `json:"message"`
}

// ❓ Backoff

// ... normal function definition
func StepOne(ctx worker.HatchetContext) (result *stepOneOutput, err error) {
	if ctx.RetryCount() < 3 {
		return nil, fmt.Errorf("failure")
	}

	return &stepOneOutput{
		Message: "Success!",
	}, nil
}

// ,

func main() {
	// ...
	err := godotenv.Load()

	if err != nil {
		panic("Failed to load environment variables")
	}

	c, err := client.New()

	if err != nil {
		panic("Failed to initialize client")
	}

	w, err := worker.NewWorker(
		worker.WithClient(
			c,
		),
	)

	if err != nil {
		panic("Failed to initialize worker")
	}

	// ,

	err = w.RegisterWorkflow(
		&worker.WorkflowJob{
			Name:        "retry-with-backoff-workflow",
			On:          worker.NoTrigger(),
			Description: "Demonstrates retry with exponential backoff.",
			Steps: []*worker.WorkflowStep{
				worker.Fn(StepOne).SetName("with-backoff").
					SetRetries(10).
					// 👀 Backoff configuration
					// 👀 Maximum number of seconds to wait between retries
					SetRetryBackoffFactor(2.0).
					// 👀 Factor to increase the wait time between retries.
					// This sequence will be 2s, 4s, 8s, 16s, 32s, 60s... due to the maxSeconds limit
					SetRetryMaxBackoffSeconds(60),
			},
		},
	)

	// ...

	if err != nil {
		panic("Failed to register workflow")
	}

	interruptCtx, cancel := cmdutils.InterruptContextFromChan(cmdutils.InterruptChan())
	defer cancel()

	cleanup, err := w.Start()
	if err != nil {
		panic("Failed to start worker")
	}

	<-interruptCtx.Done()

	if err := cleanup(); err != nil {
		panic("Failed to clean up worker resources")
	}

	// ,
}

// ‼️