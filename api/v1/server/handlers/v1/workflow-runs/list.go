package workflowruns

import (
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/hatchet-dev/hatchet/api/v1/server/oas/gen"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/dbsqlc"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/sqlchelpers"
	v1 "github.com/hatchet-dev/hatchet/pkg/repository/v1"
	"github.com/hatchet-dev/hatchet/pkg/repository/v1/sqlcv1"

	transformers "github.com/hatchet-dev/hatchet/api/v1/server/oas/transformers/v1"
)

// isValidMetadataKeyValue validates a key-value pair for metadata.
// It ensures keys are alphanumeric with underscores and dashes, and values are not empty and have a reasonable length.
func isValidMetadataKeyValue(key, value string) bool {
    // Validate key: only allow alphanumeric, underscore, and dash
    if key == "" {
        return false
    }
    
    for _, r := range key {
        if !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9' || r == '_' || r == '-') {
            return false
        }
    }
    
    // Validate value: not empty, reasonable length
    return value != "" && len(value) <= 1000  // 1000 is an arbitrary limit
}

func (t *V1WorkflowRunsService) WithDags(ctx echo.Context, request gen.V1WorkflowRunListRequestObject) (gen.V1WorkflowRunListResponseObject, error) {
	tenant := ctx.Get("tenant").(*dbsqlc.Tenant)
	tenantId := sqlchelpers.UUIDToStr(tenant.ID)

	var (
		statuses = []sqlcv1.V1ReadableStatusOlap{
			sqlcv1.V1ReadableStatusOlapQUEUED,
			sqlcv1.V1ReadableStatusOlapRUNNING,
			sqlcv1.V1ReadableStatusOlapFAILED,
			sqlcv1.V1ReadableStatusOlapCOMPLETED,
			sqlcv1.V1ReadableStatusOlapCANCELLED,
		}
		since             = request.Params.Since
		workflowIds       = []uuid.UUID{}
		limit       int64 = 50
		offset      int64
	)

	if request.Params.Statuses != nil {
		if len(*request.Params.Statuses) > 0 {
			statuses = []sqlcv1.V1ReadableStatusOlap{}
			for _, status := range *request.Params.Statuses {
				statuses = append(statuses, sqlcv1.V1ReadableStatusOlap(status))
			}
		}
	}

	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}

	if request.Params.Offset != nil {
		offset = *request.Params.Offset
	}

	if request.Params.WorkflowIds != nil {
		workflowIds = *request.Params.WorkflowIds
	}

	opts := v1.ListWorkflowRunOpts{
		CreatedAfter: since,
		Statuses:     statuses,
		WorkflowIds:  workflowIds,
		Limit:        limit,
		Offset:       offset,
	}

	additionalMetadataFilters := make(map[string]interface{})

	if request.Params.AdditionalMetadata != nil {
		for _, v := range *request.Params.AdditionalMetadata {
			kv_pairs := strings.Split(v, ":")
			if len(kv_pairs) == 2 {
				key := strings.TrimSpace(kv_pairs[0])
				value := strings.TrimSpace(kv_pairs[1])
				
				if isValidMetadataKeyValue(key, value) {
					additionalMetadataFilters[key] = value
				}
			}
		}

		opts.AdditionalMetadata = additionalMetadataFilters
	}

	if request.Params.Until != nil {
		opts.FinishedBefore = request.Params.Until
	}

	if request.Params.ParentTaskExternalId != nil {
		parentTaskExternalId := request.Params.ParentTaskExternalId.String()
		id := sqlchelpers.UUIDFromStr(parentTaskExternalId)
		opts.ParentTaskExternalId = &id
	}

	dags, total, err := t.config.V1.OLAP().ListWorkflowRuns(
		ctx.Request().Context(),
		tenantId,
		opts,
	)

	if err != nil {
		return nil, err
	}

	dagExternalIds := make([]pgtype.UUID, 0)

	for _, dag := range dags {
		if dag.Kind == sqlcv1.V1RunKindDAG {
			dagExternalIds = append(dagExternalIds, dag.ExternalID)
		}
	}

	tasks, taskIdToDagExternalId, err := t.config.V1.OLAP().ListTasksByDAGId(
		ctx.Request().Context(),
		tenantId,
		dagExternalIds,
	)

	if err != nil {
		return nil, err
	}

	pgWorkflowIds := make([]pgtype.UUID, 0)

	for _, wf := range dags {
		pgWorkflowIds = append(pgWorkflowIds, wf.WorkflowID)
	}

	workflowNames, err := t.config.V1.Workflows().ListWorkflowNamesByIds(
		ctx.Request().Context(),
		tenantId,
		pgWorkflowIds,
	)

	if err != nil {
		return nil, err
	}

	taskIdToWorkflowName := make(map[int64]string)

	for _, task := range tasks {
		if name, ok := workflowNames[task.WorkflowID]; ok {
			taskIdToWorkflowName[task.ID] = name
		}
	}

	parsedTasks := transformers.TaskRunDataRowToWorkflowRunsMany(tasks, taskIdToWorkflowName, total, limit, offset)

	dagChildren := make(map[uuid.UUID][]gen.V1TaskSummary)

	for _, task := range parsedTasks.Rows {
		dagExternalId := taskIdToDagExternalId[int64(task.TaskId)]
		existing, ok := dagChildren[dagExternalId]

		if ok {
			dagChildren[dagExternalId] = append(existing, task)
		} else {
			dagChildren[dagExternalId] = []gen.V1TaskSummary{task}
		}
	}

	result := transformers.ToWorkflowRunMany(dags, dagChildren, workflowNames, total, limit, offset)

	// Search for api errors to see how we handle errors in other cases
	return gen.V1WorkflowRunList200JSONResponse(
		result,
	), nil
}

