package tenants

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/hatchet-dev/hatchet/api/v1/server/oas/apierrors"
	"github.com/hatchet-dev/hatchet/api/v1/server/oas/gen"
	"github.com/hatchet-dev/hatchet/pkg/repository"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/dbsqlc"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/sqlchelpers"
)

// validateMetadataField ensures that a metadata field contains only allowed characters
// and is within a reasonable length
func validateMetadataField(field string, isKey bool) error {
	// Maximum length for keys and values
	maxKeyLength := 50
	maxValueLength := 256
	
	// Check length based on whether it's a key or value
	if isKey && len(field) > maxKeyLength {
		return fmt.Errorf("metadata key exceeds maximum length of %d characters", maxKeyLength)
	} else if !isKey && len(field) > maxValueLength {
		return fmt.Errorf("metadata value exceeds maximum length of %d characters", maxValueLength)
	}

	// Regex pattern for allowed characters
	// For keys: alphanumeric plus underscore and dash
	// For values: more permissive but still restricted
	var validPattern *regexp.Regexp
	if isKey {
		validPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	} else {
		// Allow a broader range of characters for values, but still restrict potentially dangerous ones
		validPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-\.,\s]+$`)
	}

	if !validPattern.MatchString(field) {
		if isKey {
			return fmt.Errorf("metadata key contains invalid characters (only alphanumeric, underscore, and dash are allowed)")
		} else {
			return fmt.Errorf("metadata value contains invalid characters")
		}
	}

	return nil
}

func (t *TenantService) TenantGetQueueMetrics(ctx echo.Context, request gen.TenantGetQueueMetricsRequestObject) (gen.TenantGetQueueMetricsResponseObject, error) {
	tenant := ctx.Get("tenant").(*dbsqlc.Tenant)
	tenantId := sqlchelpers.UUIDToStr(tenant.ID)

	opts := repository.GetQueueMetricsOpts{}

	if request.Params.AdditionalMetadata != nil {
		additionalMetadata := make(map[string]interface{}, len(*request.Params.AdditionalMetadata))

		for _, v := range *request.Params.AdditionalMetadata {
			splitValue := strings.Split(fmt.Sprintf("%v", v), ":")

			if len(splitValue) == 2 {
				// Validate both key and value
				if err := validateMetadataField(splitValue[0], true); err != nil {
					return gen.TenantGetQueueMetrics400JSONResponse(apierrors.NewAPIErrors(fmt.Sprintf("Invalid metadata key: %s", err.Error()))), nil
				}
				
				if err := validateMetadataField(splitValue[1], false); err != nil {
					return gen.TenantGetQueueMetrics400JSONResponse(apierrors.NewAPIErrors(fmt.Sprintf("Invalid metadata value: %s", err.Error()))), nil
				}
				
				additionalMetadata[splitValue[0]] = splitValue[1]
			} else {
				return gen.TenantGetQueueMetrics400JSONResponse(apierrors.NewAPIErrors("Additional metadata filters must be in the format key:value.")), nil

			}
		}

		opts.AdditionalMetadata = additionalMetadata
	}

	if request.Params.Workflows != nil {
		opts.WorkflowIds = *request.Params.Workflows
	}

	metrics, err := t.config.APIRepository.Tenant().GetQueueMetrics(ctx.Request().Context(), tenantId, &opts)

	if err != nil {
		return nil, err
	}

	stepRunQueueCounts, err := t.config.EngineRepository.StepRun().GetQueueCounts(ctx.Request().Context(), tenantId)

	if err != nil {
		return nil, err
	}

	respWorkflowMap := make(map[string]gen.QueueMetrics, len(metrics.ByWorkflowId))

	for k, v := range metrics.ByWorkflowId {
		respWorkflowMap[k] = gen.QueueMetrics{
			NumPending: v.Pending,
			NumQueued:  v.PendingAssignment,
			NumRunning: v.Running,
		}
	}

	resp := gen.TenantQueueMetrics{
		Total: &gen.QueueMetrics{
			NumPending: metrics.Total.Pending,
			NumQueued:  metrics.Total.PendingAssignment,
			NumRunning: metrics.Total.Running,
		},
		Workflow: &respWorkflowMap,
		Queues:   &stepRunQueueCounts,
	}

	return gen.TenantGetQueueMetrics200JSONResponse(resp), nil
}