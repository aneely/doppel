package main

import (
	"path/filepath"
)

// Matcher groups files by common prefix.
type Matcher struct {
	minPrefixLength int
}

// NewMatcher creates a new Matcher with the specified minimum prefix length.
func NewMatcher(minPrefixLength int) *Matcher {
	return &Matcher{minPrefixLength: minPrefixLength}
}

// Group groups files by their common prefix.
// Returns a slice of groups, where each group contains files that share a common prefix.
// Only groups with 2 or more files are returned.
func (m *Matcher) Group(files []string) [][]string {
	if len(files) < 2 {
		return nil
	}

	// Extract just the filenames (without directory path) for prefix matching
	type fileInfo struct {
		filename string
		fullPath string
	}
	var fileInfos []fileInfo
	for _, file := range files {
		filename := filepath.Base(file)
		fileInfos = append(fileInfos, fileInfo{filename: filename, fullPath: file})
	}

	// Build groups: files that share a prefix of sufficient length belong to the same group
	// Use a union-find approach: each file starts in its own group, then merge groups
	// that share a common prefix
	groupID := make([]int, len(fileInfos))
	for i := range groupID {
		groupID[i] = i
	}

	// Find all pairs that share a prefix and merge their groups
	for i := 0; i < len(fileInfos); i++ {
		for j := i + 1; j < len(fileInfos); j++ {
			prefix := commonPrefix(fileInfos[i].filename, fileInfos[j].filename)
			if len(prefix) >= m.minPrefixLength {
				// Merge groups: make j's group point to i's group
				rootI := findRoot(groupID, i)
				rootJ := findRoot(groupID, j)
				if rootI != rootJ {
					groupID[rootJ] = rootI
				}
			}
		}
	}

	// Collect files by their group
	groups := make(map[int][]string)
	for i, fileInfo := range fileInfos {
		root := findRoot(groupID, i)
		groups[root] = append(groups[root], fileInfo.fullPath)
	}

	// Filter to only groups with 2+ files and convert to slice
	var result [][]string
	for _, group := range groups {
		if len(group) >= 2 {
			result = append(result, group)
		}
	}

	return result
}

// findRoot finds the root of a group using path compression.
func findRoot(groupID []int, x int) int {
	if groupID[x] != x {
		groupID[x] = findRoot(groupID, groupID[x])
	}
	return groupID[x]
}

// commonPrefix returns the common prefix of two strings.
func commonPrefix(a, b string) string {
	var i int
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i < minLen && a[i] == b[i] {
		i++
	}
	return a[:i]
}
