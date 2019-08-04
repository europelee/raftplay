package utils

import "os"

// PathExists check if path exists
func PathExists(p string) bool {
	if _, err := os.Lstat(p); err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
