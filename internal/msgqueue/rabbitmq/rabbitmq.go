package rabbitmq

[Previous code content unchanged until line 311...]

// validateExchangeName ensures that the tenant ID is safe to use as an exchange name
// by checking for length limits and valid characters.
func validateExchangeName(name string) error {
	if name == "" {
		return fmt.Errorf("exchange name cannot be empty")
	}
	
	if len(name) > 255 {
		return fmt.Errorf("exchange name too long, maximum is 255 characters")
	}
	
	// Check for valid characters
	for i, r := range name {
		if !((r >= 'a' && r <= 'z') || 
			 (r >= 'A' && r <= 'Z') || 
			 (r >= '0' && r <= '9') || 
			 r == '-' || r == '_' || r == '.' || r == ':') {
			return fmt.Errorf("invalid character '%c' at position %d in exchange name", r, i)
		}
	}
	
	return nil
}

func (t *MessageQueueImpl) RegisterTenant(ctx context.Context, tenantId string) error {
	// Validate the tenantId before using it as an exchange name
	if err := validateExchangeName(tenantId); err != nil {
		t.l.Error().Err(err).Msgf("invalid tenant ID: %s", tenantId)
		return fmt.Errorf("invalid tenant ID for exchange name: %w", err)
	}
	
	// create a new fanout exchange for the tenant
	sub := <-<-t.sessions

	t.l.Debug().Msgf("registering tenant exchange: %s", tenantId)

	// create a fanout exchange for the tenant. each consumer of the fanout exchange will get notified
	// with the tenant events.
	err := sub.ExchangeDeclare(
		tenantId,
		"fanout",
		true,  // durable
		false, // auto-deleted
		false, // not internal, accepts publishings
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		t.l.Error().Msgf("cannot declare exchange: %q, %v", tenantId, err)
		return err
	}

	t.tenantIdCache.Add(tenantId, true)

	return nil
}

[Remaining code content unchanged...]