package worker

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/hatchet-dev/hatchet/internal/whrequest"
	"github.com/hatchet-dev/hatchet/pkg/client"
	"github.com/hatchet-dev/hatchet/pkg/client/rest"
)

type RegisterWebhookWorkerOpts struct {
	Name   string
	URL    string
	Secret *string
}

type ActionPayload struct {
	*client.Action

	ActionPayload string `json:"actionPayload"`
}

func (w *Worker) RegisterWebhook(ww RegisterWebhookWorkerOpts) error {
	// Validate the URL before registering the webhook
	if err := isURLAllowed(ww.URL); err != nil {
		return fmt.Errorf("invalid webhook URL during registration: %w", err)
	}

	tenantId := openapi_types.UUID{}
	if err := tenantId.Scan(w.client.TenantId()); err != nil {
		return fmt.Errorf("error getting tenant id: %w", err)
	}

	res, err := w.client.API().WebhookCreate(context.Background(), tenantId, rest.WebhookCreateJSONRequestBody{
		Url:    ww.URL,
		Secret: ww.Secret,
		Name:   ww.Name,
	})
	if err != nil {
		return fmt.Errorf("error creating webhook worker: %w", err)
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("error creating webhook, failed with status code %d", res.StatusCode)
	}

	return nil
}

type WebhookWorkerOpts struct {
	URL       string
	Secret    string
	WebhookId string
}

// FIXME do not expose this to the end-user client somehow
func (w *Worker) StartWebhook(ww WebhookWorkerOpts) (func() error, error) {
	// Validate the URL when starting the webhook worker
	if err := isURLAllowed(ww.URL); err != nil {
		return nil, fmt.Errorf("invalid webhook URL when starting webhook: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	listener, _, err := w.client.Dispatcher().GetActionListener(ctx, &client.GetActionListenerRequest{
		WorkerName: w.name,
		Actions:    w.initActionNames,
		MaxRuns:    w.maxRuns,
		WebhookId:  &ww.WebhookId,
	})

	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not get action listener: %w", err)
	}

	actionCh, errCh, err := listener.Actions(ctx)

	if err != nil {
		cancel()
		return nil, fmt.Errorf("could not get action channel: %w", err)
	}

	go func() {
		for {
			select {
			case err := <-errCh:
				// NOTE: this matches the behavior of the old worker, until we change the signature of the webhook workers
				panic(err)
			case action := <-actionCh:
				go func(action *client.Action) {
					err := w.sendWebhook(context.Background(), action, ww)

					if err != nil {
						w.l.Error().Err(err).Msgf("could not execute action: %s", action.ActionId)
					}

					w.l.Debug().Msgf("action %s completed", action.ActionId)
				}(action)
			case <-ctx.Done():
				w.l.Debug().Msgf("worker %s received context done, stopping", w.name)
				return
			}
		}
	}()

	cleanup := func() error {
		cancel()

		w.l.Debug().Msgf("worker %s stopped", w.name)

		return nil
	}

	return cleanup, nil
}

// isURLAllowed checks if a URL is valid and allowed to be called
func isURLAllowed(urlStr string) error {
	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	
	// Check for http or https scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}
	
	// Check for empty host
	if parsedURL.Host == "" {
		return fmt.Errorf("URL host cannot be empty")
	}
	
	// Extract hostname without port
	hostname := parsedURL.Hostname()
	
	// Check for localhost, loopback, or unspecified addresses
	lowerHost := strings.ToLower(hostname)
	if lowerHost == "localhost" || 
		strings.HasPrefix(lowerHost, "127.") || 
		lowerHost == "::1" ||
		lowerHost == "0.0.0.0" {
		return fmt.Errorf("URL cannot point to localhost or loopback addresses")
	}
	
	// Parse hostname as IP if possible
	ip := net.ParseIP(hostname)
	if ip != nil {
		// Check for private IPv4 ranges
		if ip4 := ip.To4(); ip4 != nil {
			// Check for private IPv4 ranges
			// 10.0.0.0/8
			if ip4[0] == 10 {
				return fmt.Errorf("URL cannot point to private IP ranges (10.0.0.0/8)")
			}
			// 172.16.0.0/12
			if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
				return fmt.Errorf("URL cannot point to private IP ranges (172.16.0.0/12)")
			}
			// 192.168.0.0/16
			if ip4[0] == 192 && ip4[1] == 168 {
				return fmt.Errorf("URL cannot point to private IP ranges (192.168.0.0/16)")
			}
			// 169.254.0.0/16 (link-local)
			if ip4[0] == 169 && ip4[1] == 254 {
				return fmt.Errorf("URL cannot point to link-local IP ranges (169.254.0.0/16)")
			}
		} else {
			// Check for private IPv6 ranges
			// fd00::/8 (private)
			if len(ip) == 16 && ip[0] == 0xfd {
				return fmt.Errorf("URL cannot point to private IPv6 ranges (fd00::/8)")
			}
		}
	}
	
	return nil
}

func (w *Worker) sendWebhook(ctx context.Context, action *client.Action, ww WebhookWorkerOpts) error {
	w.l.Debug().Msgf("action received from step run %s, sending webhook at %s", action.StepRunId, time.Now())

	// Validate the URL before sending the webhook
	if err := isURLAllowed(ww.URL); err != nil {
		errMsg := fmt.Errorf("invalid webhook URL: %w", err)
		w.l.Warn().Msgf("step run %s has invalid webhook URL %s: %s", action.StepRunId, ww.URL, errMsg)
		if markErr := w.markFailed(action, errMsg); markErr != nil {
			return fmt.Errorf("invalid webhook URL and then could not send failed action event: %w", markErr)
		}
		return errMsg
	}

	actionWithPayload := ActionPayload{
		Action:        action,
		ActionPayload: string(action.ActionPayload),
	}

	_, statusCode, err := whrequest.Send(ctx, ww.URL, ww.Secret, actionWithPayload)

	if statusCode != nil && *statusCode != 200 {
		w.l.Debug().Msgf("step run %s webhook sent with status code %d", action.StepRunId, *statusCode)
	}

	if err != nil {
		w.l.Warn().Msgf("step run %s could not send webhook to %s: %s", action.StepRunId, ww.URL, err)
		if err := w.markFailed(action, fmt.Errorf("could not send webhook: %w", err)); err != nil {
			return fmt.Errorf("could not send webhook and then could not send failed action event: %w", err)
		}
		return err
	}

	return nil
}

func (w *Worker) markFailed(action *client.Action, err error) error {
	failureEvent := w.getActionEvent(action, client.ActionEventTypeFailed)

	w.alerter.SendAlert(context.Background(), err, map[string]interface{}{
		"actionId":      action.ActionId,
		"workerId":      action.WorkerId,
		"workflowRunId": action.WorkflowRunId,
		"jobName":       action.JobName,
		"actionType":    action.ActionType,
	})

	failureEvent.EventPayload = err.Error()

	_, err = w.client.Dispatcher().SendStepActionEvent(
		context.Background(),
		failureEvent,
	)

	if err != nil {
		return fmt.Errorf("could not send action event: %w", err)
	}

	return nil
}