package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDiffExecutor_DiffSideBySide_IdenticalFiles tests diffing two identical files.
func TestDiffExecutor_DiffSideBySide_IdenticalFiles(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	content := "line 1\nline 2\nline 3\n"
	file1 := createFileWithContent(t, tmpDir, "file1.txt", content)
	file2 := createFileWithContent(t, tmpDir, "file2.txt", content)

	executor := NewDiffExecutor("")
	output, err := executor.DiffSideBySide(file1, file2)

	if err != nil {
		t.Fatalf("DiffSideBySide() returned error: %v", err)
	}
	// For identical files, diff -y may return empty output or show content side by side
	// Both are acceptable - the important thing is no error occurred
	_ = output
}

// TestDiffExecutor_DiffSideBySide_DifferentFiles tests diffing two different files.
func TestDiffExecutor_DiffSideBySide_DifferentFiles(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	file1 := createFileWithContent(t, tmpDir, "file1.txt", "line 1\nline 2\n")
	file2 := createFileWithContent(t, tmpDir, "file2.txt", "line 1\nline 3\n")

	executor := NewDiffExecutor("")
	output, err := executor.DiffSideBySide(file1, file2)

	if err != nil {
		t.Fatalf("DiffSideBySide() returned error: %v", err)
	}
	if output == "" {
		t.Error("DiffSideBySide() returned empty output for different files")
	}
	if !strings.Contains(output, "line 2") || !strings.Contains(output, "line 3") {
		t.Error("DiffSideBySide() output should contain differences")
	}
}

// TestDiffExecutor_DiffUnified_IdenticalFiles tests unified diff of identical files.
func TestDiffExecutor_DiffUnified_IdenticalFiles(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	content := "line 1\nline 2\n"
	file1 := createFileWithContent(t, tmpDir, "file1.txt", content)
	file2 := createFileWithContent(t, tmpDir, "file2.txt", content)

	executor := NewDiffExecutor("")
	output, err := executor.DiffUnified(file1, file2)

	if err != nil {
		t.Fatalf("DiffUnified() returned error: %v", err)
	}
	// For identical files, unified diff should be empty or minimal
	_ = output // Output may be empty for identical files
}

// TestDiffExecutor_DiffUnified_DifferentFiles tests unified diff of different files.
func TestDiffExecutor_DiffUnified_DifferentFiles(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	file1 := createFileWithContent(t, tmpDir, "file1.txt", "line 1\nline 2\n")
	file2 := createFileWithContent(t, tmpDir, "file2.txt", "line 1\nline 3\n")

	executor := NewDiffExecutor("")
	output, err := executor.DiffUnified(file1, file2)

	if err != nil {
		t.Fatalf("DiffUnified() returned error: %v", err)
	}
	if !strings.Contains(output, "-") || !strings.Contains(output, "+") {
		t.Error("DiffUnified() output should contain diff markers")
	}
}

// TestDiffExecutor_FilesIdentical_IdenticalFiles tests checking if identical files are identical.
func TestDiffExecutor_FilesIdentical_IdenticalFiles(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	content := "test content\n"
	file1 := createFileWithContent(t, tmpDir, "file1.txt", content)
	file2 := createFileWithContent(t, tmpDir, "file2.txt", content)

	executor := NewDiffExecutor("")
	identical, err := executor.FilesIdentical(file1, file2)

	if err != nil {
		t.Fatalf("FilesIdentical() returned error: %v", err)
	}
	if !identical {
		t.Error("FilesIdentical() should return true for identical files")
	}
}

// TestDiffExecutor_FilesIdentical_DifferentFiles tests checking if different files are identical.
func TestDiffExecutor_FilesIdentical_DifferentFiles(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	file1 := createFileWithContent(t, tmpDir, "file1.txt", "content 1\n")
	file2 := createFileWithContent(t, tmpDir, "file2.txt", "content 2\n")

	executor := NewDiffExecutor("")
	identical, err := executor.FilesIdentical(file1, file2)

	if err != nil {
		t.Fatalf("FilesIdentical() returned error: %v", err)
	}
	if identical {
		t.Error("FilesIdentical() should return false for different files")
	}
}

// TestDiffExecutor_FilesIdentical_NonexistentFile tests checking with non-existent file.
func TestDiffExecutor_FilesIdentical_NonexistentFile(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	file1 := createFileWithContent(t, tmpDir, "file1.txt", "content\n")
	file2 := filepath.Join(tmpDir, "nonexistent.txt")

	executor := NewDiffExecutor("")
	_, err := executor.FilesIdentical(file1, file2)

	if err == nil {
		t.Error("FilesIdentical() should return error for non-existent file")
	}
}

// TestDiffExecutor_CustomDiffCommand tests using a custom diff command.
func TestDiffExecutor_CustomDiffCommand(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	file1 := createFileWithContent(t, tmpDir, "file1.txt", "content\n")
	file2 := createFileWithContent(t, tmpDir, "file2.txt", "content\n")

	executor := NewDiffExecutor("diff")
	identical, err := executor.FilesIdentical(file1, file2)

	if err != nil {
		t.Fatalf("FilesIdentical() with custom command returned error: %v", err)
	}
	if !identical {
		t.Error("FilesIdentical() should work with custom diff command")
	}
}

// TestDiffExecutor_DefaultDiffCommand tests that empty string defaults to "diff".
func TestDiffExecutor_DefaultDiffCommand(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	file1 := createFileWithContent(t, tmpDir, "file1.txt", "content\n")
	file2 := createFileWithContent(t, tmpDir, "file2.txt", "content\n")

	executor := NewDiffExecutor("")
	identical, err := executor.FilesIdentical(file1, file2)

	if err != nil {
		t.Fatalf("FilesIdentical() with default command returned error: %v", err)
	}
	if !identical {
		t.Error("FilesIdentical() should work with default diff command")
	}
}

// Helper functions

func createFileWithContent(t *testing.T, dir, fileName, content string) string {
	filePath := filepath.Join(dir, fileName)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file %q: %v", filePath, err)
	}
	return filePath
}
