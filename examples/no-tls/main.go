package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/hatchet-dev/hatchet/pkg/client"
	"github.com/hatchet-dev/hatchet/pkg/cmdutils"
	"github.com/hatchet-dev/hatchet/pkg/worker"
)

type stepOutput struct{}

// handleError logs the detailed error for debugging purposes
// but terminates the program with a generic message to avoid
// exposing sensitive information.
func handleError(context string, err error) {
	// Log the detailed error for debugging (could go to a file)
	log.Printf("%s: %v", context, err)
	
	// Exit with a generic message for the user
	fmt.Fprintf(os.Stderr, "Error: %s. Check logs for details.\n", context)
	os.Exit(1)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		handleError("Failed to load environment", err)
	}

	c, err := client.New()

	if err != nil {
		handleError("Failed to create client", err)
	}

	w, err := worker.NewWorker(
		worker.WithClient(
			c,
		),
		worker.WithMaxRuns(1),
	)
	if err != nil {
		handleError("Failed to create worker", err)
	}

	testSvc := w.NewService("test")

	err = testSvc.On(
		worker.Events("simple"),
		&worker.WorkflowJob{
			Name:        "simple-workflow",
			Description: "Simple one-step workflow.",
			Steps: []*worker.WorkflowStep{
				worker.Fn(func(ctx worker.HatchetContext) (result *stepOutput, err error) {
					fmt.Println("executed step 1")

					return &stepOutput{}, nil
				},
				).SetName("step-one"),
			},
		},
	)
	if err != nil {
		handleError("Failed to register workflow", err)
	}

	interruptCtx, cancel := cmdutils.InterruptContextFromChan(cmdutils.InterruptChan())
	defer cancel()

	cleanup, err := w.Start()
	if err != nil {
		handleError("Failed to start worker", err)
	}

	<-interruptCtx.Done()
	if err := cleanup(); err != nil {
		handleError("Failed to clean up worker resources", err)
	}
}