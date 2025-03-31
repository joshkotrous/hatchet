package logs

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/hatchet-dev/hatchet/api/v1/server/oas/gen"
	"github.com/hatchet-dev/hatchet/api/v1/server/oas/transformers"
	"github.com/hatchet-dev/hatchet/pkg/repository"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/dbsqlc"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/sqlchelpers"
)

// validateSearchParameter checks if the search parameter contains potentially dangerous SQL patterns
func validateSearchParameter(search string) error {
	// Pattern to catch common SQL injection attempts
	dangerousPattern := `(?i)(--|;|\/\*|\*\/|@@|@|\bAND\b|\bOR\b|\bUNION\b|\bSELECT\b|\bFROM\b|\bWHERE\b|\bINSERT\b|\bUPDATE\b|\bDELETE\b|\bDROP\b|\bCREATE\b|\bALTER\b|\bTRUNCATE\b)`
	
	matched, err := regexp.MatchString(dangerousPattern, search)
	if err != nil {
		return err
	}
	
	if matched {
		return fmt.Errorf("search parameter contains potentially malicious patterns")
	}
	
	return nil
}

func (t *LogService) LogLineList(ctx echo.Context, request gen.LogLineListRequestObject) (gen.LogLineListResponseObject, error) {
	tenant := ctx.Get("tenant").(*dbsqlc.Tenant)
	tenantId := sqlchelpers.UUIDToStr(tenant.ID)
	stepRun := ctx.Get("step-run").(*repository.GetStepRunFull)

	limit := 1000
	offset := 0

	stepRunId := sqlchelpers.UUIDToStr(stepRun.ID)

	listOpts := &repository.ListLogsOpts{
		Limit:     &limit,
		Offset:    &offset,
		StepRunId: &stepRunId,
	}

	if request.Params.Search != nil {
		if err := validateSearchParameter(*request.Params.Search); err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		listOpts.Search = request.Params.Search
	}

	if request.Params.Levels != nil {
		levels := make([]string, len(*request.Params.Levels))

		for i, level := range *request.Params.Levels {
			levels[i] = string(level)
		}

		listOpts.Levels = levels
	}

	// Define allowed order by fields
	allowedOrderByFields := map[string]bool{
		"timestamp": true,
		"level":     true,
		"message":   true,
		// Add other valid fields as needed based on your database schema
	}

	// Define allowed order directions
	allowedOrderDirections := map[string]bool{
		"ASC":  true,
		"DESC": true,
	}

	if request.Params.OrderByField != nil {
		fieldValue := string(*request.Params.OrderByField)
		if allowedOrderByFields[fieldValue] {
			listOpts.OrderBy = repository.StringPtr(fieldValue)
		}
		// Silently ignore invalid fields
	}

	if request.Params.OrderByDirection != nil {
		directionValue := strings.ToUpper(string(*request.Params.OrderByDirection))
		if allowedOrderDirections[directionValue] {
			listOpts.OrderDirection = repository.StringPtr(directionValue)
		}
		// Silently ignore invalid directions
	}

	if request.Params.Limit != nil {
		limit = int(*request.Params.Limit)
		listOpts.Limit = &limit
	}

	if request.Params.Offset != nil {
		offset = int(*request.Params.Offset)
		listOpts.Offset = &offset
	}

	listRes, err := t.config.APIRepository.Log().ListLogLines(tenantId, listOpts)

	if err != nil {
		return nil, err
	}

	rows := make([]gen.LogLine, len(listRes.Rows))

	for i, log := range listRes.Rows {
		rows[i] = *transformers.ToLogFromSQLC(log)
	}

	// use the total rows and limit to calculate the total pages
	totalPages := int64(math.Ceil(float64(listRes.Count) / float64(limit)))
	currPage := 1 + int64(math.Ceil(float64(offset)/float64(limit)))
	nextPage := currPage + 1

	if currPage == totalPages {
		nextPage = currPage
	}

	return gen.LogLineList200JSONResponse(
		gen.LogLineList{
			Rows: &rows,
			Pagination: &gen.PaginationResponse{
				NumPages:    &totalPages,
				NextPage:    &nextPage,
				CurrentPage: &currPage,
			},
		},
	), nil
}