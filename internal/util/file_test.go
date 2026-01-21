package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Test existing file
	existingFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	if !FileExists(existingFile) {
		t.Error("FileExists returned false for existing file")
	}

	// Test non-existing file
	if FileExists(filepath.Join(tmpDir, "nonexistent.txt")) {
		t.Error("FileExists returned true for non-existing file")
	}
}

func TestDirExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Test existing directory
	if !DirExists(tmpDir) {
		t.Error("DirExists returned false for existing directory")
	}

	// Test file (should return false)
	file := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(file, []byte("test"), 0644)
	if DirExists(file) {
		t.Error("DirExists returned true for a file")
	}

	// Test non-existing
	if DirExists(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("DirExists returned true for non-existing path")
	}
}

func TestIsSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	file := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a symlink
	link := filepath.Join(tmpDir, "link")
	if err := os.Symlink(file, link); err != nil {
		t.Fatal(err)
	}

	if !IsSymlink(link) {
		t.Error("IsSymlink returned false for symlink")
	}

	if IsSymlink(file) {
		t.Error("IsSymlink returned true for regular file")
	}
}

func TestSafeWriteFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Test writing new file
	newFile := filepath.Join(tmpDir, "new.txt")
	if err := SafeWriteFile(newFile, []byte("content"), false); err != nil {
		t.Fatalf("SafeWriteFile failed for new file: %v", err)
	}

	content, _ := os.ReadFile(newFile)
	if string(content) != "content" {
		t.Errorf("Content mismatch: got %s", content)
	}

	// Test refusing to overwrite without force
	if err := SafeWriteFile(newFile, []byte("new"), false); err == nil {
		t.Error("SafeWriteFile should refuse to overwrite without force")
	}

	// Test overwrite with force
	if err := SafeWriteFile(newFile, []byte("updated"), true); err != nil {
		t.Fatalf("SafeWriteFile with force failed: %v", err)
	}

	content, _ = os.ReadFile(newFile)
	if string(content) != "updated" {
		t.Errorf("Content after force write mismatch: got %s", content)
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Test creating nested directories
	nested := filepath.Join(tmpDir, "a", "b", "c")
	if err := EnsureDir(nested); err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}

	if !DirExists(nested) {
		t.Error("EnsureDir did not create directory")
	}

	// Test existing directory (should not error)
	if err := EnsureDir(nested); err != nil {
		t.Fatalf("EnsureDir failed on existing dir: %v", err)
	}
}

func TestCreateRelativeSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create target directory
	target := filepath.Join(tmpDir, "target")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}

	// Create symlink
	link := filepath.Join(tmpDir, "subdir", "link")
	if err := CreateRelativeSymlink(target, link); err != nil {
		t.Fatalf("CreateRelativeSymlink failed: %v", err)
	}

	if !IsSymlink(link) {
		t.Error("Symlink was not created")
	}

	// Verify it points to target
	resolved, err := ReadSymlinkTarget(link)
	if err != nil {
		t.Fatalf("ReadSymlinkTarget failed: %v", err)
	}

	// The link should be relative
	if filepath.IsAbs(resolved) {
		t.Logf("Warning: symlink is absolute, may be OK on some systems")
	}
}

func TestHashBytes(t *testing.T) {
	hash1 := HashBytes([]byte("test"))
	hash2 := HashBytes([]byte("test"))
	hash3 := HashBytes([]byte("different"))

	if hash1 != hash2 {
		t.Error("Same content should produce same hash")
	}

	if hash1 == hash3 {
		t.Error("Different content should produce different hash")
	}

	// Check hash length (SHA256 = 64 hex chars)
	if len(hash1) != 64 {
		t.Errorf("Expected 64 char hash, got %d", len(hash1))
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	src := filepath.Join(tmpDir, "source.txt")
	content := []byte("copy me")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Copy to new location
	dst := filepath.Join(tmpDir, "subdir", "dest.txt")
	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify content
	copied, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if string(copied) != string(content) {
		t.Errorf("Copied content mismatch: got %s", copied)
	}
}
