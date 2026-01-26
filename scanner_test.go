package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestScanner_Scan_EmptyDirectory tests scanning an empty directory.
func TestScanner_Scan_EmptyDirectory(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	scanner := NewScanner(tmpDir)
	files, err := scanner.Scan()

	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Scan() returned %d files, expected 0", len(files))
	}
}

// TestScanner_Scan_SingleFile tests scanning a directory with a single file.
func TestScanner_Scan_SingleFile(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	fileName := "test.txt"
	createFile(t, tmpDir, fileName)

	scanner := NewScanner(tmpDir)
	files, err := scanner.Scan()

	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("Scan() returned %d files, expected 1", len(files))
	}
	expectedPath := filepath.Join(tmpDir, fileName)
	if files[0] != expectedPath {
		t.Errorf("Scan() returned %q, expected %q", files[0], expectedPath)
	}
}

// TestScanner_Scan_MultipleFiles tests scanning a directory with multiple files.
func TestScanner_Scan_MultipleFiles(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	filesToCreate := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, fileName := range filesToCreate {
		createFile(t, tmpDir, fileName)
	}

	scanner := NewScanner(tmpDir)
	files, err := scanner.Scan()

	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}
	if len(files) != len(filesToCreate) {
		t.Fatalf("Scan() returned %d files, expected %d", len(files), len(filesToCreate))
	}

	// Verify all expected files are present
	fileMap := make(map[string]bool)
	for _, file := range files {
		fileMap[file] = true
	}

	for _, expectedFile := range filesToCreate {
		expectedPath := filepath.Join(tmpDir, expectedFile)
		if !fileMap[expectedPath] {
			t.Errorf("Scan() did not return expected file: %q", expectedPath)
		}
	}
}

// TestScanner_Scan_IgnoresDirectories tests that directories are not included in results.
func TestScanner_Scan_IgnoresDirectories(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	// Create a file and a directory
	createFile(t, tmpDir, "file.txt")
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	scanner := NewScanner(tmpDir)
	files, err := scanner.Scan()

	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Scan() returned %d files, expected 1 (directory should be ignored)", len(files))
	}
}

// TestScanner_Scan_NonexistentDirectory tests scanning a non-existent directory.
func TestScanner_Scan_NonexistentDirectory(t *testing.T) {
	scanner := NewScanner("/nonexistent/directory/path")
	_, err := scanner.Scan()

	if err == nil {
		t.Error("Scan() should return error for non-existent directory")
	}
}

// Helper functions

func createTempDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "scanner_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	return tmpDir
}

func createFile(t *testing.T, dir, fileName string) {
	filePath := filepath.Join(dir, fileName)
	if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create file %q: %v", filePath, err)
	}
}
