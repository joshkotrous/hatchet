package rate_limits

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/hatchet-dev/hatchet/api/v1/server/oas/gen"
	"github.com/hatchet-dev/hatchet/api/v1/server/oas/transformers"
	"github.com/hatchet-dev/hatchet/pkg/repository"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/dbsqlc"
	"github.com/hatchet-dev/hatchet/pkg/repository/postgres/sqlchelpers"
)

// validateSearchInput validates and sanitizes the search input
func validateSearchInput(input string) (string, error) {
    // Check for excessively long input
    if len(input) > 200 {  // Arbitrary reasonable limit
        return "", echo.NewHTTPError(400, "Search term too long")
    }
    
    // Check for potentially harmful characters or patterns
    harmfulPatterns := []string{"--", ";", "/*", "*/"}
    
    for _, pattern := range harmfulPatterns {
        if strings.Contains(input, pattern) {
            return "", echo.NewHTTPError(400, "Invalid search term")
        }
    }
    
    return input, nil
}

func (t *RateLimitService) RateLimitList(ctx echo.Context, request gen.RateLimitListRequestObject) (gen.RateLimitListResponseObject, error) {
	tenant := ctx.Get("tenant").(*dbsqlc.Tenant)
	tenantId := sqlchelpers.UUIDToStr(tenant.ID)

	limit := 50
	offset := 0

	listOpts := &repository.ListRateLimitOpts{
		Limit:  &limit,
		Offset: &offset,
	}

	if request.Params.Search != nil {
		validatedSearch, err := validateSearchInput(*request.Params.Search)
		if err != nil {
			return nil, err
		}
		listOpts.Search = &validatedSearch
	}

	// Define allowed OrderByField values - these should be actual column names in the rate_limits table
	allowedOrderByFields := map[string]bool{
		"id":         true,
		"name":       true,
		"created_at": true,
		"updated_at": true,
		// Add other allowed column names as needed
	}

	if request.Params.OrderByField != nil {
		orderByField := string(*request.Params.OrderByField)
		// Only set the order by field if it's in the whitelist
		if allowedOrderByFields[orderByField] {
			listOpts.OrderBy = repository.StringPtr(orderByField)
		}
	}

	if request.Params.OrderByDirection != nil {
		orderDirection := strings.ToUpper(string(*request.Params.OrderByDirection))
		// Only set the order direction if it's ASC or DESC
		if orderDirection == "ASC" || orderDirection == "DESC" {
			listOpts.OrderDirection = repository.StringPtr(orderDirection)
		}
	}

	if request.Params.Limit != nil {
		limit = int(*request.Params.Limit)
		listOpts.Limit = &limit
	}

	if request.Params.Offset != nil {
		offset = int(*request.Params.Offset)
		listOpts.Offset = &offset
	}

	dbCtx, cancel := context.WithTimeout(ctx.Request().Context(), 30*time.Second)
	defer cancel()

	listRes, err := t.config.EngineRepository.RateLimit().ListRateLimits(dbCtx, tenantId, listOpts)

	if err != nil {
		return nil, err
	}

	rows := make([]gen.RateLimit, len(listRes.Rows))

	for i, RateLimit := range listRes.Rows {
		RateLimitData, err := transformers.ToRateLimitFromSQLC(RateLimit)
		if err != nil {
			return nil, err
		}
		rows[i] = *RateLimitData
	}

	// use the total rows and limit to calculate the total pages
	totalPages := int64(math.Ceil(float64(listRes.Count) / float64(limit)))
	currPage := 1 + int64(math.Ceil(float64(offset)/float64(limit)))
	nextPage := currPage + 1

	if currPage == totalPages {
		nextPage = currPage
	}

	return gen.RateLimitList200JSONResponse(
		gen.RateLimitList{
			Rows: &rows,
			Pagination: &gen.PaginationResponse{
				NumPages:    &totalPages,
				NextPage:    &nextPage,
				CurrentPage: &currPage,
			},
		},
	), nil
}