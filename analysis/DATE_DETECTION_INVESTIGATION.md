# Date Detection Investigation Findings

## Executive Summary

All tests pass, but there is a **false positive issue** in the date detection logic. The `isLikelyDatePattern()` function incorrectly identifies some legitimate version patterns (like `file-1-2.txt`) as dates, causing them to be excluded when using the `--suffix` flag.

## Current Behavior

### How Date Detection Works

The `isLikelyDatePattern(baseFilename, baseName)` function uses three heuristics:

1. **Heuristic 1 (Trailing Hyphen+Digits)**: If removing the matched suffix leaves trailing hyphen+digits, treat as date
   - Example: `file-2026-01-30` → pattern matches `-30` → leaves `file-2026-01` → ends with `-01` → detected as date ✓

2. **Heuristic 2 (3+ Sequences)**: If filename has 3+ hyphen+digit sequences, treat as date
   - Example: `file-2026-01-30` has 3 sequences → detected as date ✓

3. **Heuristic 3 (4+ Digit Sequences)**: If any sequence has 4+ digits, treat as year
   - Example: `file-2024` has `-2024` (4 digits) → detected as date ✓

### The Problem: False Positives

**Problematic Case**: `file-1-2.txt` with pattern `-\d{1,2}$`

- Pattern matches `-2` at the end
- Removing `-2` leaves `file-1`
- `file-1` ends with `-1` (trailing hyphen+digits)
- **Result**: Detected as date → **EXCLUDED** ❌

This is incorrect because `file-1-2.txt` is clearly a version pattern, not a date pattern.

## Real-World Impact Analysis

### Scenarios Tested

#### ✅ Scenario 1: Simple Versioned Files (Works Correctly)
```
Files: document.txt, document-1.txt, document-2.txt
Pattern: -\d{1,2}$
Result: All 3 files included ✓
```

#### ✅ Scenario 2: Versioned Files with Date File (Works Correctly)
```
Files: document.txt, document-1.txt, document-2.txt, document-2024.txt
Pattern: -\d{1,2}$
Result: document.txt, document-1.txt, document-2.txt included
        document-2024.txt excluded (doesn't match pattern) ✓
```

#### ❌ Scenario 3: Nested Versions (False Positive)
```
Files: file.txt, file-1.txt, file-1-2.txt, file-2.txt
Pattern: -\d{1,2}$
Result: file.txt, file-1.txt, file-2.txt included
        file-1-2.txt EXCLUDED (incorrectly detected as date) ❌
```

#### ❌ Scenario 4: Real Date vs Nested Version (Both Excluded)
```
Files: report.txt, report-2026-01-30.txt, report-1-2.txt
Pattern: -\d{1,2}$
Result: All files excluded
        - report.txt: doesn't match pattern
        - report-2026-01-30.txt: correctly detected as date ✓
        - report-1-2.txt: incorrectly detected as date ❌
```

#### ❌ Scenario 5: Multiple Nested Versions (False Positives)
```
Files: project.txt, project-1.txt, project-1-2.txt, project-1-2-3.txt, project-2.txt
Pattern: -\d{1,2}$
Result: project.txt, project-1.txt, project-2.txt included
        project-1-2.txt EXCLUDED (incorrectly detected as date) ❌
        project-1-2-3.txt EXCLUDED (correctly detected as date via H2) ✓
```

#### ❌ Scenario 6: Two-Digit Nested Versions (False Positive)
```
Files: data.txt, data-10.txt, data-10-20.txt, data-20.txt
Pattern: -\d{1,2}$
Result: data.txt, data-10.txt, data-20.txt included
        data-10-20.txt EXCLUDED (incorrectly detected as date) ❌
```

## Detailed Analysis

### Heuristic 1 Breakdown

**Current Logic**: If `baseName` ends with `-\d+`, return `true` (date)

**Problem**: This heuristic is too aggressive. It catches legitimate version patterns:

| Filename | Pattern Match | baseName | Trailing? | Detected As | Correct? |
|----------|---------------|----------|-----------|-------------|----------|
| `file-1-2` | `-2` | `file-1` | Yes (`-1`) | Date | ❌ No |
| `file-2026-01-30` | `-30` | `file-2026-01` | Yes (`-01`) | Date | ✓ Yes |
| `file-10-20` | `-20` | `file-10` | Yes (`-10`) | Date | ❌ No |
| `file-1-2-3` | `-3` | `file-1-2` | Yes (`-2`) | Date | ❌ No |

**Key Insight**: The heuristic doesn't distinguish between:
- Date patterns: `-2026-01-30` → removing `-30` leaves `-2026-01` (date components)
- Version patterns: `-1-2` → removing `-2` leaves `-1` (version component)

### Heuristic 2 Breakdown

**Current Logic**: If filename has 3+ hyphen+digit sequences, return `true` (date)

