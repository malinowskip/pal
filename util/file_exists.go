package util

import "os"

// Checks if the file at the provided path exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	return err == nil
}
