package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hatchet-dev/hatchet/pkg/encryption"
)

var (
	encryptionKeyDir        string
	cloudKMSCredentialsPath string
	cloudKMSKeyURI          string
)

var keysetCmd = &cobra.Command{
	Use:   "keyset",
	Short: "command for managing keysets.",
}

var keysetCreateLocalKeysetsCmd = &cobra.Command{
	Use:   "create-local-keys",
	Short: "create a new local master keyset and JWT public/private keyset.",
	Run: func(cmd *cobra.Command, args []string) {
		err := runCreateLocalKeysets()

		if err != nil {
			log.Printf("Fatal: could not run [keyset create-local-keys] command: %v", err)
			os.Exit(1)
		}
	},
}

var keysetCreateCloudKMSJWTCmd = &cobra.Command{
	Use:   "create-cloudkms-jwt",
	Short: "create a new JWT keyset encrypted by a remote CloudKMS repository.",
	Run: func(cmd *cobra.Command, args []string) {
		err := runCreateCloudKMSJWTKeyset()

		if err != nil {
			log.Printf("Fatal: could not run [keyset create-cloudkms-jwt] command: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(keysetCmd)
	keysetCmd.AddCommand(keysetCreateLocalKeysetsCmd)
	keysetCmd.AddCommand(keysetCreateCloudKMSJWTCmd)

	keysetCmd.PersistentFlags().StringVar(
		&encryptionKeyDir,
		"key-dir",
		"",
		"path to the directory where encryption keys should be stored (required for security reasons)",
	)

	keysetCreateCloudKMSJWTCmd.PersistentFlags().StringVar(
		&cloudKMSCredentialsPath,
		"credentials",
		"",
		"path to the JSON credentials file for the CloudKMS repository",
	)

	keysetCreateCloudKMSJWTCmd.PersistentFlags().StringVar(
		&cloudKMSKeyURI,
		"key-uri",
		"",
		"URI of the key in the CloudKMS repository",
	)
}

// safePath constructs a safe file path by joining the directory and filename,
// preventing path traversal attacks.
func safePath(dir, filename string) (string, error) {
	// Get the absolute path to eliminate any relative path components
	absDir, err := filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	// Join the directory and filename using the OS-specific path separator
	filePath := filepath.Join(absDir, filename)
	
	// Verify the resulting path is still under the intended directory
	// This defends against filenames containing path traversal elements
	if !strings.HasPrefix(filepath.Clean(filePath), absDir) {
		return "", fmt.Errorf("invalid filename path traversal detected")
	}
	
	return filePath, nil
}

func runCreateLocalKeysets() error {
	if encryptionKeyDir == "" {
		return fmt.Errorf("for security reasons, keys cannot be printed to standard output. Please provide a directory path using the --key-dir flag")
	}

	masterKeyBytes, privateEc256, publicEc256, err := encryption.GenerateLocalKeys()

	if err != nil {
		return err
	}

	// Ensure the directory exists
	if err := os.MkdirAll(encryptionKeyDir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// we write these as .key files so that they're gitignored by default
	masterKeyPath, err := safePath(encryptionKeyDir, "master.key")
	if err != nil {
		return err
	}
	
	err = os.WriteFile(masterKeyPath, masterKeyBytes, 0600)
	if err != nil {
		return err
	}

	privateKeyPath, err := safePath(encryptionKeyDir, "private_ec256.key")
	if err != nil {
		return err
	}
	
	err = os.WriteFile(privateKeyPath, privateEc256, 0600)
	if err != nil {
		return err
	}

	publicKeyPath, err := safePath(encryptionKeyDir, "public_ec256.key")
	if err != nil {
		return err
	}
	
	err = os.WriteFile(publicKeyPath, publicEc256, 0600)
	if err != nil {
		return err
	}

	fmt.Printf("Keys successfully written to %s\n", encryptionKeyDir)
	return nil
}

func runCreateCloudKMSJWTKeyset() error {
	if encryptionKeyDir == "" {
		return fmt.Errorf("for security reasons, keys cannot be printed to standard output. Please provide a directory path using the --key-dir flag")
	}

	if cloudKMSCredentialsPath == "" {
		return fmt.Errorf("missing required flag --credentials")
	}

	if cloudKMSKeyURI == "" {
		return fmt.Errorf("missing required flag --key-uri")
	}

	credentials, err := os.ReadFile(cloudKMSCredentialsPath)

	if err != nil {
		return err
	}

	privateEc256, publicEc256, err := encryption.GenerateJWTKeysetsFromCloudKMS(cloudKMSKeyURI, credentials)

	if err != nil {
		return err
	}

	// Ensure the directory exists
	if err := os.MkdirAll(encryptionKeyDir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// we write these as .key files so that they're gitignored by default
	privateKeyPath, err := safePath(encryptionKeyDir, "private_ec256.key")
	if err != nil {
		return err
	}
	
	err = os.WriteFile(privateKeyPath, privateEc256, 0600)
	if err != nil {
		return err
	}

	publicKeyPath, err := safePath(encryptionKeyDir, "public_ec256.key")
	if err != nil {
		return err
	}
	
	err = os.WriteFile(publicKeyPath, publicEc256, 0600)
	if err != nil {
		return err
	}

	fmt.Printf("Keys successfully written to %s\n", encryptionKeyDir)
	return nil
}