# Doppel

![Doppel](assets/doppel-which-is-which.png)

Ever had a sync conflict leave you with a sinking feeling?  Doppel is a CLI tool that finds groups of files with nearly the same name and gives you an interactive interface to compare them with side-by-side diffs.

## Overview

Doppel scans a directory for files that share common filename prefixes (e.g., `document.txt`, `document-1.txt`, `document_copy.txt`) and helps you identify duplicates or near-duplicates by comparing their contents with side-by-side diffs.

## Features

- **Prefix-based matching**: Groups files that share a common filename prefix
- **Suffix filtering**: Filter files by suffix pattern to focus on versioned files while excluding dates
- **Interactive TUI**: Navigate through groups and select files using a modern terminal UI (bubbletea)
- **Two-step file selection**: Pick two files one at a time for comparison
- **Side-by-side diffs**: Compare files using the system `diff` command
- **Configurable**: Adjust minimum prefix length, suffix pattern, and diff tool

## Installation

Build from source:

```bash
go build -o doppel .
```

## Usage

### Basic Usage

Scan the current directory:

```bash
./doppel
```

Scan a specific directory:

```bash
./doppel /path/to/directory
```

### Options

- `--diff-tool <command>`: Override the default diff command (default: `diff`)
- `--min-prefix <length>`: Minimum prefix length for grouping files (default: 3)
- `--suffix <pattern>`: Only consider files whose names match the indicated suffix pattern (regex). Files matching the pattern and their corresponding base files (without the suffix) are included. Useful for focusing on versioned files while excluding date suffixes.
- `--help`: Show usage information
- `--version`: Show version information

### Examples

Scan with custom minimum prefix length:

```bash
./doppel --min-prefix 5 /path/to/directory
```

Use a custom diff tool:

```bash
./doppel --diff-tool "git diff" /path/to/directory
```

Filter files by suffix pattern to focus on versioned files:

```bash
# Filter to files ending with hyphen + 1-2 digits (versions, not dates)
./doppel --suffix '-\d{1,2}' /path/to/directory

# Filter to files ending with space + digits
./doppel --suffix ' \d+' /path/to/directory
```

### Suffix Filtering

The `--suffix` flag allows you to focus on files with specific suffix patterns (like version numbers) while excluding files with date suffixes. The filter includes:

1. **Files matching the pattern**: Files whose filename ends with a match to the given regex pattern (e.g., `document-1.txt`, `document-2.txt`)
2. **Base files**: Files without the suffix pattern that correspond to matching files (e.g., `document.txt` is included when `document-1.txt` matches)

**Example**: With files `document.txt`, `document-1.txt`, `document-2.txt`, and `document-2026-01-30.txt`:

- Using `--suffix '-\d{1,2}'` will include `document.txt`, `document-1.txt`, and `document-2.txt` (grouped together)
- The date file `document-2026-01-30.txt` is excluded because it doesn't match the 1-2 digit pattern

**Common patterns**:
- `'-\d{1,2}'` - Hyphen + 1-2 digits (versions like `-1`, `-2`, excludes years like `-2024`)
- `' \d+'` - Space + one or more digits (versions like ` 1`, ` 2`)
- `'-\d+'` - Hyphen + one or more digits (includes dates, use with caution)

## How It Works

1. **Scan**: The tool scans the specified directory (non-recursive) for all files
2. **Filter** (optional): If `--suffix` is provided, files are filtered to include only those matching the suffix pattern and their corresponding base files
3. **Match**: Files are grouped by common filename prefixes
4. **Compare**: You can interactively select file pairs to compare using side-by-side diffs

### Interactive TUI

When similar files are found, you'll see an interactive terminal UI with:

1. **Group Selection**: A list of groups showing the filenames in each group:
   ```
   Found 3 group(s) of similar files

   >  Group 1: 3 files
       document-1.txt, document.txt, document_copy.txt

      Group 2: 2 files
       image-1.png, image.png

      Group 3: 3 files
       report-2024.txt, report.txt, report_backup.txt
   ```

2. **First File Selection**: After selecting a group, choose the first file to compare

3. **Second File Selection**: Choose the second file (the first file is automatically skipped in navigation)

4. **Diff View**: The side-by-side diff is automatically displayed after selecting both files

#### Keyboard Controls

- **↑/↓ or j/k**: Navigate up/down through items
- **Enter**: Select the current item
- **Esc**: Go back to the previous screen
- **q**: Quit the application
- **n**: (In group selection) Move to the next group

## Requirements

- Go 1.16 or later (for building)
- Unix-like system with `diff` command available
- Terminal that supports ANSI escape codes (for the TUI)

## Testing

### Automated Tests

Run all tests:

```bash
go test -v ./...
```

### Manual Testing

The project includes a `testdata/` directory with sample files for manual testing. This directory contains:

- **document.txt**, **document-1.txt**, **document_copy.txt** - Three similar files with lorem ipsum content showing slight variations
- **image.png**, **image-1.png** - Placeholder image files (as text) for testing
- **report.txt**, **report-2024.txt**, **report_backup.txt** - Report files with similar names
- **unrelated.txt** - A file that should not be grouped with others

To test the tool manually:

```bash
./doppel testdata
```

This will demonstrate the tool's ability to:
- Group files by common prefix (e.g., `document*`, `image*`, `report*`)
- Show an interactive TUI with groups and their filenames
- Navigate through groups and files using keyboard controls
- Select two files for comparison
- Display side-by-side diffs showing the variations between files

When run against `testdata/`, you should see 3 groups:
1. **document** group: `document.txt`, `document-1.txt`, `document_copy.txt` (3 files)
2. **image** group: `image.png`, `image-1.png` (2 files)
3. **report** group: `report.txt`, `report-2024.txt`, `report_backup.txt` (3 files)

The `unrelated.txt` file will not be grouped with any others, demonstrating that only files with common prefixes are matched.

## Project Structure

```
doppel/
├── main.go              # Entry point, CLI argument parsing
├── scanner.go           # Directory scanning logic
├── scanner_test.go      # Unit tests for scanner
├── matcher.go           # Prefix-based filename matching
├── matcher_test.go      # Unit tests for matcher
├── diff.go              # External diff command execution
├── diff_test.go         # Unit tests for diff executor
├── tui.go               # Interactive TUI interface (bubbletea)
├── interactive.go       # Legacy interactive CLI interface (deprecated)
├── interactive_test.go  # Unit tests for interactive CLI
├── integration_test.go  # Integration tests for common code paths
├── filter_test.go       # Unit tests for suffix filtering
├── assets/              # Project assets (e.g. hero image)
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── project_context.md   # Project preferences and context
└── README.md            # This file
```

## Development

This project follows these principles:

- **Modularity**: Each component is independently testable
- **Test-Driven Development**: Tests are written alongside implementation
- **Executable Documentation**: Tests serve as documentation
- **Iterative Development**: Work in small, incremental steps

See `project_context.md` for more details on development preferences.

## Future Enhancements

- **Auto-diff for pairs**: When a file group contains exactly two files, automatically skip to diff mode instead of requiring file selection
- **Colorized diff output**: Use color coding in the diff display to better differentiate additions, deletions, and changes
- **Prefix filter flag**: Add a command-line flag (e.g., `--prefix`) to filter files by a specific prefix before grouping, allowing users to focus on files matching a particular pattern

## License

GPHL License

Copyright (c) 2025 Doppel contributors

All code is covered under the GPHL, or Giant Pile of Hacks License, which
guarantees nothing more than hacks, which may or may not be piled up to
any specific size, gigantic or otherwise.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
