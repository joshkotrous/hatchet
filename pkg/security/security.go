package security

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hatchet-dev/hatchet/pkg/repository"

	"github.com/rs/zerolog"
)

type SecurityCheck interface {
	Check()
}

type DefaultSecurityCheck struct {
	Enabled  bool
	Endpoint string
	Logger   *zerolog.Logger
	Version  string
	Repo     repository.SecurityCheckRepository
}

func NewSecurityCheck(opts *DefaultSecurityCheck, repo repository.SecurityCheckRepository) SecurityCheck {
	return DefaultSecurityCheck{
		Enabled:  opts.Enabled,
		Endpoint: opts.Endpoint,
		Logger:   opts.Logger,
		Version:  opts.Version,
		Repo:     repo,
	}
}

func (a DefaultSecurityCheck) Check() {
	if !a.Enabled {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic in check: %v", r)
		}
	}()

	a.Logger.Debug().Msgf("Fetching security alerts for version %s", a.Version)

	ident, err := a.Repo.GetIdent()
	if err != nil {
		a.Logger.Debug().Msgf("Error fetching security alerts: %s", err)
		return
	}

	// Validate endpoint is a proper URL
	endpointURL, err := url.Parse(a.Endpoint)
	if err != nil {
		a.Logger.Debug().Msgf("Invalid endpoint URL: %s", err)
		return
	}
	
	// Ensure endpoint is HTTP or HTTPS
	if endpointURL.Scheme != "http" && endpointURL.Scheme != "https" {
		a.Logger.Debug().Msgf("Endpoint URL must use HTTP or HTTPS scheme")
		return
	}
	
	// Handle path joining correctly
	endpointPath := strings.TrimSuffix(endpointURL.Path, "/")
	endpointURL.Path = endpointPath + "/check"
	
	// Construct the query parameters safely
	query := url.Values{}
	query.Add("version", a.Version)
	query.Add("tag", ident)
	endpointURL.RawQuery = query.Encode()
	
	// Make the request with the safely constructed URL
	resp, err := http.Get(endpointURL.String())
	if err != nil {
		a.Logger.Debug().Msgf("Error making request to security endpoint: %s", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.Logger.Debug().Msgf("Unexpected status code from security endpoint: %d", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.Logger.Debug().Msgf("Error reading response body: %s", err)
		return
	}

	if len(body) == 0 {
		a.Logger.Debug().Msg("No security alerts found")
		return
	}

	a.Logger.Error().Msgf("Security Alert:\n\n%s\n******************\n", body)
}