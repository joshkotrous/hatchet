package workflows

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/hatchet-dev/hatchet/api/v1/server/oas/apierrors"
	"github.com/hatchet-dev/hatchet/api/v1/server/oas/gen"
	"github.com/hatchet-dev/hatchet/api/v1/server/oas/transformers"
	"github.com/hatchet-dev/hatchet/pkg/repository"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/dbsqlc"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/sqlchelpers"
)

// metadataKeyRegex validates that metadata keys only contain alphanumeric characters,
// hyphens, underscores, and periods.
var metadataKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_.]+$`)

func (t *WorkflowService) WorkflowRunList(ctx echo.Context, request gen.WorkflowRunListRequestObject) (gen.WorkflowRunListResponseObject, error) {
	tenant := ctx.Get("tenant").(*dbsqlc.Tenant)
	tenantId := sqlchelpers.UUIDToStr(tenant.ID)

	limit := 50
	offset := 0
	orderDirection := "DESC"
	orderBy := "createdAt"

	listOpts := &repository.ListWorkflowRunsOpts{
		Limit:          &limit,
		Offset:         &offset,
		OrderBy:        &orderBy,
		OrderDirection: &orderDirection,
	}

	// Define allowed values for orderBy and orderDirection
	allowedOrderByFields := map[string]bool{
		"createdAt":   true,
		"updatedAt":   true,
		"startedAt":   true,
		"finishedAt":  true,
		"status":      true,
		"workflowId":  true,
		"eventId":     true,
		"tenantId":    true,
		"displayName": true,
	}

	allowedOrderDirections := map[string]bool{
		"ASC":  true,
		"DESC": true,
	}

	if request.Params.OrderByField != nil {
		orderByValue := string(*request.Params.OrderByField)
		if allowedOrderByFields[orderByValue] {
			orderBy = orderByValue
			listOpts.OrderBy = &orderBy
		} else {
			return gen.WorkflowRunList400JSONResponse(apierrors.NewAPIErrors("Invalid orderBy field provided.")), nil
		}
	}

	if request.Params.OrderByDirection != nil {
		orderDirectionValue := strings.ToUpper(string(*request.Params.OrderByDirection))
		if allowedOrderDirections[orderDirectionValue] {
			orderDirection = orderDirectionValue
			listOpts.OrderDirection = &orderDirection
		} else {
			return gen.WorkflowRunList400JSONResponse(apierrors.NewAPIErrors("Invalid orderDirection. Allowed values: ASC, DESC.")), nil
		}
	}

	if request.Params.CreatedAfter != nil {
		listOpts.CreatedAfter = request.Params.CreatedAfter
	}

	if request.Params.CreatedBefore != nil {
		listOpts.CreatedBefore = request.Params.CreatedBefore
	}

	if request.Params.FinishedAfter != nil {
		listOpts.FinishedAfter = request.Params.FinishedAfter
	}

	if request.Params.FinishedBefore != nil {
		listOpts.FinishedBefore = request.Params.FinishedBefore
	}

	if request.Params.Limit != nil {
		limit = int(*request.Params.Limit)
		listOpts.Limit = &limit
	}

	if request.Params.Offset != nil {
		offset = int(*request.Params.Offset)
		listOpts.Offset = &offset
	}

	if request.Params.WorkflowId != nil {
		workflowIdStr := request.Params.WorkflowId.String()
		listOpts.WorkflowId = &workflowIdStr
	}

	if request.Params.EventId != nil {
		eventIdStr := request.Params.EventId.String()
		listOpts.EventId = &eventIdStr
	}

	if request.Params.ParentWorkflowRunId != nil {
		parentWorkflowRunIdStr := request.Params.ParentWorkflowRunId.String()
		listOpts.ParentId = &parentWorkflowRunIdStr
	}

	if request.Params.ParentStepRunId != nil {
		parentStepRunIdStr := request.Params.ParentStepRunId.String()
		listOpts.ParentStepRunId = &parentStepRunIdStr
	}

	if request.Params.Statuses != nil {
		statuses := make([]dbsqlc.WorkflowRunStatus, len(*request.Params.Statuses))

		for i, status := range *request.Params.Statuses {
			statuses[i] = dbsqlc.WorkflowRunStatus(status)
		}

		listOpts.Statuses = &statuses
	}

	if request.Params.AdditionalMetadata != nil {
		additionalMetadata := make(map[string]interface{}, len(*request.Params.AdditionalMetadata))

		for _, v := range *request.Params.AdditionalMetadata {
			splitValue := strings.SplitN(fmt.Sprintf("%v", v), ":", 2)

			if len(splitValue) != 2 {
				return gen.WorkflowRunList400JSONResponse(apierrors.NewAPIErrors("Additional metadata filters must be in the format key:value.")), nil
			}

			key := strings.TrimSpace(splitValue[0])
			value := strings.TrimSpace(splitValue[1])

			// Validate key format using regex
			if !metadataKeyRegex.MatchString(key) {
				return gen.WorkflowRunList400JSONResponse(apierrors.NewAPIErrors("Metadata key contains invalid characters. Only alphanumeric characters, hyphens, underscores, and periods are allowed.")), nil
			}

			// Check key length
			if len(key) > 64 {
				return gen.WorkflowRunList400JSONResponse(apierrors.NewAPIErrors("Metadata key exceeds maximum length of 64 characters.")), nil
			}

			// Check value length
			if len(value) > 1024 {
				return gen.WorkflowRunList400JSONResponse(apierrors.NewAPIErrors("Metadata value exceeds maximum length of 1024 characters.")), nil
			}

			additionalMetadata[key] = value
		}

		listOpts.AdditionalMetadata = additionalMetadata
	}

	dbCtx, cancel := context.WithTimeout(ctx.Request().Context(), 30*time.Second)
	defer cancel()

	workflowRuns, err := t.config.APIRepository.WorkflowRun().ListWorkflowRuns(dbCtx, tenantId, listOpts)

	if err != nil {
		return nil, err
	}

	rows := make([]gen.WorkflowRun, len(workflowRuns.Rows))

	for i, workflow := range workflowRuns.Rows {
		workflowCp := workflow
		rows[i] = *transformers.ToWorkflowRunFromSQLC(workflowCp)
	}

	// use the total rows and limit to calculate the total pages
	totalPages := int64(math.Ceil(float64(workflowRuns.Count) / float64(limit)))
	currPage := 1 + int64(math.Ceil(float64(offset)/float64(limit)))
	nextPage := currPage + 1

	if currPage == totalPages {
		nextPage = currPage
	}

	return gen.WorkflowRunList200JSONResponse(
		gen.WorkflowRunList{
			Rows: &rows,
			Pagination: &gen.PaginationResponse{
				NumPages:    &totalPages,
				CurrentPage: &currPage,
				NextPage:    &nextPage,
			},
		},
	), nil
}