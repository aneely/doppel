---
name: Suffix Pattern Specification
overview: Add a --suffix command-line flag to filter files by suffix pattern before grouping, allowing users to focus on files matching specific patterns (like hyphen/space + digits for versions) while excluding files with date suffixes.
todos:
  - id: cli-flag
    content: Add --suffix CLI flag in main.go with parsing and validation
    status: pending
  - id: filter-function
    content: Implement filterFilesBySuffix() helper function that: (1) finds files matching suffix pattern, (2) extracts base names from matching files, (3) includes base files corresponding to extracted base names
    status: pending
  - id: integrate-filtering
    content: Integrate suffix filtering in run() function between Scan() and Group()
    status: pending
  - id: unit-tests
    content: Add unit tests for filterFilesBySuffix() function
    status: pending
  - id: integration-tests
    content: Add integration tests for suffix filtering workflow
    status: pending
  - id: documentation
    content: Update README.md with --suffix option documentation and examples
    status: pending
isProject: false
---

# Suffix Pattern Filter Proposal (Issue #7)

## Problem Statement

Users want to focus on files with specific suffix patterns (like `-1`, `-2`,  `1`,  `2`) that indicate duplicates or versions, while excluding files with date suffixes (like `-2024`). Currently, all files are considered for grouping, which can create unwanted groups mixing versioned files with dated files.

## Proposed Solution

### Approach: Pre-filtering by Suffix Pattern with Base File Inclusion

Add a `--suffix` command-line flag that filters the file list **before** grouping. The filter includes:
1. Files whose filename ends with a match to the given pattern (e.g., `document-1.txt`, `document-2.txt`)
2. Base files (without the suffix pattern) that share a prefix with files matching the pattern (e.g., `document.txt`)

This allows users to:

1. **Focus on version patterns**: Use `--suffix '-\d{1,2}'` to match `-1`, `-2`, etc., but exclude `-2024`
2. **Include base files**: Files like `document.txt` are included when `document-1.txt` matches the pattern
3. **Exclude dates**: By using restrictive patterns (1-2 digits), 4-digit years are automatically excluded
4. **Focus on space patterns**: Use `--suffix ' \d+'` to match `file 1`, `file 2`, and include `file.txt`

### Pattern Specification Format

**Recommended: Regex-based patterns**

- Accept a regex pattern via `--suffix` flag
- Pattern matches against the end of the filename (before extension)
- Supports both regex patterns and named shortcuts for common cases

**Named shortcuts (optional enhancement)**:

- `sn` or `space-number`: Matches space + digits (e.g., `file 1.txt`)
- `hn` or `hyphen-number`: Matches hyphen + digits (e.g., `file-1.txt`)
- `hn-small`: Matches hyphen + 1-2 digits (excludes years like `-2024`)

**Recommendation**: Start with regex-only for simplicity, add named shortcuts if users request them.

### Implementation Design

#### 1. Pattern Parsing and Compilation

- Add `suffixPattern *regexp.Regexp` variable in `main.go`
- Parse `--suffix` flag: `flag.String("suffix", "", "Only consider files whose names match the indicated suffix pattern (regex)")`
- Compile regex pattern at CLI parsing time
- Validate pattern syntax before proceeding

#### 2. File Filtering Function

- Create `filterFilesBySuffix(files []string, pattern *regexp.Regexp) []string` function
- If pattern is nil/empty, return all files (backward compatibility)
- **Step 1**: Find all files matching the suffix pattern
  - For each file, extract filename using `filepath.Base()`
  - Strip extension, check if pattern matches at end of filename
  - Collect matching files and their base names (filename with suffix removed)
- **Step 2**: Include base files
  - For each base name extracted from matching files, find files whose base name (without extension) exactly matches
  - Include both: files matching the pattern AND base files that correspond to them
- Return the combined filtered list

#### 3. Integration Point

- In `run()` function, filter files after `scanner.Scan()` and before `matcher.Group()`
- Filtered list is passed to `matcher.Group()` - grouping logic unchanged
- Error handling: if no files match pattern, show "Not enough files" or "No groups" message

