package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInteractiveCLI_Run_NoGroups tests running with no groups.
func TestInteractiveCLI_Run_NoGroups(t *testing.T) {
	diffExec := NewDiffExecutor("")
	cli := NewInteractiveCLI(nil, diffExec)

	var output bytes.Buffer
	cli.writer = &output

	err := cli.Run()
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if !strings.Contains(output.String(), "No groups") {
		t.Error("Run() should indicate no groups found")
	}
}

// TestInteractiveCLI_GeneratePairs tests pair generation.
func TestInteractiveCLI_GeneratePairs(t *testing.T) {
	cli := NewInteractiveCLI(nil, nil)

	tests := []struct {
		name     string
		group    []string
		expected int
	}{
		{"Two files", []string{"file1.txt", "file2.txt"}, 1},
		{"Three files", []string{"file1.txt", "file2.txt", "file3.txt"}, 3},
		{"Four files", []string{"a.txt", "b.txt", "c.txt", "d.txt"}, 6},
		{"Empty group", []string{}, 0},
		{"Single file", []string{"file1.txt"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pairs := cli.generatePairs(tt.group)
			if len(pairs) != tt.expected {
				t.Errorf("generatePairs() returned %d pairs, expected %d", len(pairs), tt.expected)
			}

			// Verify each pair has exactly 2 files
			for i, pair := range pairs {
				if len(pair) != 2 {
					t.Errorf("generatePairs()[%d] has %d files, expected 2", i, len(pair))
				}
			}
		})
	}
}

// TestInteractiveCLI_GeneratePairs_UniquePairs tests that pairs are unique.
func TestInteractiveCLI_GeneratePairs_UniquePairs(t *testing.T) {
	cli := NewInteractiveCLI(nil, nil)
	group := []string{"a.txt", "b.txt", "c.txt"}
	pairs := cli.generatePairs(group)

	// Verify no duplicate pairs
	pairSet := make(map[string]bool)
	for _, pair := range pairs {
		key := pair[0] + ":" + pair[1]
		if pairSet[key] {
			t.Errorf("generatePairs() returned duplicate pair: %v", pair)
		}
		pairSet[key] = true
	}
}

// TestInteractiveCLI_ShowDiff tests showing a diff between two files.
func TestInteractiveCLI_ShowDiff(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	file1 := createFileWithContent(t, tmpDir, "file1.txt", "line 1\nline 2\n")
	file2 := createFileWithContent(t, tmpDir, "file2.txt", "line 1\nline 3\n")

	diffExec := NewDiffExecutor("")
	cli := NewInteractiveCLI(nil, diffExec)

	var output bytes.Buffer
	cli.writer = &output

	err := cli.showDiff(file1, file2)
	if err != nil {
		t.Fatalf("showDiff() returned error: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, filepath.Base(file1)) {
		t.Error("showDiff() should display first filename")
	}
	if !strings.Contains(outputStr, filepath.Base(file2)) {
		t.Error("showDiff() should display second filename")
	}
}

// TestInteractiveCLI_ShowDiff_IdenticalFiles tests showing diff of identical files.
func TestInteractiveCLI_ShowDiff_IdenticalFiles(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	content := "same content\n"
	file1 := createFileWithContent(t, tmpDir, "file1.txt", content)
	file2 := createFileWithContent(t, tmpDir, "file2.txt", content)

	diffExec := NewDiffExecutor("")
	cli := NewInteractiveCLI(nil, diffExec)

	var output bytes.Buffer
	cli.writer = &output

	err := cli.showDiff(file1, file2)
	if err != nil {
		t.Fatalf("showDiff() returned error: %v", err)
	}

	// Should not error even if files are identical
	_ = output.String()
}

// TestInteractiveCLI_HandleGroup_EmptyGroup tests handling an empty group.
func TestInteractiveCLI_HandleGroup_EmptyGroup(t *testing.T) {
	diffExec := NewDiffExecutor("")
	cli := NewInteractiveCLI(nil, diffExec)

	var output bytes.Buffer
	cli.writer = &output
	input := strings.NewReader("")
	cli.scanner = newScannerFromReader(input)

	err := cli.handleGroup(1, []string{})
	if err != nil {
		t.Fatalf("handleGroup() returned error: %v", err)
	}

	if !strings.Contains(output.String(), "No pairs") {
		t.Error("handleGroup() should indicate no pairs for empty group")
	}
}

// TestInteractiveCLI_HandleGroup_SingleFileGroup tests handling a group with one file.
func TestInteractiveCLI_HandleGroup_SingleFileGroup(t *testing.T) {
	diffExec := NewDiffExecutor("")
	cli := NewInteractiveCLI(nil, diffExec)

	var output bytes.Buffer
	cli.writer = &output
	input := strings.NewReader("")
	cli.scanner = newScannerFromReader(input)

	err := cli.handleGroup(1, []string{"/path/to/file.txt"})
	if err != nil {
		t.Fatalf("handleGroup() returned error: %v", err)
	}

	if !strings.Contains(output.String(), "No pairs") {
		t.Error("handleGroup() should indicate no pairs for single file group")
	}
}

// Helper function to create a scanner from a reader for testing
func newScannerFromReader(r *strings.Reader) *bufio.Scanner {
	return bufio.NewScanner(r)
}
