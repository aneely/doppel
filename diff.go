package main

import (
	"fmt"
	"os/exec"
)

// DiffExecutor executes system diff commands to compare files.
type DiffExecutor struct {
	diffCmd string
}

// NewDiffExecutor creates a new DiffExecutor with the specified diff command.
// If diffCmd is empty, defaults to "diff".
func NewDiffExecutor(diffCmd string) *DiffExecutor {
	if diffCmd == "" {
		diffCmd = "diff"
	}
	return &DiffExecutor{diffCmd: diffCmd}
}

// DiffSideBySide executes a side-by-side diff between two files.
// Returns the diff output as a string, or an error if the diff command fails.
func (d *DiffExecutor) DiffSideBySide(file1, file2 string) (string, error) {
	// Use diff -y for side-by-side output
	cmd := exec.Command(d.diffCmd, "-y", "--width=120", file1, file2)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// diff returns non-zero exit code when files differ, which is expected
		// Only return error if command execution itself failed
		if _, ok := err.(*exec.ExitError); !ok {
			return "", fmt.Errorf("failed to execute diff command: %w", err)
		}
	}
	return string(output), nil
}

// DiffUnified executes a unified diff between two files.
// Returns the diff output as a string, or an error if the diff command fails.
func (d *DiffExecutor) DiffUnified(file1, file2 string) (string, error) {
	cmd := exec.Command(d.diffCmd, "-u", file1, file2)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// diff returns non-zero exit code when files differ, which is expected
		// Only return error if command execution itself failed
		if _, ok := err.(*exec.ExitError); !ok {
			return "", fmt.Errorf("failed to execute diff command: %w", err)
		}
	}
	return string(output), nil
}

// FilesIdentical checks if two files are identical by comparing their content.
// Returns true if files are identical, false if they differ, and an error if comparison fails.
func (d *DiffExecutor) FilesIdentical(file1, file2 string) (bool, error) {
	cmd := exec.Command(d.diffCmd, "-q", file1, file2)
	err := cmd.Run()
	if err == nil {
		// Exit code 0 means files are identical
		return true, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		// Exit code 1 means files differ (this is expected)
		if exitErr.ExitCode() == 1 {
			return false, nil
		}
		// Other exit codes indicate an error
		return false, fmt.Errorf("diff command failed: %w", err)
	}
	// Non-exit error indicates command execution failure
	return false, fmt.Errorf("failed to execute diff command: %w", err)
}
