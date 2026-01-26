package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIntegration_ScanMatchGroup tests the full workflow: scan -> match -> group.
func TestIntegration_ScanMatchGroup(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	// Create test files with similar names
	filesToCreate := []string{
		"document.txt",
		"document-1.txt",
		"document_copy.txt",
		"image.png",
		"image-1.png",
		"unrelated.txt",
	}

	for _, fileName := range filesToCreate {
		createFile(t, tmpDir, fileName)
	}

	// Step 1: Scan
	scanner := NewScanner(tmpDir)
	files, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}
	if len(files) != len(filesToCreate) {
		t.Fatalf("Scan() returned %d files, expected %d", len(files), len(filesToCreate))
	}

	// Step 2: Group
	matcher := NewMatcher(3)
	groups := matcher.Group(files)

	// Should find 2 groups: document* and image*
	if len(groups) != 2 {
		t.Fatalf("Group() returned %d groups, expected 2", len(groups))
	}

	// Verify document group
	documentGroupFound := false
	imageGroupFound := false
	for _, group := range groups {
		if len(group) == 3 {
			// Should be the document group
			documentGroupFound = true
			groupFilenames := make(map[string]bool)
			for _, file := range group {
				groupFilenames[filepath.Base(file)] = true
			}
			expectedDocs := []string{"document.txt", "document-1.txt", "document_copy.txt"}
			for _, expected := range expectedDocs {
				if !groupFilenames[expected] {
					t.Errorf("Document group missing file: %s", expected)
				}
			}
		} else if len(group) == 2 {
			// Should be the image group
			imageGroupFound = true
			groupFilenames := make(map[string]bool)
			for _, file := range group {
				groupFilenames[filepath.Base(file)] = true
			}
			expectedImages := []string{"image.png", "image-1.png"}
			for _, expected := range expectedImages {
				if !groupFilenames[expected] {
					t.Errorf("Image group missing file: %s", expected)
				}
			}
		}
	}

	if !documentGroupFound {
		t.Error("Document group not found")
	}
	if !imageGroupFound {
		t.Error("Image group not found")
	}
}

// TestIntegration_ScanMatchDiff tests the workflow: scan -> match -> diff.
func TestIntegration_ScanMatchDiff(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	// Create two similar files with different content
	file1 := createFileWithContent(t, tmpDir, "document.txt", "line 1\nline 2\n")
	file2 := createFileWithContent(t, tmpDir, "document-1.txt", "line 1\nline 3\n")

	// Step 1: Scan
	scanner := NewScanner(tmpDir)
	files, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}

	// Step 2: Group
	matcher := NewMatcher(3)
	groups := matcher.Group(files)

	if len(groups) != 1 {
		t.Fatalf("Group() returned %d groups, expected 1", len(groups))
	}
	if len(groups[0]) != 2 {
		t.Fatalf("Group()[0] has %d files, expected 2", len(groups[0]))
	}

	// Step 3: Diff
	diffExec := NewDiffExecutor("")
	diff, err := diffExec.DiffSideBySide(file1, file2)
	if err != nil {
		t.Fatalf("DiffSideBySide() failed: %v", err)
	}
	if diff == "" {
		t.Error("DiffSideBySide() returned empty output for different files")
	}

	// Verify files are different
	identical, err := diffExec.FilesIdentical(file1, file2)
	if err != nil {
		t.Fatalf("FilesIdentical() failed: %v", err)
	}
	if identical {
		t.Error("FilesIdentical() should return false for different files")
	}
}

// TestIntegration_IdenticalFiles tests that identical files are detected correctly.
func TestIntegration_IdenticalFiles(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	content := "same content\n"
	file1 := createFileWithContent(t, tmpDir, "document.txt", content)
	file2 := createFileWithContent(t, tmpDir, "document-1.txt", content)

	// Scan and group
	scanner := NewScanner(tmpDir)
	files, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}

	matcher := NewMatcher(3)
	groups := matcher.Group(files)

	if len(groups) != 1 {
		t.Fatalf("Group() returned %d groups, expected 1", len(groups))
	}

	// Check if files are identical
	diffExec := NewDiffExecutor("")
	identical, err := diffExec.FilesIdentical(file1, file2)
	if err != nil {
		t.Fatalf("FilesIdentical() failed: %v", err)
	}
	if !identical {
		t.Error("FilesIdentical() should return true for identical files")
	}
}

// TestIntegration_EmptyDirectory tests the workflow with an empty directory.
func TestIntegration_EmptyDirectory(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	scanner := NewScanner(tmpDir)
	files, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}

	matcher := NewMatcher(3)
	groups := matcher.Group(files)

	if groups != nil {
		t.Errorf("Group() returned groups for empty directory, expected nil")
	}
}

// TestIntegration_MinPrefixLength tests that min prefix length is respected in full workflow.
func TestIntegration_MinPrefixLength(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	// Create files with short common prefix
	createFile(t, tmpDir, "ab.txt")
	createFile(t, tmpDir, "ab-1.txt") // "ab" is only 2 characters

	scanner := NewScanner(tmpDir)
	files, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}

	// Use min prefix of 3 - should not group
	matcher := NewMatcher(3)
	groups := matcher.Group(files)

	if groups != nil {
		t.Errorf("Group() returned groups with min prefix 3, expected nil")
	}

	// Use min prefix of 2 - should group
	matcher2 := NewMatcher(2)
	groups2 := matcher2.Group(files)

	if len(groups2) != 1 {
		t.Errorf("Group() with min prefix 2 returned %d groups, expected 1", len(groups2))
	}
}
