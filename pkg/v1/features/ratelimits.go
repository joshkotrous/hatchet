// package features provides functionality for interacting with hatchet features.
package features

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	v0Client "github.com/hatchet-dev/hatchet/pkg/client"
	"github.com/hatchet-dev/hatchet/pkg/client/rest"
	"github.com/hatchet-dev/hatchet/pkg/client/types"
)

// createRatelimitOpts contains options for creating or updating a rate limit.
type CreateRatelimitOpts struct {
	// key is the unique identifier for the rate limit
	Key string
	// limit is the maximum number of requests allowed within the duration
	Limit int
	// duration specifies the time period for the rate limit
	Duration types.RateLimitDuration
}

// rateLimitsClient provides an interface for managing rate limits.
type RateLimitsClient interface {
	// upsert creates or updates a rate limit with the provided options.
	Upsert(opts CreateRatelimitOpts) error

	// list retrieves rate limits based on the provided parameters (optional).
	List(ctx context.Context, opts *rest.RateLimitListParams) (*rest.RateLimitListResponse, error)
}

// rlClientImpl implements the rateLimitsClient interface.
type rlClientImpl struct {
	api      *rest.ClientWithResponses
	admin    *v0Client.AdminClient
	tenantId uuid.UUID
}

// newRateLimitsClient creates a new rateLimitsClient with the provided api client, tenant id, and admin client.
func NewRateLimitsClient(
	api *rest.ClientWithResponses,
	tenantId *string,
	admin *v0Client.AdminClient,
) (RateLimitsClient, error) {
	if tenantId == nil {
		return nil, fmt.Errorf("tenantId cannot be nil")
	}
	
	tenantIdUUid, err := uuid.Parse(*tenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format for tenantId: %w", err)
	}

	return &rlClientImpl{
		api:      api,
		tenantId: tenantIdUUid,
		admin:    admin,
	}, nil
}

// upsert creates or updates a rate limit with the provided options.
func (c *rlClientImpl) Upsert(opts CreateRatelimitOpts) error {
	return (*c.admin).PutRateLimit(opts.Key, &types.RateLimitOpts{
		Max:      opts.Limit,
		Duration: opts.Duration,
	})
}

// list retrieves rate limits based on the provided parameters (optional).
func (c *rlClientImpl) List(ctx context.Context, opts *rest.RateLimitListParams) (*rest.RateLimitListResponse, error) {
	return c.api.RateLimitListWithResponse(
		ctx,
		c.tenantId,
		opts,
	)
}