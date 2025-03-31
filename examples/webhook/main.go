package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/hatchet-dev/hatchet/pkg/client"
	"github.com/hatchet-dev/hatchet/pkg/worker"
)

type userCreateEvent struct {
	Username string            `json:"username"`
	UserID   string            `json:"user_id"`
	Data     map[string]string `json:"data"`
}

type output struct {
	Message string `json:"message"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load environment variables")
		os.Exit(1)
	}

	c, err := client.New()
	if err != nil {
		log.Println("Failed to initialize client")
		os.Exit(1)
	}

	w, err := worker.NewWorker(
		worker.WithClient(
			c,
		),
	)
	if err != nil {
		log.Println("Failed to initialize worker")
		os.Exit(1)
	}

	workflow := "webhook"
	event := "user:create:webhook"
	wf := &worker.WorkflowJob{
		Name:        workflow,
		Description: workflow,
		Steps: []*worker.WorkflowStep{
			worker.Fn(func(ctx worker.HatchetContext) (result *output, err error) {
				log.Printf("step name: %s", ctx.StepName())
				return &output{
					Message: "hi from " + ctx.StepName(),
				}, nil
			}).SetName("webhook-step-one").SetTimeout("10s"),
			worker.Fn(func(ctx worker.HatchetContext) (result *output, err error) {
				log.Printf("step name: %s", ctx.StepName())
				return &output{
					Message: "hi from " + ctx.StepName(),
				}, nil
			}).SetName("webhook-step-one").SetTimeout("10s"),
		},
	}

	// Get webhook secret from environment variable
	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Println("WEBHOOK_SECRET environment variable is required but not set")
		os.Exit(1)
	}

	handler := w.WebhookHttpHandler(worker.WebhookHandlerOptions{
		Secret: webhookSecret,
	}, wf)
	port := "8741"
	err = run("webhook-demo", w, port, handler, c, workflow, event)
	if err != nil {
		log.Println("Failed to run webhook demo")
		os.Exit(1)
	}
}