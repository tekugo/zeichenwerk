package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CodebaseMetrics holds the analysis results for a Go codebase
type CodebaseMetrics struct {
	TotalFiles         int // Total number of .go files
	CodeFiles          int // Number of non-test .go files
	TestFiles          int // Number of *_test.go files
	TotalLines         int // Total lines across all files
	CodeLines          int // Lines containing actual code (excluding comments and blanks)
	DocumentationLines int // Lines containing comments
	BlankLines         int // Empty or whitespace-only lines
	TestCodeLines      int // Lines of code in test files
}

// ProductionCodeLines returns the number of production code lines (excluding tests)
func (m CodebaseMetrics) ProductionCodeLines() int {
	return m.CodeLines - m.TestCodeLines
}

// DocumentationRatio returns the percentage of lines that are documentation
func (m CodebaseMetrics) DocumentationRatio() float64 {
	if m.TotalLines == 0 {
		return 0
	}
	return float64(m.DocumentationLines) * 100 / float64(m.TotalLines)
}

// TestCoverageRatio returns the percentage of lines that are test code
func (m CodebaseMetrics) TestCoverageRatio() float64 {
	if m.TotalLines == 0 {
		return 0
	}
	return float64(m.TestCodeLines) * 100 / float64(m.TotalLines)
}

// String returns a formatted string representation of the metrics
func (m CodebaseMetrics) String() string {
	productionCode := m.ProductionCodeLines()
	if m.TotalLines == 0 {
		return "No Go files found"
	}
	
	return fmt.Sprintf(`=== CODEBASE ANALYSIS ===
Total Go files: %d
Code files: %d
Test files: %d

=== LINE BREAKDOWN ===
Total lines: %d
Production code: %d lines (%.1f%%)
Test code: %d lines (%.1f%%)
Documentation: %d lines (%.1f%%)
Blank lines: %d lines (%.1f%%)`,
		m.TotalFiles, m.CodeFiles, m.TestFiles,
		m.TotalLines,
		productionCode, float64(productionCode)*100/float64(m.TotalLines),
		m.TestCodeLines, m.TestCoverageRatio(),
		m.DocumentationLines, m.DocumentationRatio(),
		m.BlankLines, float64(m.BlankLines)*100/float64(m.TotalLines))
}

// AnalyzeGoCodebase analyzes a Go codebase starting from the given root directory
// and returns detailed metrics about the code structure.
//
// The function recursively walks through all subdirectories and analyzes .go files,
// categorizing them as either regular code files or test files (*_test.go).
// It counts lines of code, documentation (comments), and blank lines.
//
// Parameters:
//   - rootDir: The root directory to start the analysis from
//
// Returns:
//   - *CodebaseMetrics: Detailed metrics about the codebase
//   - error: Any error encountered during analysis
func AnalyzeGoCodebase(rootDir string) (*CodebaseMetrics, error) {
	metrics := &CodebaseMetrics{}
	
	// Regular expressions for identifying different types of lines
	commentRegex := regexp.MustCompile(`^\s*(//|/\*|\*|.*\*/)`)
	blankRegex := regexp.MustCompile(`^\s*$`)
	
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		
		// Count files
		metrics.TotalFiles++
		isTestFile := strings.HasSuffix(path, "_test.go")
		if isTestFile {
			metrics.TestFiles++
		} else {
			metrics.CodeFiles++
		}
		
		// Analyze file content
		fileMetrics, err := analyzeFile(path, commentRegex, blankRegex)
		if err != nil {
			return fmt.Errorf("failed to analyze file %s: %w", path, err)
		}
		
		// Update totals
		metrics.TotalLines += fileMetrics.totalLines
		metrics.DocumentationLines += fileMetrics.docLines
		metrics.BlankLines += fileMetrics.blankLines
		
		// Calculate code lines for this file
		codeLines := fileMetrics.totalLines - fileMetrics.docLines - fileMetrics.blankLines
		metrics.CodeLines += codeLines
		
		// If it's a test file, add to test code lines
		if isTestFile {
			metrics.TestCodeLines += codeLines
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", rootDir, err)
	}
	
	return metrics, nil
}

// fileMetrics holds metrics for a single file
type fileMetrics struct {
	totalLines int
	docLines   int
	blankLines int
}

// analyzeFile analyzes a single Go file and returns its metrics
func analyzeFile(filePath string, commentRegex, blankRegex *regexp.Regexp) (*fileMetrics, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	metrics := &fileMetrics{}
	scanner := bufio.NewScanner(file)
	inBlockComment := false
	
	for scanner.Scan() {
		line := scanner.Text()
		metrics.totalLines++
		
		trimmed := strings.TrimSpace(line)
		
		// Check for blank lines first
		if blankRegex.MatchString(line) {
			metrics.blankLines++
			continue
		}
		
		// Handle block comments - start of block comment without end
		if strings.Contains(trimmed, "/*") && !strings.Contains(trimmed, "*/") {
			inBlockComment = true
			metrics.docLines++
			continue
		}
		
		// Handle block comments - continuation or end
		if inBlockComment {
			metrics.docLines++
			if strings.Contains(trimmed, "*/") {
				inBlockComment = false
			}
			continue
		}
		
		// Handle single-line comments and complete block comments on one line
		if commentRegex.MatchString(line) {
			metrics.docLines++
			continue
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return metrics, nil
}

// Example usage function
func main() {
	// Default to current directory if no argument provided
	rootDir := "."
	if len(os.Args) >= 2 {
		rootDir = os.Args[1]
	}
	
	fmt.Printf("Analyzing Go codebase in: %s\n\n", rootDir)
	
	// Analyze the codebase
	metrics, err := AnalyzeGoCodebase(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing codebase: %v\n", err)
		os.Exit(1)
	}
	
	// Print the results
	fmt.Println(metrics)
	
	// Additional metrics
	fmt.Printf("\n=== ADDITIONAL METRICS ===\n")
	fmt.Printf("Production code lines: %d\n", metrics.ProductionCodeLines())
	fmt.Printf("Documentation ratio: %.1f%%\n", metrics.DocumentationRatio())
	fmt.Printf("Test coverage ratio: %.1f%%\n", metrics.TestCoverageRatio())
}