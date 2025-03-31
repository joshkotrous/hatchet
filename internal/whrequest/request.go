package whrequest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/hatchet-dev/hatchet/internal/signature"
)

// validateURL validates that the provided URL doesn't target internal or restricted networks
func validateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Validate scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	// Extract hostname
	hostname := parsedURL.Hostname()

	// Check for localhost
	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
		return fmt.Errorf("requests to localhost are not allowed")
	}

	// Check if hostname is an IP address
	if ip := net.ParseIP(hostname); ip != nil {
		// Check for loopback
		if ip.IsLoopback() {
			return fmt.Errorf("requests to loopback addresses are not allowed")
		}

		// Check for private/internal network ranges
		// IPv4 private networks
		if ip4 := ip.To4(); ip4 != nil {
			// 10.0.0.0/8
			if ip4[0] == 10 {
				return fmt.Errorf("requests to private networks (10.0.0.0/8) are not allowed")
			}
			// 172.16.0.0/12
			if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
				return fmt.Errorf("requests to private networks (172.16.0.0/12) are not allowed")
			}
			// 192.168.0.0/16
			if ip4[0] == 192 && ip4[1] == 168 {
				return fmt.Errorf("requests to private networks (192.168.0.0/16) are not allowed")
			}
			// 169.254.0.0/16 (link local)
			if ip4[0] == 169 && ip4[1] == 254 {
				return fmt.Errorf("requests to link local addresses (169.254.0.0/16) are not allowed")
			}
		}

		// IPv6 link-local addresses (fe80::/10)
		if len(ip) == 16 && ip[0] == 0xfe && (ip[1]&0xc0) == 0x80 {
			return fmt.Errorf("requests to IPv6 link-local addresses are not allowed")
		}

		// IPv6 unique local addresses (fc00::/7)
		if len(ip) == 16 && (ip[0]&0xfe) == 0xfc {
			return fmt.Errorf("requests to IPv6 unique local addresses are not allowed")
		}
	}

	return nil
}

func Send(ctx context.Context, url string, secret string, data any, headers ...func(req *http.Request)) ([]byte, *int, error) {
	// Validate URL before sending request
	if err := validateURL(url); err != nil {
		return nil, nil, fmt.Errorf("URL validation failed: %w", err)
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	sig, err := signature.Sign(string(body), secret)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("X-Hatchet-Signature", sig)
	req.Header.Set("Content-Type", "application/json")

	for _, h := range headers {
		h(req)
	}

	httpClient := &http.Client{
		// use 10 minutes timeout
		Timeout: time.Second * 600,
	}

	// TODO block-list

	// nolint:gosec
	resp, err := httpClient.Do(req)
	if err != nil {
		connRefused := 502
		return nil, &connRefused, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &resp.StatusCode, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &resp.StatusCode, fmt.Errorf("could not read response body: %w", err)
	}

	return res, &resp.StatusCode, nil
}