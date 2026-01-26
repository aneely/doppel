package main

import (
	"os"
	"path/filepath"
)

// Scanner scans a directory and collects all files.
type Scanner struct {
	dir string
}

// NewScanner creates a new Scanner for the given directory.
func NewScanner(dir string) *Scanner {
	return &Scanner{dir: dir}
}

// Scan collects all files in the directory (non-recursive).
// Returns a slice of file paths relative to the scanned directory.
func (s *Scanner) Scan() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(s.dir, entry.Name()))
		}
	}

	return files, nil
}