func (t *V1WorkflowRunsService) OnlyTasks(ctx echo.Context, request gen.V1WorkflowRunListRequestObject) (gen.V1WorkflowRunListResponseObject, error) {
	tenant := ctx.Get("tenant").(*dbsqlc.Tenant)
	tenantId := sqlchelpers.UUIDToStr(tenant.ID)

	var (
		statuses = []sqlcv1.V1ReadableStatusOlap{
			sqlcv1.V1ReadableStatusOlapQUEUED,
			sqlcv1.V1ReadableStatusOlapRUNNING,
			sqlcv1.V1ReadableStatusOlapFAILED,
			sqlcv1.V1ReadableStatusOlapCOMPLETED,
			sqlcv1.V1ReadableStatusOlapCANCELLED,
		}
		since             = request.Params.Since
		workflowIds       = []uuid.UUID{}
		limit       int64 = 50
		offset      int64
	)

	if request.Params.Statuses != nil {
		if len(*request.Params.Statuses) > 0 {
			statuses = []sqlcv1.V1ReadableStatusOlap{}
			for _, status := range *request.Params.Statuses {
				statuses = append(statuses, sqlcv1.V1ReadableStatusOlap(status))
			}
		}
	}

	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}

	if request.Params.Offset != nil {
		offset = *request.Params.Offset
	}

	if request.Params.WorkflowIds != nil {
		workflowIds = *request.Params.WorkflowIds
	}

	opts := v1.ListTaskRunOpts{
		CreatedAfter: since,
		Statuses:     statuses,
		WorkflowIds:  workflowIds,
		Limit:        limit,
		Offset:       offset,
		WorkerId:     request.Params.WorkerId,
	}

	additionalMetadataFilters := make(map[string]interface{})

	if request.Params.AdditionalMetadata != nil {
		for _, v := range *request.Params.AdditionalMetadata {
			kv_pairs := strings.Split(v, ":")
			if len(kv_pairs) == 2 {
				key := strings.TrimSpace(kv_pairs[0])
				value := strings.TrimSpace(kv_pairs[1])
				
				if isValidMetadataKeyValue(key, value) {
					additionalMetadataFilters[key] = value
				}
			}
		}

		opts.AdditionalMetadata = additionalMetadataFilters
	}

	if request.Params.Until != nil {
		opts.FinishedBefore = request.Params.Until
	}

	tasks, total, err := t.config.V1.OLAP().ListTasks(
		ctx.Request().Context(),
		tenantId,
		opts,
	)

	if err != nil {
		return nil, err
	}

	taskIdToWorkflowName := make(map[int64]string)

	result := transformers.TaskRunDataRowToWorkflowRunsMany(tasks, taskIdToWorkflowName, total, limit, offset)

	// Search for api errors to see how we handle errors in other cases
	return gen.V1WorkflowRunList200JSONResponse(
		result,
	), nil
}

func (t *V1WorkflowRunsService) V1WorkflowRunList(ctx echo.Context, request gen.V1WorkflowRunListRequestObject) (gen.V1WorkflowRunListResponseObject, error) {
	if request.Params.OnlyTasks {
		return t.OnlyTasks(ctx, request)
	} else {
		return t.WithDags(ctx, request)
	}
}

func (t *V1WorkflowRunsService) V1WorkflowRunDisplayNamesList(ctx echo.Context, request gen.V1WorkflowRunDisplayNamesListRequestObject) (gen.V1WorkflowRunDisplayNamesListResponseObject, error) {
	tenant := ctx.Get("tenant").(*dbsqlc.Tenant)

	externalIds := make([]pgtype.UUID, len(request.Params.ExternalIds))

	for i, id := range request.Params.ExternalIds {
		externalIds[i] = sqlchelpers.UUIDFromStr(id.String())
	}

	displayNames, err := t.config.V1.OLAP().ListWorkflowRunDisplayNames(
		ctx.Request().Context(),
		tenant.ID,
		externalIds,
	)

	if err != nil {
		return nil, err
	}

	result := transformers.ToWorkflowRunDisplayNamesList(displayNames)

	return gen.V1WorkflowRunDisplayNamesList200JSONResponse(
		result,
	), nil
}