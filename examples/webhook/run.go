package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hatchet-dev/hatchet/pkg/client"
	"github.com/hatchet-dev/hatchet/pkg/worker"
)

func run(
	name string,
	w *worker.Worker,
	port string,
	handler func(w http.ResponseWriter, r *http.Request), c client.Client, workflow string, event string,
) error {
	// create webserver to handle webhook requests
	mux := http.NewServeMux()

	// Register the HelloHandler to the /hello route
	mux.HandleFunc("/webhook", handler)

	// Create a custom server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	defer func(server *http.Server, ctx context.Context) {
		err := server.Shutdown(ctx)
		if err != nil {
			panic(err)
		}
	}(server, context.Background())

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	// Get webhook secret from environment variable
	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Printf("WARNING: WEBHOOK_SECRET environment variable not set. Generating random secret.")
		
		// Generate a secure random secret
		randomBytes := make([]byte, 16)
		if _, err := rand.Read(randomBytes); err != nil {
			return fmt.Errorf("error generating random secret: %w", err)
		}
		webhookSecret = hex.EncodeToString(randomBytes)
	}

	if err := w.RegisterWebhook(worker.RegisterWebhookWorkerOpts{
		Name:   "test-" + name,
		URL:    fmt.Sprintf("http://localhost:%s/webhook", port),
		Secret: &webhookSecret,
	}); err != nil {
		return fmt.Errorf("error setting up webhook: %w", err)
	}

	time.Sleep(30 * time.Second)

	log.Printf("pushing event")

	testEvent := userCreateEvent{
		Username: "echo-test",
		UserID:   "1234",
		Data: map[string]string{
			"test": "test",
		},
	}

	// push an event
	err := c.Event().Push(
		context.Background(),
		event,
		testEvent,
	)
	if err != nil {
		return fmt.Errorf("error pushing event: %w", err)
	}

	time.Sleep(5 * time.Second)

	return nil
}