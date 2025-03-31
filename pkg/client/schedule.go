package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/hatchet-dev/hatchet/api/v1/server/oas/gen"
	"github.com/hatchet-dev/hatchet/pkg/client/rest"
	"github.com/hatchet-dev/hatchet/pkg/validator"
)

type ScheduleOpts struct {
	// TriggerAt is the time at which the scheduled run should be triggered
	TriggerAt time.Time

	// Input is the input to the workflow
	Input map[string]interface{}

	// AdditionalMetadata is additional metadata to be stored with the cron trigger
	AdditionalMetadata map[string]string
}

type ScheduleClient interface {
	// Create creates a new scheduled workflow run
	Create(ctx context.Context, workflow string, opts *ScheduleOpts) (*gen.ScheduledWorkflows, error)

	// Delete deletes a scheduled workflow run
	Delete(ctx context.Context, id string) error

	// List lists all scheduled workflow runs
	List(ctx context.Context) (*gen.ScheduledWorkflowsList, error)
}

type scheduleClientImpl struct {
	restClient *rest.ClientWithResponses

	l *zerolog.Logger

	v validator.Validator

	tenantId uuid.UUID

	namespace string
}

func NewScheduleClient(restClient *rest.ClientWithResponses, l *zerolog.Logger, v validator.Validator, tenantId, namespace string) (ScheduleClient, error) {
	tenantIdUUID, err := uuid.Parse(tenantId)

	if err != nil {
		return nil, err
	}

	return &scheduleClientImpl{
		restClient: restClient,
		l:          l,
		v:          v,
		namespace:  namespace,
		tenantId:   tenantIdUUID,
	}, nil
}

func (c *scheduleClientImpl) Create(ctx context.Context, workflow string, opts *ScheduleOpts) (*gen.ScheduledWorkflows, error) {
	// Basic validation
	if opts == nil {
		return nil, fmt.Errorf("schedule options cannot be nil")
	}
	
	if workflow == "" {
		return nil, fmt.Errorf("workflow ID cannot be empty")
	}

	if opts.TriggerAt.IsZero() {
		return nil, fmt.Errorf("trigger time cannot be zero")
	}

	// Validate input map keys are not empty
	if opts.Input != nil {
		for k := range opts.Input {
			if k == "" {
				c.l.Warn().Msg("Empty key found in input map")
				return nil, fmt.Errorf("input map contains empty key")
			}
		}
	}

	// Validate metadata keys are not empty
	if opts.AdditionalMetadata != nil {
		for k := range opts.AdditionalMetadata {
			if k == "" {
				c.l.Warn().Msg("Empty key found in additional metadata")
				return nil, fmt.Errorf("additional metadata contains empty key")
			}
		}
	}

	// Use the validator if available
	if c.v != nil {
		type scheduleValidation struct {
			TriggerAt          time.Time              `validate:"required"`
			WorkflowID         string                 `validate:"required"`
			Input              map[string]interface{} `validate:"omitempty,dive,keys"`
			AdditionalMetadata map[string]string      `validate:"omitempty,dive,keys"`
		}

		sv := scheduleValidation{
			TriggerAt:          opts.TriggerAt,
			WorkflowID:         workflow,
			Input:              opts.Input,
			AdditionalMetadata: opts.AdditionalMetadata,
		}

		if err := c.v.Validate(sv); err != nil {
			c.l.Warn().Err(err).Msg("Input validation failed")
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	additionalMeta := make(map[string]any)

	for k, v := range opts.AdditionalMetadata {
		additionalMeta[k] = v
	}

	resp, err := c.restClient.ScheduledWorkflowRunCreate(
		ctx,
		c.tenantId,
		workflow,
		rest.ScheduleWorkflowRunRequest{
			TriggerAt:          opts.TriggerAt,
			Input:              opts.Input,
			AdditionalMetadata: additionalMeta,
		},
	)

	if err != nil {
		return nil, err
	}

	// if response code is not 200-level, return an error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// parse the response body into a scheduled workflow run
	scheduledWorkflow := &gen.ScheduledWorkflows{}

	err = json.NewDecoder(resp.Body).Decode(scheduledWorkflow)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %w", err)
	}

	return scheduledWorkflow, nil
}

func (c *scheduleClientImpl) Delete(ctx context.Context, id string) error {
	idUUID, err := uuid.Parse(id)

	if err != nil {
		return fmt.Errorf("could not parse id: %w", err)
	}

	resp, err := c.restClient.WorkflowScheduledDelete(
		ctx,
		c.tenantId,
		idUUID,
	)

	if err != nil {
		return err
	}

	// if response code is not 200-level, return an error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *scheduleClientImpl) List(ctx context.Context) (*gen.ScheduledWorkflowsList, error) {
	resp, err := c.restClient.WorkflowScheduledList(
		ctx,
		c.tenantId,
		&rest.WorkflowScheduledListParams{},
	)

	if err != nil {
		return nil, err
	}

	// parse the response body into a list of schedules
	scheduleList := &gen.ScheduledWorkflowsList{}

	err = json.NewDecoder(resp.Body).Decode(&scheduleList)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %w", err)
	}

	return scheduleList, nil
}