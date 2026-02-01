package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	defaultMinPrefixLength = 3
)

func main() {
	var (
		diffTool      = flag.String("diff-tool", "", "Override default diff command (default: 'diff')")
		minPrefix     = flag.Int("min-prefix", defaultMinPrefixLength, "Minimum prefix length for grouping files")
		suffixPattern = flag.String("suffix", "", "Only consider files whose names match the indicated suffix pattern (regex)")
		showHelp      = flag.Bool("help", false, "Show usage information")
		showVersion   = flag.Bool("version", false, "Show version information")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [directory]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Scans a directory for files with similar names and provides an interactive interface\n")
		fmt.Fprintf(os.Stderr, "to compare them using side-by-side diffs.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nIf no directory is specified, the current directory is used.\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Println("doppel version 0.1.0")
		return
	}

	if *showHelp {
		flag.Usage()
		return
	}

	// Get directory from arguments or use current directory
	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	// Validate directory exists
	info, err := os.Stat(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", dir)
		os.Exit(1)
	}

	// Validate min prefix length
	if *minPrefix < 1 {
		fmt.Fprintf(os.Stderr, "Error: min-prefix must be at least 1\n")
		os.Exit(1)
	}

	// Compile suffix pattern if provided
	var compiledPattern *regexp.Regexp
	if *suffixPattern != "" {
		// Anchor pattern to end of string (before extension)
		patternStr := *suffixPattern + "$"
		pattern, err := regexp.Compile(patternStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid suffix pattern: %v\n", err)
			os.Exit(1)
		}
		compiledPattern = pattern
	}

	// Execute the workflow
	if err := run(dir, *diffTool, *minPrefix, compiledPattern); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the main workflow: scan, match, and interact.
func run(dir, diffTool string, minPrefix int, suffixPattern *regexp.Regexp) error {
	// Step 1: Scan directory
	scanner := NewScanner(dir)
	files, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	// Step 1.5: Filter files by suffix pattern if provided
	if suffixPattern != nil {
		files = filterFilesBySuffix(files, suffixPattern)
	}

	if len(files) < 2 {
		fmt.Println("Not enough files found to compare (need at least 2).")
		return nil
	}

	// Step 2: Group files by prefix
	matcher := NewMatcher(minPrefix)
	groups := matcher.Group(files)

	if len(groups) == 0 {
		fmt.Println("No groups of similar files found.")
		return nil
	}

	// Step 3: Interactive TUI
	diffExec := NewDiffExecutor(diffTool)
	m := initialModel(groups, diffExec)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

// filterFilesBySuffix filters files to include:
// 1. Files whose filename ends with a match to the given pattern
// 2. Base files (without the suffix pattern) that correspond to matching files
// If pattern is nil, returns all files (backward compatibility).
func filterFilesBySuffix(files []string, pattern *regexp.Regexp) []string {
	if pattern == nil {
		return files
	}

	// Compile regexes for checking date patterns (used to detect date patterns)
	// These are compiled once outside the loop for efficiency
	trailingHyphenDigits := regexp.MustCompile(`-\d+$`)
	hyphenDigitPattern := regexp.MustCompile(`-\d+`)

	// Step 1: Find files matching the suffix pattern and extract base names
	type fileMatch struct {
		file     string
		baseName string // filename without extension and without matched suffix
	}
	var matchingFiles []fileMatch
	baseNames := make(map[string]bool) // Track unique base names

	for _, file := range files {
		filename := filepath.Base(file)
		ext := filepath.Ext(filename)
		baseFilename := filename[:len(filename)-len(ext)]

		// Check if pattern matches at end of base filename
		if pattern.MatchString(baseFilename) {
			// Extract base name by removing the matched suffix
			// Use ReplaceAllString to remove the matched portion
			baseName := pattern.ReplaceAllString(baseFilename, "")
			
			// Additional check: if the base name ends with a hyphen followed by digits,
			// this suggests the match was part of a date pattern (e.g., "2026-01-30" where "-30" matched)
			// and we should exclude it. We want to match simple version patterns like "file-1" but
			// not date patterns like "file-2026-01-30".
			// Also check if the original filename has multiple hyphen+digit sequences, which indicates a date pattern
			if trailingHyphenDigits.MatchString(baseName) {
				// This is likely a date pattern - exclude it
				continue
			}
			
			// Additional check: if the original filename has 3+ hyphen+digit sequences, it's likely a date
			// Count hyphen+digit patterns in the original filename
			matches := hyphenDigitPattern.FindAllString(baseFilename, -1)
			if len(matches) >= 3 {
				// Multiple hyphen+digit sequences suggest a date pattern (e.g., "2026-01-30")
				continue
			}
			
			// Additional check: if any hyphen+digit sequence has 4+ digits, it's likely a year (e.g., "-2024")
			hasLongSequence := false
			for _, match := range matches {
				// match is like "-2024", check if it has 4+ digits (excluding the hyphen)
				if len(match) >= 5 { // "-" + 4+ digits
					hasLongSequence = true
					break
				}
			}
			if hasLongSequence {
				// Contains a 4+ digit sequence (likely a year) - exclude it
				continue
			}
			
			matchingFiles = append(matchingFiles, fileMatch{
				file:     file,
				baseName: baseName,
			})
			baseNames[baseName] = true
		}
	}

	// Step 2: Build result list with matching files and corresponding base files
	// Use a map to track included files and avoid duplicates
	included := make(map[string]bool)
	var result []string

	// Add all matching files
	for _, fm := range matchingFiles {
		if !included[fm.file] {
			result = append(result, fm.file)
			included[fm.file] = true
		}
	}

	// Add base files that correspond to matching files
	for _, file := range files {
		if included[file] {
			continue // Already included
		}

		filename := filepath.Base(file)
		ext := filepath.Ext(filename)
		baseFilename := filename[:len(filename)-len(ext)]

		// Check if this file's base name matches one of the extracted base names
		if baseNames[baseFilename] {
			result = append(result, file)
			included[file] = true
		}
	}

	return result
}
