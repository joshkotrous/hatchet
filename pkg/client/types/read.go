package types

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadHatchetYAMLFileBytes reads a given YAML file from a filepath and return the parsed workflow file
func ReadHatchetYAMLFileBytes(filepath string) (*Workflow, error) {
	yamlFileBytes, err := readHatchetYAMLFileBytes(filepath)

	if err != nil {
		return nil, err
	}

	// Basic validation before parsing
	if err := validateYAMLInput(yamlFileBytes); err != nil {
		return nil, fmt.Errorf("invalid YAML input: %w", err)
	}

	workflowFile, err := ParseYAML(context.Background(), yamlFileBytes)

	if err != nil {
		return nil, err
	}

	return &workflowFile, nil
}

func readHatchetYAMLFileBytes(filepath string) ([]byte, error) {
	// Clean and get the absolute path to protect against path traversal
	cleanPath := filepath.Clean(filepath)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("could not get absolute path: %w", err)
	}

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get current working directory: %w", err)
	}

	// Ensure the path is within the current working directory
	if !strings.HasPrefix(absPath, cwd) {
		return nil, fmt.Errorf("access denied: path '%s' is outside the allowed directory", filepath)
	}

	if !fileExists(absPath) {
		return nil, fmt.Errorf("file does not exist: %s", filepath)
	}

	yamlFileBytes, err := os.ReadFile(absPath) // Now using absPath which is sanitized

	if err != nil {
		return nil, err // Fixed: Return error instead of panic
	}

	return yamlFileBytes, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// validateYAMLInput performs basic validation on YAML content
// to prevent potential attacks and ensure input quality
func validateYAMLInput(content []byte) error {
	// Check for empty content
	if len(content) == 0 {
		return fmt.Errorf("empty YAML content")
	}
	
	// Size limit check to prevent DoS attacks
	const maxSize = 10 * 1024 * 1024 // 10MB example limit
	if len(content) > maxSize {
		return fmt.Errorf("YAML content exceeds maximum allowed size of %d bytes", maxSize)
	}
	
	// Additional validation could be added here as requirements become clearer
	
	return nil
}