# Enhanced GrepTool - Test Data

This directory contains test files for the enhanced GrepTool that works like ripgrep.

## Test Files

- `sample.txt` - Plain text file with various test patterns
- `sample.js` - JavaScript file with test functions and classes
- `sample.ts` - TypeScript file with test interfaces, functions, and classes
- `binary.dat` - File with embedded binary data (null bytes)
- `.git/config` - Git config file (normally ignored)
- `.hidden/hidden.txt` - File in a hidden directory (normally ignored)

## Example Search Commands

Once the enhanced GrepTool is built, you can test it with these commands:

```
# Basic search for "test" in all files
./grep_tool --pattern "test" --path /path/to/testdata

# Search with context lines
./grep_tool --pattern "test" --path /path/to/testdata --context 2

# Search only in JavaScript files
./grep_tool --pattern "function" --path /path/to/testdata --include "*.js"

# Case-insensitive search
./grep_tool --pattern "TEST" --path /path/to/testdata --ignore_case

# Don't ignore version control directories
./grep_tool --pattern "user" --path /path/to/testdata --no_ignore_vcs
```

## Running Tests

You can run the automated tests with:

```
cd /path/to/cmd
go test -v
```

This will run all the unit tests and integration tests for the GrepTool functionality.