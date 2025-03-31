package loader

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type tokenConf struct {
	serverURL            string
	grpcBroadcastAddress string
	tenantId             string
}

// JWTSecretKey is the key used to verify JWT signatures
// This must be set by the application before using this package for JWT verification
var JWTSecretKey []byte

// SkipJWTVerification allows bypassing JWT signature verification (HIGHLY INSECURE)
// This should ONLY be set to true in development/testing environments
var SkipJWTVerification = false

func getConfFromJWT(token string) (*tokenConf, error) {
	claims, err := extractClaimsFromJWT(token)
	if err != nil {
		return nil, err
	}

	serverURL, ok := claims["server_url"].(string)
	if !ok {
		return nil, fmt.Errorf("server_url claim not found")
	}

	grpcBroadcastAddress, ok := claims["grpc_broadcast_address"].(string)
	if !ok {
		return nil, fmt.Errorf("grpc_broadcast_address claim not found")
	}

	tenantId, ok := claims["sub"].(string)

	if !ok {
		return nil, fmt.Errorf("sub claim not found")
	}

	return &tokenConf{
		serverURL:            serverURL,
		grpcBroadcastAddress: grpcBroadcastAddress,
		tenantId:             tenantId,
	}, nil
}

func extractClaimsFromJWT(tokenString string) (map[string]interface{}, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// In non-skip mode, verify the token signature
	if !SkipJWTVerification {
		if len(JWTSecretKey) == 0 {
			return nil, fmt.Errorf("JWT secret key not set")
		}

		// Verify signature (assuming HMAC-SHA256)
		headerAndPayload := parts[0] + "." + parts[1]
		signature, err := base64.RawURLEncoding.DecodeString(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid signature encoding: %w", err)
		}

		mac := hmac.New(sha256.New, JWTSecretKey)
		mac.Write([]byte(headerAndPayload))
		expectedMAC := mac.Sum(nil)

		if !hmac.Equal(signature, expectedMAC) {
			return nil, fmt.Errorf("invalid token signature")
		}

		// Check token header for algorithm
		headerData, err := base64.RawURLEncoding.DecodeString(parts[0])
		if err != nil {
			return nil, err
		}

		var header map[string]interface{}
		err = json.Unmarshal(headerData, &header)
		if err != nil {
			return nil, err
		}

		// Verify it's using the expected algorithm
		alg, ok := header["alg"].(string)
		if !ok || alg != "HS256" {
			return nil, fmt.Errorf("unsupported signing method: %v", alg)
		}
	}

	// Extract claims
	claimsData, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims map[string]interface{}
	err = json.Unmarshal(claimsData, &claims)
	if err != nil {
		return nil, err
	}

	// If we're verifying, check for token expiration
	if !SkipJWTVerification {
		// Check expiration if present
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				return nil, fmt.Errorf("token has expired")
			}
		}
	}

	return claims, nil
}