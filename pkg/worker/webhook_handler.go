package worker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hatchet-dev/hatchet/internal/signature"
	"github.com/hatchet-dev/hatchet/pkg/client"
)

type WebhookHandlerOptions struct {
	Secret string
}

type HealthCheckResponse struct {
	Actions []string `json:"actions"`
}

func (w *Worker) WebhookHttpHandler(opts WebhookHandlerOptions, workflows ...workflowConverter) http.HandlerFunc {
	return func(writer http.ResponseWriter, r *http.Request) {
		// health check with actions
		if r.Method == http.MethodGet {
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write([]byte("OK!"))
			return
		}
		if r.Method != http.MethodPut && r.Method != http.MethodPost {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = writer.Write([]byte("Method not allowed"))
			return
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			w.l.Error().Err(err).Msg("error reading body")
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte("Failed to read request body"))
			return
		}

		expected := r.Header.Get("X-Hatchet-Signature")
		actual, err := signature.Sign(string(data), opts.Secret)
		if err != nil {
			w.l.Error().Err(err).Msg("error signing request")
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte("Failed to verify request signature"))
			return
		}

		if expected != actual {
			w.l.Error().Err(fmt.Errorf("expected signature %s, got %s", expected, actual)).Msg("error in request signature")
			writer.WriteHeader(http.StatusUnauthorized)
			_, _ = writer.Write([]byte("Invalid request signature"))
			return
		}

		if r.Method == http.MethodPut {
			for _, wf := range workflows {
				err = w.RegisterWorkflow(wf)
				if err != nil {
					w.l.Error().Err(err).Msg("error registering workflow")
					writer.WriteHeader(http.StatusInternalServerError)
					_, _ = writer.Write([]byte("Failed to register workflow"))
					return
				}
			}

			var actions []string
			for _, action := range w.actions {
				actions = append(actions, action.Name())
			}

			res := HealthCheckResponse{
				Actions: actions,
			}
			data, err := json.Marshal(res)
			if err != nil {
				w.l.Error().Err(err).Msg("error marshalling health check response")
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = writer.Write([]byte("Internal server error"))
				return
			}
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write(data)
			return
		}

		var action ActionPayload
		if err := json.Unmarshal(data, &action); err != nil {
			w.l.Error().Err(err).Msg("error unmarshalling action")
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte("Invalid request format"))
			return
		}

		actionWithPayload := action.Action
		actionWithPayload.ActionPayload = []byte(action.ActionPayload)

		timestamp := time.Now().UTC()
		_, err = w.client.Dispatcher().SendStepActionEvent(r.Context(),
			&client.ActionEvent{
				Action:         actionWithPayload,
				EventTimestamp: &timestamp,
				EventType:      client.ActionEventTypeStarted,
			},
		)
		if err != nil {
			w.l.Error().Err(err).Msg("error dispatching event")
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte("Failed to dispatch action event"))
			return
		}

		ctx, err := newHatchetContext(r.Context(), actionWithPayload, w.client, w.l, w)
		if err != nil {
			w.l.Error().Err(err).Msg("error creating context")
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte("Failed to process action context"))
			return
		}
		resp, err := w.webhookProcess(ctx)
		if err != nil {
			// FIXME handle error gracefully and send a failed event
			w.l.Error().Err(err).Msg("error processing request")
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte("Failed to process webhook"))
			return
		}

		timestamp = time.Now().UTC()
		_, err = w.client.Dispatcher().SendStepActionEvent(r.Context(),
			&client.ActionEvent{
				Action:         actionWithPayload,
				EventTimestamp: &timestamp,
				EventType:      client.ActionEventTypeCompleted,
				EventPayload:   resp,
			},
		)
		if err != nil {
			w.l.Error().Err(err).Msg("error dispatching event")
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte("Failed to dispatch completion event"))
			return
		}

		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("OK"))
	}
}

func (w *Worker) webhookProcess(ctx HatchetContext) (interface{}, error) {
	var do Action
	for _, action := range w.actions {
		split := strings.Split(action.Name(), ":") // service:action
		if split[1] == ctx.StepName() {
			do = action
			break
		}
	}

	if do == nil {
		return nil, fmt.Errorf("fatal: could not find action for step run %s", ctx.StepName())
	}

	res := do.Run(ctx)

	if len(res) != 2 {
		return nil, fmt.Errorf("invalid response from action, expected 2 values")
	}

	if res[1] != nil {
		err, ok := res[1].(error)
		if !ok {
			return nil, fmt.Errorf("invalid response from action, expected error")
		}

		if err != nil {
			return nil, err
		}
	}

	return res[0], nil
}