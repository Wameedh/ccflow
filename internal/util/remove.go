package util

import "os"

// RemoveSymlink removes a symbolic link
func RemoveSymlink(path string) error {
	return os.Remove(path)
}

// RemoveAll removes a path and all its contents
func RemoveAll(path string) error {
	return os.RemoveAll(path)
}