**Analysis**: This works correctly:
- `file-2026-01-30` → 3 sequences → date ✓
- `file-1-2-3` → 3 sequences → date ✓ (correctly catches this)
- `file-1-2` → 2 sequences → not caught by H2 (relies on H1)

### Heuristic 3 Breakdown

**Current Logic**: If any sequence has 4+ digits, return `true` (year)

**Analysis**: This works correctly:
- `file-2024` → `-2024` has 4 digits → date ✓
- `file-1-2` → all sequences are 1 digit → not caught by H3

## Test Coverage Analysis

### Current Test Cases

Looking at `TestIsLikelyDatePattern`:

1. ✅ `file-2026-01-30` → `file-2026-01` → true (correct)
2. ✅ `file-2024` → `file` → true (correct)
3. ✅ `file-1` → `file` → false (correct)
4. ✅ `file-1-backup` → `file-1-backup` → false (correct)
5. ✅ `report-2026-01-30` → `report-2026-01` → true (correct)
6. ❌ `file-1-2` → `file-1` → true (test expects true, but this is wrong!)
7. ✅ `document-2026-01-15` → `document` → true (correct via H2)

**Issue**: Test case #6 expects `file-1-2` to be detected as a date, but this is a false positive. The test description says "conservatively treated as date", suggesting this was intentional, but it's problematic.

## Impact Assessment

### Severity: Medium

**Why Medium, not High:**
- The primary use case (`--suffix '-\d{1,2}'` for simple versioned files) works correctly
- Most users likely have simple patterns like `file-1.txt`, `file-2.txt` which work fine
- The false positive only affects nested version patterns like `file-1-2.txt`

**Why Not Low:**
- Users with nested version patterns will lose files unexpectedly
- No warning or indication that files were excluded
- Could lead to data loss if users rely on the tool to find duplicates

### Affected Use Cases

**Not Affected:**
- Simple versioned files: `file-1.txt`, `file-2.txt` ✓
- Date files: `file-2024.txt`, `file-2026-01-30.txt` ✓

**Affected:**
- Nested version patterns: `file-1-2.txt`, `file-10-20.txt` ❌
- Multi-level versions: `file-1-2.txt` (but `file-1-2-3.txt` is correctly excluded)

## Recommendations

### Option 1: Refine Heuristic 1 (Recommended)

**Approach**: Only treat as date if trailing sequence has 2 digits (month/day range)

**Logic**: 
- If `baseName` ends with `-\d{2}` (two digits), treat as date
- If `baseName` ends with `-\d{1}` (single digit), don't treat as date

**Rationale**: 
- Dates typically use 2-digit months/days (`-01`, `-30`)
- Single digits are more likely versions (`-1`, `-2`)
- Still catches `file-2026-01-30` → removing `-30` leaves `-2026-01` → ends with `-01` (2 digits) ✓

**Trade-off**: 
- Might miss some edge cases, but reduces false positives significantly

### Option 2: Require Multiple Heuristics

**Approach**: Require H1 AND (H2 OR H3) to be true

**Logic**: Only exclude if:
- H1 is true (trailing hyphen+digits) AND
- (H2 is true (3+ sequences) OR H3 is true (4+ digits))

**Rationale**: 
- More conservative approach
- `file-1-2` would pass: H1=true, but H2=false and H3=false → not excluded ✓

**Trade-off**: 
- Might allow some date patterns through if they don't meet H2 or H3

### Option 3: Check Digit Length in Trailing Sequence

**Approach**: In H1, check if trailing sequence is 2 digits (likely month/day)

**Logic**: 
- If `baseName` ends with `-\d{2}`, treat as date
- If `baseName` ends with `-\d{1}`, don't treat as date

**Implementation**:
```go
trailingMatch := trailingHyphenDigits.FindString(baseName)
if trailingMatch != "" {
    // Check if it's 2 digits (month/day) vs 1 digit (version)
    if len(trailingMatch) == 3 { // "-" + 2 digits
        return true
    }
    // Single digit trailing sequence is likely version, not date
}
```

## Next Steps

1. **Decide on approach**: Choose one of the options above
2. **Update test expectations**: Fix test case #6 in `TestIsLikelyDatePattern`
3. **Add test cases**: Add tests for nested version patterns to ensure they're included
4. **Update documentation**: Document the behavior for nested version patterns

## Questions for Review

1. **Is `file-1-2.txt` a legitimate version pattern** that should be included?
   - If yes → fix the heuristic
   - If no → update documentation to clarify this is intentional

2. **What is the intended behavior** for nested version patterns?
   - Should `file-1-2.txt` be treated differently from `file-1-2-3.txt`?

3. **How common are nested version patterns** in real-world usage?
   - If rare → lower priority
   - If common → higher priority fix
