package util

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsSymlink checks if a path is a symbolic link
func IsSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// ReadSymlinkTarget returns the target of a symbolic link
func ReadSymlinkTarget(path string) (string, error) {
	return os.Readlink(path)
}

// EnsureDir creates a directory and all parent directories if they don't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// SafeWriteFile writes content to a file, refusing to overwrite unless force is true
func SafeWriteFile(path string, content []byte, force bool) error {
	if FileExists(path) && !force {
		return fmt.Errorf("file already exists: %s (use --force to overwrite)", path)
	}

	// Ensure parent directory exists
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, content, 0644)
}

// SafeWriteExecutable writes content to a file with executable permissions
func SafeWriteExecutable(path string, content []byte, force bool) error {
	if FileExists(path) && !force {
		return fmt.Errorf("file already exists: %s (use --force to overwrite)", path)
	}

	// Ensure parent directory exists
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, content, 0755)
}

// MakeExecutable sets the executable bit on a file
func MakeExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.Chmod(path, info.Mode()|0111)
}

// IsExecutable checks if a file has executable permissions
func IsExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}

// ReadStdin reads all content from stdin
func ReadStdin() ([]byte, error) {
	reader := bufio.NewReader(os.Stdin)
	var content []byte
	for {
		b, err := reader.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		content = append(content, b)
	}
	return content, nil
}

// CreateRelativeSymlink creates a relative symbolic link
func CreateRelativeSymlink(target, linkPath string) error {
	// Calculate relative path from link location to target
	linkDir := filepath.Dir(linkPath)
	relTarget, err := filepath.Rel(linkDir, target)
	if err != nil {
		// Fall back to absolute path if relative fails
		relTarget = target
	}

	// Ensure parent directory exists
	if err := EnsureDir(linkDir); err != nil {
		return fmt.Errorf("failed to create directory for symlink: %w", err)
	}

	return os.Symlink(relTarget, linkPath)
}

// HashFileContent returns SHA256 hash of file content
func HashFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return HashBytes(content), nil
}

// HashBytes returns SHA256 hash of bytes
func HashBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Get source file info for permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	if dirErr := EnsureDir(filepath.Dir(dst)); dirErr != nil {
		return dirErr
	}

	destFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
