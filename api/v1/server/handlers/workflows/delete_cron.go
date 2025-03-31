package workflows

import (
	"context"
	"errors"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/hatchet-dev/hatchet/api/v1/server/oas/gen"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/dbsqlc"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/sqlchelpers"
)

func (t *WorkflowService) WorkflowCronDelete(ctx echo.Context, request gen.WorkflowCronDeleteRequestObject) (gen.WorkflowCronDeleteResponseObject, error) {
	tenantVal := ctx.Get("tenant")
	if tenantVal == nil {
		return nil, errors.New("tenant not found in context")
	}
	
	_, ok := tenantVal.(*dbsqlc.Tenant)
	if !ok {
		return nil, errors.New("invalid tenant type in context")
	}
	
	cronVal := ctx.Get("cron-workflow")
	if cronVal == nil {
		return nil, errors.New("cron-workflow not found in context")
	}
	
	cron, ok := cronVal.(*dbsqlc.ListCronWorkflowsRow)
	if !ok {
		return nil, errors.New("invalid cron-workflow type in context")
	}

	dbCtx, cancel := context.WithTimeout(ctx.Request().Context(), 30*time.Second)
	defer cancel()

	err := t.config.APIRepository.Workflow().DeleteCronWorkflow(dbCtx,
		sqlchelpers.UUIDToStr(cron.TenantId),
		sqlchelpers.UUIDToStr(cron.CronId),
	)

	if err != nil {
		return nil, err
	}

	return gen.WorkflowCronDelete204Response{}, nil
}