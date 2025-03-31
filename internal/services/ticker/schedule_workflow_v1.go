package ticker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	msgqueuev1 "github.com/hatchet-dev/hatchet/internal/msgqueue/v1"
	tasktypes "github.com/hatchet-dev/hatchet/internal/services/shared/tasktypes/v1"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/dbsqlc"
	v1 "github.com/hatchet-dev/hatchet/pkg/repository/v1"
)

// validateAndSanitizeInput validates the input JSON and performs basic sanitization.
// It returns the sanitized input or an error if validation fails.
func validateAndSanitizeInput(input json.RawMessage, maxSize int) (json.RawMessage, error) {
	if input == nil {
		return nil, nil // Empty input is valid
	}

	// Check size
	if len(input) > maxSize {
		return nil, fmt.Errorf("input size exceeds maximum allowed (%d bytes)", maxSize)
	}

	// Validate and sanitize by parsing and re-serializing
	var parsedData interface{}
	if err := json.Unmarshal(input, &parsedData); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Re-serialize to create a sanitized copy
	sanitized, err := json.Marshal(parsedData)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize JSON: %w", err)
	}

	return sanitized, nil
}

func (t *TickerImpl) runScheduledWorkflowV1(ctx context.Context, tenantId string, workflowVersion *dbsqlc.GetWorkflowVersionForEngineRow, scheduledWorkflowId string, scheduled *dbsqlc.PollScheduledWorkflowsRow) error {
	// Define maximum allowed size for input and metadata
	const maxInputSize = 1024 * 1024 // 1MB
	const maxMetadataSize = 1024 * 1024 // 1MB

	// Validate and sanitize input data
	sanitizedInput, err := validateAndSanitizeInput(scheduled.Input, maxInputSize)
	if err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	// Validate and sanitize additional metadata
	sanitizedMetadata, err := validateAndSanitizeInput(scheduled.AdditionalMetadata, maxMetadataSize)
	if err != nil {
		return fmt.Errorf("additional metadata validation failed: %w", err)
	}

	// send workflow run to task controller
	opt := &v1.WorkflowNameTriggerOpts{
		TriggerTaskData: &v1.TriggerTaskData{
			WorkflowName:       workflowVersion.WorkflowName,
			Data:               sanitizedInput,
			AdditionalMetadata: sanitizedMetadata,
		},
		ExternalId: uuid.NewString(),
		ShouldSkip: false,
	}

	msg, err := tasktypes.TriggerTaskMessage(
		tenantId,
		opt,
	)

	if err != nil {
		return fmt.Errorf("could not create trigger task message: %w", err)
	}

	err = t.mqv1.SendMessage(ctx, msgqueuev1.TASK_PROCESSING_QUEUE, msg)

	if err != nil {
		return fmt.Errorf("could not send message to task queue: %w", err)
	}

	// delete the scheduled workflow
	return t.repo.WorkflowRun().DeleteScheduledWorkflow(ctx, tenantId, scheduledWorkflowId)
}