#### 4. CLI Integration

- Add `suffix` flag in `main.go`
- Parse and compile pattern before calling `run()`
- Pass compiled pattern (or nil) to `run()` function

### Default Behavior

**Without `--suffix`**: Current behavior (all files considered)
**With `--suffix`**: Files matching suffix pattern AND their corresponding base files (without suffix) are considered

### Example Usage

```bash
# Filter to files ending with hyphen + 1-2 digits (versions, not dates)
./doppel --suffix '-\d{1,2}' testdata

# Filter to files ending with space + digits
./doppel --suffix ' \d+' testdata

# Filter to files ending with hyphen + any digits (includes dates)
./doppel --suffix '-\d+' testdata

# Using named shortcut (if implemented)
./doppel --suffix hn-small testdata
```

### Expected Behavior Examples

**Scenario 1: Version files with base file inclusion**

- Files in directory: `document.txt`, `document-1.txt`, `document-2.txt`, `document-2026-01-30.txt`, `report-2024.txt`
- Command: `./doppel --suffix '-\d{1,2}' testdata`
- **Step 1 - Pattern matches**: `document-1.txt`, `document-2.txt` (matches `-\d{1,2}`)
- **Step 2 - Extract base names**: `document` (from both `document-1.txt` and `document-2.txt`)
- **Step 3 - Include base files**: `document.txt` (base name matches `document`)
- **Filtered files**: `document.txt`, `document-1.txt`, `document-2.txt`
- **Excluded files**: 
  - `document-2026-01-30.txt` (doesn't match `-\d{1,2}` pattern - too many digits)
  - `report-2024.txt` (matches pattern but different prefix, and no `report.txt` base file)
- **Result**: One group containing `document.txt`, `document-1.txt`, and `document-2.txt` ✓

**Scenario 2: Space pattern with base file**

- Files: `file.txt`, `file 1.txt`, `file 2.txt`, `file-backup.txt`
- Command: `./doppel --suffix ' \d+' testdata`
- **Pattern matches**: `file 1.txt`, `file 2.txt`
- **Extract base names**: `file` (from both)
- **Include base files**: `file.txt` (base name matches `file`)
- **Filtered files**: `file.txt`, `file 1.txt`, `file 2.txt`
- **Result**: One group with `file.txt`, `file 1.txt`, and `file 2.txt` ✓

**Scenario 3: Excluding dates - no base file**

- Files: `report.txt`, `report-1.txt`, `report-2024.txt`
- Command: `./doppel --suffix '-\d{1,2}' testdata`
- **Pattern matches**: `report-1.txt` (matches `-\d{1,2}`)
- **Extract base names**: `report` (from `report-1.txt`)
- **Include base files**: `report.txt` (base name matches `report`)
- **Filtered files**: `report.txt`, `report-1.txt`
- **Excluded**: `report-2024.txt` (doesn't match `-\d{1,2}` - 4 digits)
- **Result**: One group with `report.txt` and `report-1.txt` ✓

**Scenario 4: Date file excluded even with matching prefix**

- Files: `document.txt`, `document-1.txt`, `document-2026-01-30.txt`
- Command: `./doppel --suffix '-\d{1,2}' testdata`
- **Pattern matches**: `document-1.txt` only (`document-2026-01-30.txt` doesn't match `-\d{1,2}`)
- **Extract base names**: `document` (from `document-1.txt`)
- **Include base files**: `document.txt` (base name matches `document`)
- **Filtered files**: `document.txt`, `document-1.txt`
- **Excluded**: `document-2026-01-30.txt` (doesn't match pattern)
- **Result**: One group with `document.txt` and `document-1.txt` ✓

**Scenario 5: Without flag**

- Files: `document.txt`, `document-1.txt`, `report-2024.txt`
- Command: `./doppel testdata`
- **Filtered files**: All files (no filtering)
- **Result**: Groups as before (current behavior) ✓

### Files to Modify

1. `**main.go**`:
   - Add `suffix` flag parsing: `flag.String("suffix", "", "Only consider files whose names match the indicated suffix pattern (regex)")`
   - Compile regex pattern (validate syntax)
   - Pass pattern to `run()` function
   - Update help text
2. `**main.go` (run function)**:
   - Add `suffixPattern *regexp.Regexp` parameter to `run()`
   - Call `filterFilesBySuffix()` after `scanner.Scan()` and before `matcher.Group()`
   - Handle empty filtered list gracefully
3. **New helper function** (in `main.go` or separate file):
   - `filterFilesBySuffix(files []string, pattern *regexp.Regexp) []string`
   - Extract filename, strip extension, match pattern at end
   - Return filtered list
4. `**matcher_test.go**` (or new `filter_test.go`):
   - Add tests for `filterFilesBySuffix()`
   - Test various regex patterns
   - Test edge cases (no extension, multiple dots, empty pattern, etc.)
5. `**integration_test.go**`:
   - Add test for full workflow with `--suffix` flag
   - Verify filtering works correctly
   - Verify grouping still works on filtered list
6. `**README.md**`:
   - Document `--suffix` option
   - Add usage examples
   - Explain pattern format and common use cases

### Implementation Considerations

1. **Pattern Matching**: Pattern should match at the **end** of the filename (before extension), not anywhere in the name
2. **Extension Handling**: Strip extension before matching, but pattern applies to base filename
3. **Regex Anchoring**: Use `$` anchor to ensure pattern matches at end: compile pattern as `pattern + "$"` or use `MatchString()` with end anchor
4. **Base Name Extraction**: When a file matches the pattern, extract the base name by removing the matched suffix. Use regex `FindStringSubmatch()` or `ReplaceAllString()` to remove the matched portion
5. **Base File Matching**: For each extracted base name, find files whose base name (without extension) exactly equals the extracted base name. Use exact string matching, not prefix matching (prefix matching happens later in grouping)
6. **Empty Pattern**: If pattern is empty/nil, return all files (backward compatibility)
7. **Error Handling**: Invalid regex should be caught at flag parsing time with clear error message
8. **Performance**: Filtering happens once after scan, minimal performance impact. Base file lookup can use a map for O(1) lookups
9. **Duplicate Handling**: Ensure each file appears only once in the filtered list (use a map/set to track included files)

### Testing Strategy

1. **Unit Tests** (new `filter_test.go` or in `main_test.go`):
   - Test `filterFilesBySuffix()` with various patterns
   - Test that patterns match at end of filename only
   - Test extension handling (files with/without extensions)
   - Test base file inclusion: verify base files are included when versioned files match
   - Test that base files are only included if they correspond to matching files
   - Test edge cases: empty pattern, no matches, all matches
   - Test regex patterns: `-\d+`, `-\d{1,2}`, ` \d+`, etc.
   - Test date exclusion: verify files with date-like suffixes are excluded when using restrictive patterns
2. **Integration Tests** (`integration_test.go`):
   - Test full workflow: scan → filter → group
   - Verify filtering excludes unwanted files (dates, different prefixes)
   - Verify base files are included in groups with versioned files
   - Verify grouping works correctly on filtered list
   - Test that "No groups" message appears when filter results in < 2 files
   - Test Scenario 1: `document.txt`, `document-1.txt`, `document-2.txt` grouped together
   - Test date exclusion: `document-2026-01-30.txt` excluded from group
3. **Manual Testing**:
   - Test with `testdata/` directory
   - Verify `--suffix '-\d{1,2}'` excludes `report-2024.txt`
   - Verify `--suffix ' \d+'` matches space-number patterns
   - Verify TUI still works correctly with filtered groups

### Addressing Version vs Date Distinction

The `--suffix` flag addresses the version/date distinction concern by allowing users to specify restrictive patterns:

- **Version patterns**: `--suffix '-\d{1,2}'` matches `-1`, `-2`, `-99` but excludes `-2024` (4 digits)
- **Date exclusion**: By limiting to 1-2 digits, 4-digit years are automatically excluded
- **Base file inclusion**: Files without the suffix (like `document.txt`) are included when versioned files (like `document-1.txt`) match the pattern
- **Flexibility**: Users can adjust pattern as needed (e.g., `-\d{1,3}` for 1-3 digits)

This approach is simpler than normalization and gives users direct control over which files are considered, while ensuring base files are grouped with their versioned counterparts.

### Open Questions

1. **Named shortcuts**: Should we implement named shortcuts (`sn`, `hn`, `hn-small`) or start with regex-only?
   - **Recommendation**: Start with regex-only, add shortcuts if users request them
2. **Pattern location**: Should pattern match anywhere in suffix or strictly at the end?
   - **Recommendation**: Strictly at the end (before extension) as per issue description
3. **Multiple patterns**: Should we support multiple patterns (OR logic)?
   - **Recommendation**: Start with single pattern, can add multiple patterns later if needed
4. **Case sensitivity**: Should patterns be case-sensitive?
   - **Recommendation**: Yes (standard regex behavior), but document it

### Alternative Approaches Considered

1. **Normalization approach** (original proposal): Strip suffixes before matching. More complex, changes grouping logic.
2. **Post-filtering**: Filter groups after grouping. Less efficient, doesn't address the core need.
3. **Separate date detection**: Automatically detect dates. Less flexible, harder to maintain.

**Chosen approach** (pre-filtering) is simpler, more flexible, and matches issue #7 exactly.

---

# Corrective Plan: Regex Anchoring and Date Exclusion Issues

## Issues Identified

### High Severity: Regex Anchoring Problem

**Location**: `main.go:73-74, 157`

**Problem**: 

- Line 73 adds `$` anchor: `patternStr := *suffixPattern + "$"`
- Line 157 uses `MatchString()` which doesn't require end anchoring
- If user provides `-\d{1,2}$`, we get `-\d{1,2}$$` (double anchor)
- Even with single `$`, `MatchString()` can match anywhere (e.g., `file-1-backup` matches `-1`)

**Root Cause**: `MatchString()` checks if pattern matches anywhere in the string, not requiring it to match at the end. The `$` anchor in the pattern helps, but:

1. Users might already include `$` in their pattern
2. We need to verify the match position is actually at the end

### Medium Severity: Fragile Date Exclusion Heuristics

**Location**: `main.go:153-181`

**Problem**:

- Complex, undocumented heuristics for detecting date patterns
- Multiple checks: trailing hyphens, sequence counting, length checking
- Hard to understand and maintain
- No clear documentation of what constitutes a "date pattern"

## Proposed Solutions

### Fix 1: Correct Regex Anchoring (High Priority)

**Approach**: Use `FindStringIndex()` or `FindStringSubmatchIndex()` to verify the match is at the end of the string, rather than relying solely on `MatchString()`.

**Implementation**:

1. **Normalize pattern input** (`main.go:71-79`):
  - Check if user-provided pattern already ends with `$`
  - If not, append `$` anchor
  - If yes, use pattern as-is (avoid double `$$`)
2. **Fix matching logic** (`main.go:157`):
  - Replace `pattern.MatchString(baseFilename)` 
  - Use `pattern.FindStringIndex(baseFilename)` to get match position
  - Verify match ends at `len(baseFilename)` (end of string)
  - Only proceed if match is anchored at end

**Code Changes**:

```go
// In main.go:71-79 - Normalize pattern
if *suffixPattern != "" {
    patternStr := *suffixPattern
    // Only add $ if pattern doesn't already end with it
    if !strings.HasSuffix(patternStr, "$") {
        patternStr = patternStr + "$"
    }
    pattern, err := regexp.Compile(patternStr)
    // ... rest of compilation
}

// In main.go:157 - Fix matching
// Replace: if pattern.MatchString(baseFilename)
// With:
match := pattern.FindStringIndex(baseFilename)
if match != nil && match[1] == len(baseFilename) {
    // Match is anchored at end - proceed
}
```

**Testing**:

- Test with pattern `-\d{1,2}` (should add `$`)
- Test with pattern `-\d{1,2}$` (should not double `$`)
- Test `file-1-backup` with pattern `-\d{1,2}` (should NOT match)
- Test `file-1` with pattern `-\d{1,2}` (should match)
- Verify existing tests still pass

### Fix 2: Document and Refactor Date Exclusion (Medium Priority)

**Approach**: Document the heuristics clearly and consider simplifying or extracting to a helper function.

**Options**:

**Option A: Document and Extract** (Recommended)

- Extract date detection logic to a helper function `isLikelyDatePattern(baseFilename string) bool`
- Add comprehensive documentation explaining the heuristics
- Keep current logic but make it more maintainable

**Option B: Simplify Heuristics**

- Remove complex checks, rely on pattern restrictiveness (e.g., `-\d{1,2}` naturally excludes `-2024`)
- Only keep essential checks if pattern is too permissive

**Recommendation**: Option A - document and extract, as the heuristics handle edge cases that pattern restrictiveness alone might miss.

**Implementation**:

1. Create helper function `isLikelyDatePattern()` in `main.go`
2. Document each heuristic with examples
3. Replace inline logic with function call
4. Add unit tests for date detection edge cases

**Code Structure**:

```go
// isLikelyDatePattern checks if a filename base (without extension) 
// appears to be a date pattern rather than a version pattern.
// 
// Heuristics used:
// 1. If removing the matched suffix leaves trailing hyphen+digits, 
//    it's likely part of a date (e.g., "2026-01-30" where "-30" matched)
// 2. If filename has 3+ hyphen+digit sequences, it's likely a date
// 3. If any hyphen+digit sequence has 4+ digits, it's likely a year
//
// Examples:
//   - "file-2026-01-30" -> true (multiple sequences)
//   - "file-2024" -> true (4+ digit sequence)
//   - "file-1" -> false (single sequence, short)
func isLikelyDatePattern(baseFilename string) bool {
    // Extract current heuristics logic here
}
```

**Testing**:

- Test `file-2026-01-30` (should be detected as date)
- Test `file-2024` (should be detected as date)
- Test `file-1` (should NOT be detected as date)
- Test `file-1-backup` (edge case - should NOT be date)

## Files to Modify

1. `**main.go**`:
  - Fix pattern normalization (lines 71-79)
  - Fix matching logic (line 157)
  - Extract date detection to helper function (lines 153-181)
  - Add documentation
2. `**filter_test.go**`:
  - Add tests for double `$` anchor case
  - Add tests for `file-1-backup` not matching `-\d{1,2}`
  - Add tests for date detection helper function

## Implementation Order

1. **Phase 1: Fix Regex Anchoring** (High Priority)
  - Normalize pattern input (prevent double `$`)
  - Fix matching to verify end position
  - Add tests
  - Verify existing tests pass
2. **Phase 2: Document Date Exclusion** (Medium Priority)
  - Extract to helper function
  - Add comprehensive documentation
  - Add tests for date detection
  - Verify existing tests pass

## Risk Assessment

**Low Risk**: 

- Changes are localized to `filterFilesBySuffix()` function
- Existing tests provide good coverage
- Can verify backward compatibility with test suite

**Testing Strategy**:

- Run full test suite: `go test ./...`
- Add new tests for edge cases
- Manual testing with `testdata/` directory
- Verify CLI behavior with various patterns

## Notes

- The low severity issues (regex recompilation, filename ending in `.`) can be addressed in a follow-up if needed
- Focus on high and medium severity items first
- Ensure backward compatibility with existing `--suffix` usage patterns

## Implementation Status

**Completed**: All fixes have been implemented and tested:

- ✅ Fixed pattern normalization to prevent double `$` anchor
- ✅ Fixed matching logic to use `FindStringIndex()` and verify match is at end
- ✅ Extracted date detection to `isLikelyDatePattern()` helper function
- ✅ Added comprehensive documentation for date detection heuristics
- ✅ Added tests for anchoring fixes (`TestFilterFilesBySuffix_PatternWithAnchor`, `TestFilterFilesBySuffix_AnchoredMatch`)
- ✅ Added tests for date detection (`TestIsLikelyDatePattern` with 7 test cases)

All tests pass successfully.
