package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// InteractiveCLI provides an interactive interface for navigating file groups and viewing diffs.
type InteractiveCLI struct {
	groups    [][]string
	diffExec  *DiffExecutor
	scanner   *bufio.Scanner
	writer    io.Writer
}

// NewInteractiveCLI creates a new InteractiveCLI instance.
func NewInteractiveCLI(groups [][]string, diffExec *DiffExecutor) *InteractiveCLI {
	return &InteractiveCLI{
		groups:   groups,
		diffExec: diffExec,
		scanner:  bufio.NewScanner(os.Stdin),
		writer:   os.Stdout,
	}
}

// Run starts the interactive CLI session.
func (cli *InteractiveCLI) Run() error {
	if len(cli.groups) == 0 {
		fmt.Fprintf(cli.writer, "No groups of similar files found.\n")
		return nil
	}

	fmt.Fprintf(cli.writer, "Found %d group(s) of similar files.\n\n", len(cli.groups))

	for i, group := range cli.groups {
		if err := cli.handleGroup(i+1, group); err != nil {
			return err
		}
	}

	return nil
}

// handleGroup handles interaction for a single group of files.
func (cli *InteractiveCLI) handleGroup(groupNum int, group []string) error {
	fmt.Fprintf(cli.writer, "=== Group %d: %d files ===\n", groupNum, len(group))
	for i, file := range group {
		filename := filepath.Base(file)
		fmt.Fprintf(cli.writer, "  %d. %s\n", i+1, filename)
	}
	fmt.Fprintf(cli.writer, "\n")

	// Generate all pairs in the group
	pairs := cli.generatePairs(group)
	if len(pairs) == 0 {
		fmt.Fprintf(cli.writer, "No pairs to compare in this group.\n\n")
		return nil
	}

	// Create mapping from file indices to pair index
	pairMap := cli.createPairMap(group, pairs)

	fmt.Fprintf(cli.writer, "Available pairs to compare:\n")
	for i, pair := range pairs {
		file1Idx := cli.findFileIndex(group, pair[0])
		file2Idx := cli.findFileIndex(group, pair[1])
		fmt.Fprintf(cli.writer, "  %d. File %d vs File %d (%s vs %s)\n",
			i+1, file1Idx+1, file2Idx+1,
			filepath.Base(pair[0]), filepath.Base(pair[1]))
	}
	fmt.Fprintf(cli.writer, "\n")
	fmt.Fprintf(cli.writer, "Enter pair number (1-%d), file numbers (e.g., '2-3'), 'n' for next group, 'q' to quit: ", len(pairs))

	if !cli.scanner.Scan() {
		return cli.scanner.Err()
	}

	input := cli.scanner.Text()
	switch input {
	case "q", "Q":
		return fmt.Errorf("user quit")
	case "n", "N":
		return nil
	default:
		// Try to parse as file numbers (e.g., "2-3")
		var file1Num, file2Num int
		if n, _ := fmt.Sscanf(input, "%d-%d", &file1Num, &file2Num); n == 2 {
			// User specified file numbers
			if file1Num < 1 || file1Num > len(group) || file2Num < 1 || file2Num > len(group) || file1Num == file2Num {
				fmt.Fprintf(cli.writer, "Invalid file numbers. Please enter two different numbers between 1 and %d.\n\n", len(group))
				return cli.handleGroup(groupNum, group)
			}
			// Find the pair index
			key := fmt.Sprintf("%d-%d", file1Num-1, file2Num-1)
			if pairIdx, ok := pairMap[key]; ok {
				pair := pairs[pairIdx]
				if err := cli.showDiff(pair[0], pair[1]); err != nil {
					return err
				}
			} else {
				// Try reverse order
				key = fmt.Sprintf("%d-%d", file2Num-1, file1Num-1)
				if pairIdx, ok := pairMap[key]; ok {
					pair := pairs[pairIdx]
					if err := cli.showDiff(pair[1], pair[0]); err != nil {
						return err
					}
				} else {
					fmt.Fprintf(cli.writer, "No pair found for files %d and %d.\n\n", file1Num, file2Num)
					return cli.handleGroup(groupNum, group)
				}
			}
		} else {
			// Try to parse as pair number
			var pairNum int
			if _, err := fmt.Sscanf(input, "%d", &pairNum); err != nil || pairNum < 1 || pairNum > len(pairs) {
				fmt.Fprintf(cli.writer, "Invalid input. Please enter a pair number (1-%d) or file numbers (e.g., '2-3').\n\n", len(pairs))
				return cli.handleGroup(groupNum, group)
			}

			pair := pairs[pairNum-1]
			if err := cli.showDiff(pair[0], pair[1]); err != nil {
				return err
			}
		}

		fmt.Fprintf(cli.writer, "\nPress Enter to continue...")
		cli.scanner.Scan()

		// Ask if user wants to see another pair in this group
		fmt.Fprintf(cli.writer, "View another pair in this group? (y/n): ")
		if !cli.scanner.Scan() {
			return cli.scanner.Err()
		}
		response := cli.scanner.Text()
		if response == "y" || response == "Y" {
			return cli.handleGroup(groupNum, group)
		}
	}

	return nil
}

// generatePairs generates all unique pairs from a group of files.
func (cli *InteractiveCLI) generatePairs(group []string) [][]string {
	var pairs [][]string
	for i := 0; i < len(group); i++ {
		for j := i + 1; j < len(group); j++ {
			pairs = append(pairs, []string{group[i], group[j]})
		}
	}
	return pairs
}

// createPairMap creates a mapping from file index pairs to pair index.
// Key format: "i-j" where i < j are file indices.
func (cli *InteractiveCLI) createPairMap(group []string, pairs [][]string) map[string]int {
	pairMap := make(map[string]int)
	for pairIdx, pair := range pairs {
		file1Idx := cli.findFileIndex(group, pair[0])
		file2Idx := cli.findFileIndex(group, pair[1])
		// Ensure i < j for consistent key format
		if file1Idx < file2Idx {
			key := fmt.Sprintf("%d-%d", file1Idx, file2Idx)
			pairMap[key] = pairIdx
		} else {
			key := fmt.Sprintf("%d-%d", file2Idx, file1Idx)
			pairMap[key] = pairIdx
		}
	}
	return pairMap
}

// findFileIndex finds the index of a file in the group.
func (cli *InteractiveCLI) findFileIndex(group []string, file string) int {
	for i, f := range group {
		if f == file {
			return i
		}
	}
	return -1
}

// showDiff displays a side-by-side diff between two files.
func (cli *InteractiveCLI) showDiff(file1, file2 string) error {
	fmt.Fprintf(cli.writer, "\n--- Comparing ---\n")
	fmt.Fprintf(cli.writer, "File 1: %s\n", filepath.Base(file1))
	fmt.Fprintf(cli.writer, "File 2: %s\n", filepath.Base(file2))
	fmt.Fprintf(cli.writer, "---\n\n")

	diff, err := cli.diffExec.DiffSideBySide(file1, file2)
	if err != nil {
		return fmt.Errorf("failed to generate diff: %w", err)
	}

	fmt.Fprintf(cli.writer, "%s\n", diff)
	return nil
}
