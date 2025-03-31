package postgres

import (
	"context"
	"errors"
	"fmt"
)

// validateMessageBytes validates the message bytes to prevent adversarial inputs
// that could manipulate ML model behavior or cause unexpected database operations.
func validateMessageBytes(msgBytes []byte) error {
	// Basic validation - empty check
	if len(msgBytes) == 0 {
		return errors.New("empty message content")
	}
	
	// TODO: Add specific validation logic based on application requirements
	// Examples:
	// - Validate message structure/format
	// - Check for known malicious patterns
	// - Apply content restrictions based on business rules
	// - Implement ML-specific input validation
	
	return nil
}

// validateMLOutput performs basic validation on the message bytes
// to ensure they are safe before processing
func validateMLOutput(msgBytes []byte) ([]byte, error) {
	// Check if data is empty
	if len(msgBytes) == 0 {
		return nil, errors.New("empty message data")
	}
	
	// TODO: Add specific validation for ML output format based on application requirements
	
	return msgBytes, nil
}

func (p *PostgresMessageQueue) addTenantExchangeMessage(ctx context.Context, tenantId string, msgBytes []byte) error {
	// Validate message bytes to prevent adversarial inputs
	if err := validateMessageBytes(msgBytes); err != nil {
		return errors.New("invalid message content: " + err.Error())
	}

	// Validate the message data before processing
	validatedData, err := validateMLOutput(msgBytes)
	if err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}
	
	// determine if the exchange message is greater than 8kb
	if len(validatedData) > 8000 {
		// if the message is greater than 8kb, store the message in the database
		return p.repo.AddMessage(ctx, tenantId, validatedData)
	}

	// if the message is less than 8kb, publish the message to the channel
	return p.repo.Notify(ctx, tenantId, string(validatedData))
}