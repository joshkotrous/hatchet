package types

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"bytes"
)

func DefaultLoader() []*Workflow {
	workflowFiles, err := ReadAllValidFilesInDir("./.hatchet")

	if err != nil {
		panic(err)
	}

	return workflowFiles
}

func ReadAllValidFilesInDir(filedir string) ([]*Workflow, error) {
	files, err := readYAMLFiles(filedir)

	if err != nil {
		return nil, err
	}

	var workflowFiles []*Workflow

	for _, file := range files {
		workflowFile, err := ParseYAML(context.Background(), file)

		if err != nil {
			continue
		}

		workflowFiles = append(workflowFiles, &workflowFile)
	}

	return workflowFiles, nil
}

// readYAMLFiles reads all .yaml files in a given directory, including subdirectories.
// It applies basic validation to protect against potentially malicious YAML content.
func readYAMLFiles(rootDir string) ([][]byte, error) {
	yamlFiles := make([][]byte, 0)
	const maxFileSize = 5 * 1024 * 1024 // 5MB limit for YAML files

	// Walk the directory tree
	err := filepath.WalkDir(rootDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if the file is a YAML file
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			// Check file size before reading
			fileInfo, err := info.Info()
			if err != nil {
				return fmt.Errorf("error getting file info for %s: %v", path, err)
			}
			
			if fileInfo.Size() > maxFileSize {
				return fmt.Errorf("file %s exceeds maximum allowed size of %d bytes", path, maxFileSize)
			}

			// Read the file
			data, err := os.ReadFile(path) // #nosec G304 -- files are meant to be read from user-supplied directory
			if err != nil {
				return fmt.Errorf("error reading file %s: %v", path, err)
			}

			// Perform basic validation of the YAML content
			if err := validateBasicYAMLContent(data, path); err != nil {
				return err
			}
			
			yamlFiles = append(yamlFiles, data)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error processing YAML files in %s: %v", rootDir, err)
	}

	return yamlFiles, nil
}

// validateBasicYAMLContent performs basic validation on YAML content to detect potential security issues
func validateBasicYAMLContent(content []byte, filePath string) error {
    // Check for known dangerous YAML tags
    dangerousTags := []string{
        "!!python", "!!ruby", "!!php", "!!java", "!!exec", 
        "!!binary", "!!perl", "!!js/function", "!!js/regexp",
    }
    
    for _, tag := range dangerousTags {
        if bytes.Contains(content, []byte(tag)) {
            return fmt.Errorf("potentially unsafe YAML tag '%s' detected in file %s", tag, filePath)
        }
    }
    
    // Check for excessive nesting or complexity which might cause parsing issues
    lines := bytes.Count(content, []byte("\n"))
    if lines > 10000 {
        return fmt.Errorf("YAML file %s has too many lines (%d), maximum allowed is 10000", filePath, lines)
    }
    
    // Check for balanced key-value structure (very basic check)
    colons := bytes.Count(content, []byte(":"))
    if colons > 10000 {
        return fmt.Errorf("YAML file %s has too many key-value pairs, maximum allowed is 10000", filePath)
    }
    
    return nil
}