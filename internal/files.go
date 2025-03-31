package internal

import "os"

// copyFile copies a file from src to dst and retains the file permissions.
func copyFile(src, dst string) error {
	// Read the source file
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Get the file information to retrieve permissions
	fileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Extract the permissions from the file information
	sourcePermissions := fileInfo.Mode().Perm()

	// Write the data to the destination with the same permissions
	return os.WriteFile(dst, sourceData, sourcePermissions)
}
