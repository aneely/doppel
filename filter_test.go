package main

import (
	"regexp"
	"testing"
)

// TestFilterFilesBySuffix_NilPattern tests that nil pattern returns all files.
func TestFilterFilesBySuffix_NilPattern(t *testing.T) {
	files := []string{
		"/path/to/document.txt",
		"/path/to/document-1.txt",
		"/path/to/report-2024.txt",
	}

	result := filterFilesBySuffix(files, nil)

	if len(result) != len(files) {
		t.Errorf("filterFilesBySuffix() with nil pattern returned %d files, expected %d", len(result), len(files))
	}

	// Verify all files are present
	resultMap := make(map[string]bool)
	for _, f := range result {
		resultMap[f] = true
	}
	for _, expected := range files {
		if !resultMap[expected] {
			t.Errorf("filterFilesBySuffix() missing file: %s", expected)
		}
	}
}

// TestFilterFilesBySuffix_HyphenDigits tests filtering with hyphen + digits pattern.
func TestFilterFilesBySuffix_HyphenDigits(t *testing.T) {
	files := []string{
		"/path/to/document.txt",
		"/path/to/document-1.txt",
		"/path/to/document-2.txt",
		"/path/to/document-2026-01-30.txt",
		"/path/to/report-2024.txt",
		"/path/to/unrelated.txt",
	}

	pattern := regexp.MustCompile("-\\d{1,2}$")
	result := filterFilesBySuffix(files, pattern)

	// Should include: document-1.txt, document-2.txt, document.txt (base file)
	expectedFiles := map[string]bool{
		"/path/to/document.txt":     true,
		"/path/to/document-1.txt": true,
		"/path/to/document-2.txt":  true,
	}

	if len(result) != len(expectedFiles) {
		t.Errorf("filterFilesBySuffix() returned %d files, expected %d", len(result), len(expectedFiles))
	}

	resultMap := make(map[string]bool)
	for _, f := range result {
		resultMap[f] = true
	}

	for expected := range expectedFiles {
		if !resultMap[expected] {
			t.Errorf("filterFilesBySuffix() missing expected file: %s", expected)
		}
	}

	// Verify excluded files
	excludedFiles := []string{
		"/path/to/document-2026-01-30.txt",
		"/path/to/report-2024.txt",
		"/path/to/unrelated.txt",
	}
	for _, excluded := range excludedFiles {
		if resultMap[excluded] {
			t.Errorf("filterFilesBySuffix() incorrectly included: %s", excluded)
		}
	}
}

// TestFilterFilesBySuffix_SpaceDigits tests filtering with space + digits pattern.
func TestFilterFilesBySuffix_SpaceDigits(t *testing.T) {
	files := []string{
		"/path/to/file.txt",
		"/path/to/file 1.txt",
		"/path/to/file 2.txt",
		"/path/to/file-backup.txt",
	}

	pattern := regexp.MustCompile(" \\d+$")
	result := filterFilesBySuffix(files, pattern)

	// Should include: file 1.txt, file 2.txt, file.txt (base file)
	expectedFiles := map[string]bool{
		"/path/to/file.txt":   true,
		"/path/to/file 1.txt": true,
		"/path/to/file 2.txt": true,
	}

	if len(result) != len(expectedFiles) {
		t.Errorf("filterFilesBySuffix() returned %d files, expected %d", len(result), len(expectedFiles))
	}

	resultMap := make(map[string]bool)
	for _, f := range result {
		resultMap[f] = true
	}

	for expected := range expectedFiles {
		if !resultMap[expected] {
			t.Errorf("filterFilesBySuffix() missing expected file: %s", expected)
		}
	}

	// Verify excluded files
	if resultMap["/path/to/file-backup.txt"] {
		t.Error("filterFilesBySuffix() incorrectly included file-backup.txt")
	}
}

// TestFilterFilesBySuffix_NoMatches tests filtering when no files match pattern.
func TestFilterFilesBySuffix_NoMatches(t *testing.T) {
	files := []string{
		"/path/to/document.txt",
		"/path/to/report.txt",
	}

	pattern := regexp.MustCompile("-\\d+$")
	result := filterFilesBySuffix(files, pattern)

	if len(result) != 0 {
		t.Errorf("filterFilesBySuffix() returned %d files, expected 0", len(result))
	}
}

// TestFilterFilesBySuffix_NoBaseFile tests filtering when matching files exist but no base file.
func TestFilterFilesBySuffix_NoBaseFile(t *testing.T) {
	files := []string{
		"/path/to/document-1.txt",
		"/path/to/document-2.txt",
		// No document.txt
	}

	pattern := regexp.MustCompile("-\\d{1,2}$")
	result := filterFilesBySuffix(files, pattern)

	// Should only include matching files, no base file
	expectedFiles := map[string]bool{
		"/path/to/document-1.txt": true,
		"/path/to/document-2.txt": true,
	}

	if len(result) != len(expectedFiles) {
		t.Errorf("filterFilesBySuffix() returned %d files, expected %d", len(result), len(expectedFiles))
	}

	resultMap := make(map[string]bool)
	for _, f := range result {
		resultMap[f] = true
	}

	for expected := range expectedFiles {
		if !resultMap[expected] {
			t.Errorf("filterFilesBySuffix() missing expected file: %s", expected)
		}
	}
}

