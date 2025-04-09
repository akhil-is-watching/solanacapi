package helpers

import (
	"fmt"
	"os"
)

func CreateFile(content string, path string) error {
	// Create directory structure if it doesn't exist
	dir := getDirectory(path)
	if dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// getDirectory extracts the directory part from a file path
func getDirectory(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return ""
}
