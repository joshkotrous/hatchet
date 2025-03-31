package tasks

import (
	"github.com/labstack/echo/v4"

	"github.com/hatchet-dev/hatchet/api/v1/server/oas/gen"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/dbsqlc"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/sqlchelpers"
	"github.com/hatchet-dev/hatchet/pkg/repository/v1/sqlcv1"

	transformers "github.com/hatchet-dev/hatchet/api/v1/server/oas/transformers/v1"
)

func (t *TasksService) V1TaskEventList(ctx echo.Context, request gen.V1TaskEventListRequestObject) (gen.V1TaskEventListResponseObject, error) {
	tenant := ctx.Get("tenant").(*dbsqlc.Tenant)
	tenantId := sqlchelpers.UUIDToStr(tenant.ID)
	task := ctx.Get("task").(*sqlcv1.V1TasksOlap)

	// Constants for parameter validation
	const defaultLimit = 50
	const maxLimit = 500
	const defaultOffset = 0
	const maxOffset = 10000

	// Validate and cap limit parameter
	var limit = defaultLimit
	if request.Params.Limit != nil {
		if *request.Params.Limit > 0 && *request.Params.Limit <= maxLimit {
			limit = *request.Params.Limit
		} else if *request.Params.Limit > maxLimit {
			limit = maxLimit // Cap to maximum
		}
	}

	// Validate and cap offset parameter
	var offset = defaultOffset
	if request.Params.Offset != nil {
		if *request.Params.Offset >= 0 && *request.Params.Offset <= maxOffset {
			offset = *request.Params.Offset
		} else if *request.Params.Offset > maxOffset {
			offset = maxOffset // Cap to maximum
		}
	}

	taskRunEvents, err := t.config.V1.OLAP().ListTaskRunEvents(ctx.Request().Context(), tenantId, task.ID, task.InsertedAt, limit, offset)

	if err != nil {
		return nil, err
	}

	result := transformers.ToTaskRunEventMany(taskRunEvents, sqlchelpers.UUIDToStr(task.ExternalID))

	return gen.V1TaskEventList200JSONResponse(
		result,
	), nil
}