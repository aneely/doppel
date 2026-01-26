package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	defaultMinPrefixLength = 3
)

func main() {
	var (
		diffTool      = flag.String("diff-tool", "", "Override default diff command (default: 'diff')")
		minPrefix     = flag.Int("min-prefix", defaultMinPrefixLength, "Minimum prefix length for grouping files")
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

	// Execute the workflow
	if err := run(dir, *diffTool, *minPrefix); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run executes the main workflow: scan, match, and interact.
func run(dir, diffTool string, minPrefix int) error {
	// Step 1: Scan directory
	scanner := NewScanner(dir)
	files, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
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