// TestFilterFilesBySuffix_DifferentPrefixes tests that files with different prefixes are handled correctly.
func TestFilterFilesBySuffix_DifferentPrefixes(t *testing.T) {
	files := []string{
		"/path/to/document.txt",
		"/path/to/document-1.txt",
		"/path/to/report.txt",
		"/path/to/report-1.txt",
	}

	pattern := regexp.MustCompile("-\\d{1,2}$")
	result := filterFilesBySuffix(files, pattern)

	// Should include all files (both document and report groups)
	expectedFiles := map[string]bool{
		"/path/to/document.txt": true,
		"/path/to/document-1.txt": true,
		"/path/to/report.txt": true,
		"/path/to/report-1.txt": true,
	}

	if len(result) != len(expectedFiles) {
		t.Errorf("filterFilesBySuffix() returned %d files, expected %d", len(result), len(expectedFiles))
	}

	resultMap := make(map[string]bool)
	for _, f := range result {
		resultMap[f] = true
	}

	for expected := range expectedFiles {
		if !resultMap[expected] {
			t.Errorf("filterFilesBySuffix() missing expected file: %s", expected)
		}
	}
}

// TestFilterFilesBySuffix_DateExclusion tests that date patterns are excluded with restrictive patterns.
func TestFilterFilesBySuffix_DateExclusion(t *testing.T) {
	files := []string{
		"/path/to/report.txt",
		"/path/to/report-1.txt",
		"/path/to/report-2024.txt", // 4 digits - should be excluded
	}

	pattern := regexp.MustCompile("-\\d{1,2}$") // Only 1-2 digits
	result := filterFilesBySuffix(files, pattern)

	// Should include: report.txt, report-1.txt
	// Should exclude: report-2024.txt
	expectedFiles := map[string]bool{
		"/path/to/report.txt": true,
		"/path/to/report-1.txt": true,
	}

	if len(result) != len(expectedFiles) {
		t.Errorf("filterFilesBySuffix() returned %d files, expected %d", len(result), len(expectedFiles))
	}

	resultMap := make(map[string]bool)
	for _, f := range result {
		resultMap[f] = true
	}

	for expected := range expectedFiles {
		if !resultMap[expected] {
			t.Errorf("filterFilesBySuffix() missing expected file: %s", expected)
		}
	}

	if resultMap["/path/to/report-2024.txt"] {
		t.Error("filterFilesBySuffix() incorrectly included report-2024.txt")
	}
}

// TestFilterFilesBySuffix_NoExtension tests filtering files without extensions.
func TestFilterFilesBySuffix_NoExtension(t *testing.T) {
	files := []string{
		"/path/to/file",
		"/path/to/file-1",
		"/path/to/file-2",
	}

	pattern := regexp.MustCompile("-\\d{1,2}$")
	result := filterFilesBySuffix(files, pattern)

	// Should include: file, file-1, file-2
	expectedFiles := map[string]bool{
		"/path/to/file": true,
		"/path/to/file-1": true,
		"/path/to/file-2": true,
	}

	if len(result) != len(expectedFiles) {
		t.Errorf("filterFilesBySuffix() returned %d files, expected %d", len(result), len(expectedFiles))
	}

	resultMap := make(map[string]bool)
	for _, f := range result {
		resultMap[f] = true
	}

	for expected := range expectedFiles {
		if !resultMap[expected] {
			t.Errorf("filterFilesBySuffix() missing expected file: %s", expected)
		}
	}
}

// TestFilterFilesBySuffix_MultipleDots tests filtering files with multiple dots in name.
func TestFilterFilesBySuffix_MultipleDots(t *testing.T) {
	files := []string{
		"/path/to/file.backup.txt",
		"/path/to/file.backup-1.txt",
		"/path/to/file.backup-2.txt",
	}

	pattern := regexp.MustCompile("-\\d{1,2}$")
	result := filterFilesBySuffix(files, pattern)

	// Should include: file.backup.txt, file.backup-1.txt, file.backup-2.txt
	expectedFiles := map[string]bool{
		"/path/to/file.backup.txt": true,
		"/path/to/file.backup-1.txt": true,
		"/path/to/file.backup-2.txt": true,
	}

	if len(result) != len(expectedFiles) {
		t.Errorf("filterFilesBySuffix() returned %d files, expected %d", len(result), len(expectedFiles))
	}

	resultMap := make(map[string]bool)
	for _, f := range result {
		resultMap[f] = true
	}

	for expected := range expectedFiles {
		if !resultMap[expected] {
			t.Errorf("filterFilesBySuffix() missing expected file: %s", expected)
		}
	}
}

// TestFilterFilesBySuffix_DuplicatePrevention tests that duplicates are not included.
func TestFilterFilesBySuffix_DuplicatePrevention(t *testing.T) {
	files := []string{
		"/path/to/document.txt",
		"/path/to/document-1.txt",
		"/path/to/document.txt", // Duplicate
	}

	pattern := regexp.MustCompile("-\\d{1,2}$")
	result := filterFilesBySuffix(files, pattern)

	// Should include each file only once
	seen := make(map[string]bool)
	for _, f := range result {
		if seen[f] {
			t.Errorf("filterFilesBySuffix() included duplicate: %s", f)
		}
		seen[f] = true
	}

	// Should have document.txt and document-1.txt
	expectedCount := 2
	if len(result) != expectedCount {
		t.Errorf("filterFilesBySuffix() returned %d files, expected %d", len(result), expectedCount)
	}
}
