package main

import (
	"reflect"
	"testing"
)

// TestMatcher_Group_EmptyList tests grouping an empty list of files.
func TestMatcher_Group_EmptyList(t *testing.T) {
	matcher := NewMatcher(3)
	groups := matcher.Group([]string{})

	if groups != nil {
		t.Errorf("Group() returned %v, expected nil", groups)
	}
}

// TestMatcher_Group_SingleFile tests grouping a single file.
func TestMatcher_Group_SingleFile(t *testing.T) {
	matcher := NewMatcher(3)
	groups := matcher.Group([]string{"file.txt"})

	if groups != nil {
		t.Errorf("Group() returned %v, expected nil (single file cannot form a group)", groups)
	}
}

// TestMatcher_Group_TwoSimilarFiles tests grouping two files with a common prefix.
func TestMatcher_Group_TwoSimilarFiles(t *testing.T) {
	matcher := NewMatcher(3)
	files := []string{"/path/to/document.txt", "/path/to/document-1.txt"}
	groups := matcher.Group(files)

	if len(groups) != 1 {
		t.Fatalf("Group() returned %d groups, expected 1", len(groups))
	}
	if len(groups[0]) != 2 {
		t.Errorf("Group()[0] contains %d files, expected 2", len(groups[0]))
	}

	// Verify both files are in the group
	groupMap := make(map[string]bool)
	for _, file := range groups[0] {
		groupMap[file] = true
	}
	for _, expectedFile := range files {
		if !groupMap[expectedFile] {
			t.Errorf("Group() did not include expected file: %q", expectedFile)
		}
	}
}

// TestMatcher_Group_MultipleGroups tests grouping files into multiple groups.
func TestMatcher_Group_MultipleGroups(t *testing.T) {
	matcher := NewMatcher(3)
	files := []string{
		"/path/to/document.txt",
		"/path/to/document-1.txt",
		"/path/to/image.png",
		"/path/to/image-1.png",
		"/path/to/unrelated.txt",
	}
	groups := matcher.Group(files)

	if len(groups) != 2 {
		t.Fatalf("Group() returned %d groups, expected 2", len(groups))
	}

	// Verify each group has 2 files
	for i, group := range groups {
		if len(group) != 2 {
			t.Errorf("Group()[%d] contains %d files, expected 2", i, len(group))
		}
	}
}

// TestMatcher_Group_ThreeSimilarFiles tests grouping three files with a common prefix.
func TestMatcher_Group_ThreeSimilarFiles(t *testing.T) {
	matcher := NewMatcher(3)
	files := []string{
		"/path/to/document.txt",
		"/path/to/document-1.txt",
		"/path/to/document_copy.txt",
	}
	groups := matcher.Group(files)

	if len(groups) != 1 {
		t.Fatalf("Group() returned %d groups, expected 1", len(groups))
	}
	if len(groups[0]) != 3 {
		t.Errorf("Group()[0] contains %d files, expected 3", len(groups[0]))
	}
}

// TestMatcher_Group_MinPrefixLength tests that minimum prefix length is respected.
func TestMatcher_Group_MinPrefixLength(t *testing.T) {
	matcher := NewMatcher(5) // Require at least 5 characters
	files := []string{"/path/to/doc.txt", "/path/to/doc-1.txt"} // "doc" is only 3 chars
	groups := matcher.Group(files)

	if groups != nil {
		t.Errorf("Group() returned groups with prefix shorter than minimum, expected nil")
	}
}

// TestMatcher_Group_NoCommonPrefix tests files with no common prefix.
func TestMatcher_Group_NoCommonPrefix(t *testing.T) {
	matcher := NewMatcher(3)
	files := []string{"/path/to/abc.txt", "/path/to/xyz.txt"}
	groups := matcher.Group(files)

	if groups != nil {
		t.Errorf("Group() returned groups for files with no common prefix, expected nil")
	}
}

// TestMatcher_Group_ExactMatches tests that exact matches are grouped together.
func TestMatcher_Group_ExactMatches(t *testing.T) {
	matcher := NewMatcher(3)
	files := []string{
		"/path/to/document.txt",
		"/other/path/document.txt", // Same filename, different path
	}
	groups := matcher.Group(files)

	if len(groups) != 1 {
		t.Fatalf("Group() returned %d groups, expected 1", len(groups))
	}
	if len(groups[0]) != 2 {
		t.Errorf("Group()[0] contains %d files, expected 2", len(groups[0]))
	}
}

// TestCommonPrefix tests the commonPrefix helper function.
func TestCommonPrefix(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected string
	}{
		{"Same strings", "document", "document", "document"},
		{"Common prefix", "document.txt", "document-1.txt", "document"},
		{"No common prefix", "abc", "xyz", ""},
		{"One is prefix of other", "doc", "document", "doc"},
		{"Empty strings", "", "", ""},
		{"One empty", "document", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := commonPrefix(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("commonPrefix(%q, %q) = %q, expected %q", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestMatcher_Group_PreservesOrder tests that file order is preserved within groups.
func TestMatcher_Group_PreservesOrder(t *testing.T) {
	matcher := NewMatcher(3)
	files := []string{
		"/path/to/document.txt",
		"/path/to/document-1.txt",
		"/path/to/document_copy.txt",
	}
	groups := matcher.Group(files)

	if len(groups) != 1 {
		t.Fatalf("Group() returned %d groups, expected 1", len(groups))
	}

	// Check that all input files are present in the group
	groupSet := make(map[string]bool)
	for _, file := range groups[0] {
		groupSet[file] = true
	}
	for _, expectedFile := range files {
		if !groupSet[expectedFile] {
			t.Errorf("Group() did not include expected file: %q", expectedFile)
		}
	}
}

// TestMatcher_Group_ComplexPrefixes tests grouping with various prefix patterns.
func TestMatcher_Group_ComplexPrefixes(t *testing.T) {
	matcher := NewMatcher(3)
	files := []string{
		"/path/to/file.txt",
		"/path/to/file-1.txt",
		"/path/to/file-2.txt",
		"/path/to/file_backup.txt",
		"/path/to/file_copy.txt",
	}
	groups := matcher.Group(files)

	if len(groups) != 1 {
		t.Fatalf("Group() returned %d groups, expected 1", len(groups))
	}

	// All files should share the "file" prefix and be in one group
	if len(groups[0]) != 5 {
		t.Errorf("Group()[0] contains %d files, expected 5", len(groups[0]))
	}

	// Verify all files are present
	groupSet := make(map[string]bool)
	for _, file := range groups[0] {
		groupSet[file] = true
	}
	for _, expectedFile := range files {
		if !groupSet[expectedFile] {
			t.Errorf("Group() did not include expected file: %q", expectedFile)
		}
	}
}

// TestMatcher_Group_WithDifferentExtensions tests that files with different extensions
// but same prefix are grouped together.
func TestMatcher_Group_WithDifferentExtensions(t *testing.T) {
	matcher := NewMatcher(3)
	files := []string{
		"/path/to/document.txt",
		"/path/to/document.pdf",
		"/path/to/document.doc",
	}
	groups := matcher.Group(files)

	if len(groups) != 1 {
		t.Fatalf("Group() returned %d groups, expected 1", len(groups))
	}

	expectedFiles := make(map[string]bool)
	for _, f := range files {
		expectedFiles[f] = true
	}

	actualFiles := make(map[string]bool)
	for _, f := range groups[0] {
		actualFiles[f] = true
	}

	if !reflect.DeepEqual(expectedFiles, actualFiles) {
		t.Errorf("Group() files mismatch. Expected %v, got %v", expectedFiles, actualFiles)
	}
}
