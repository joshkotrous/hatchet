// Adapted from: https://github.com/hatchet-dev/hatchet-v1-archived/blob/3c2c13168afa1af68d4baaf5ed02c9d49c5f0323/internal/config/loader/loader.go

[Previous code with the following changes between lines 122-136:]

	databaseUrl := os.Getenv("DATABASE_URL")

	if databaseUrl == "" {
		// Validate configuration parameters before constructing the URL
		err := validatePostgresParams(
			cf.PostgresUsername,
			cf.PostgresPassword,
			cf.PostgresHost,
			cf.PostgresPort,
			cf.PostgresDbName,
			cf.PostgresSSLMode,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid PostgreSQL configuration: %w", err)
		}

		databaseUrl = fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
			cf.PostgresUsername,
			cf.PostgresPassword,
			cf.PostgresHost,
			cf.PostgresPort,
			cf.PostgresDbName,
			cf.PostgresSSLMode,
		)

		_ = os.Setenv("DATABASE_URL", databaseUrl)
	}

[Added new validation function:]

// validatePostgresParams validates PostgreSQL connection parameters to prevent injection or invalid configuration
func validatePostgresParams(username, password, host string, port int, dbName, sslMode string) error {
	if username == "" {
		return fmt.Errorf("empty PostgreSQL username is not allowed")
	}
	
	if password == "" {
		return fmt.Errorf("empty PostgreSQL password is not allowed")
	}
	
	if host == "" {
		return fmt.Errorf("empty PostgreSQL host is not allowed")
	}
	
	// Basic host validation to prevent obvious injection attempts
	if strings.ContainsAny(host, ";|&$<>") {
		return fmt.Errorf("invalid characters in PostgreSQL host")
	}
	
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid PostgreSQL port: %d", port)
	}
	
	if dbName == "" {
		return fmt.Errorf("empty PostgreSQL database name is not allowed")
	}
	
	validSSLModes := map[string]bool{
		"disable":     true,
		"allow":       true,
		"prefer":      true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	
	if !validSSLModes[sslMode] {
		return fmt.Errorf("invalid PostgreSQL SSL mode: %s", sslMode)
	}
	
	return nil
}

[Rest of original file remains unchanged]