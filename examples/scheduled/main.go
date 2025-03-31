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

// handleFatalError logs the error and exits gracefully
func handleFatalError(err error, message string) {
	log.Printf("Fatal error: %s: %v\n", message, err)
	os.Exit(1)
}

// logError logs the error without exiting
func logError(err error, message string) {
	log.Printf("Error: %s: %v\n", message, err)
}

// ❓ Create
// ... normal workflow definition
type printOutput struct{}

func print(ctx context.Context) (result *printOutput, err error) {
	fmt.Println("called print:print")

	return &printOutput{}, nil
}

// ,
func main() {
	// ... initialize client, worker and workflow
	err := godotenv.Load()

	if err != nil {
		handleFatalError(err, "Failed to load environment variables")
	}

	c, err := client.New()

	if err != nil {
		handleFatalError(err, "Failed to create client")
	}

	w, err := worker.NewWorker(
		worker.WithClient(
			c,
		),
	)

	if err != nil {
		handleFatalError(err, "Failed to create worker")
	}

	err = w.RegisterWorkflow(
		&worker.WorkflowJob{
			On:          worker.NoTrigger(),
			Name:        "schedule-workflow",
			Description: "Demonstrates a simple scheduled workflow",
			Steps: []*worker.WorkflowStep{
				worker.Fn(print),
			},
		},
	)

	if err != nil {
		handleFatalError(err, "Failed to register workflow")
	}

	interrupt := cmdutils.InterruptChan()

	cleanup, err := w.Start()

	if err != nil {
		handleFatalError(err, "Failed to start worker")
	}

	// ,

	go func() {
		// 👀 define the scheduled workflow to run in a minute
		schedule, err := c.Schedule().Create(
			context.Background(),
			"schedule-workflow",
			&client.ScheduleOpts{
				// 👀 define the time to run the scheduled workflow, in UTC
				TriggerAt: time.Now().UTC().Add(time.Minute),
				Input: map[string]interface{}{
					"message": "Hello, world!",
				},
				AdditionalMetadata: map[string]string{},
			},
		)

		if err != nil {
			logError(err, "Failed to create schedule")
			return
		}

		fmt.Println(schedule.TriggerAt, schedule.WorkflowName)
	}()

	// ... wait for interrupt signal

	<-interrupt

	if err := cleanup(); err != nil {
		logError(err, "Error cleaning up")
	}

	// ,
}

// !!

func ListScheduledWorkflows() {
	c, err := client.New()

	if err != nil {
		logError(err, "Failed to create client")
		return
	}

	// ❓ List
	schedules, err := c.Schedule().List(context.Background())
	// !!

	if err != nil {
		logError(err, "Failed to list schedules")
		return
	}

	for _, schedule := range *schedules.Rows {
		fmt.Println(schedule.TriggerAt, schedule.WorkflowName)
	}
}

func DeleteScheduledWorkflow(id string) {
	c, err := client.New()

	if err != nil {
		logError(err, "Failed to create client")
		return
	}

	// ❓ Delete
	// 👀 id is the schedule's metadata id, can get it via schedule.Metadata.Id
	err = c.Schedule().Delete(context.Background(), id)
	// !!

	if err != nil {
		logError(err, "Failed to delete schedule")
		return
	}
}