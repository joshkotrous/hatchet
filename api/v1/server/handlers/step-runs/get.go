package stepruns

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/hatchet-dev/hatchet/api/v1/server/oas/gen"
	"github.com/hatchet-dev/hatchet/api/v1/server/oas/transformers"
	"github.com/hatchet-dev/hatchet/pkg/repository"
)

func (t *StepRunService) StepRunGet(ctx echo.Context, request gen.StepRunGetRequestObject) (gen.StepRunGetResponseObject, error) {
	// Safely retrieve and validate the step-run from context
	valueFromContext := ctx.Get("step-run")
	if valueFromContext == nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "step-run not found in context")
	}

	stepRun, ok := valueFromContext.(*repository.GetStepRunFull)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "invalid type for step-run in context")
	}

	return gen.StepRunGet200JSONResponse(
		*transformers.ToStepRunFull(stepRun),
	), nil
}