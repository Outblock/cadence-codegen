package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/onflow/cadence/ast"
	"github.com/onflow/cadence/common"
	"github.com/onflow/cadence/parser"
)

// SimpleMemoryGauge implements common.MemoryGauge
type SimpleMemoryGauge struct{}

func (g *SimpleMemoryGauge) MeterMemory(size common.MemoryUsage) error {
	return nil
}

// Parameter represents a single parameter in a Cadence transaction or script
type Parameter struct {
	Name     string `json:"name"`
	TypeStr  string `json:"typeStr"`
	Optional bool   `json:"optional"`
}

// Import represents a Cadence import statement
type Import struct {
	Contract string `json:"contract"`
	Address  string `json:"address"`
}

// AnalysisResult represents the analysis result of a single Cadence file
type AnalysisResult struct {
	FileName   string      `json:"fileName"`
	Type       string      `json:"type"`
	Parameters []Parameter `json:"parameters"`
	ReturnType string      `json:"returnType,omitempty"`
	Imports    []Import    `json:"imports"`
}

// Report represents the complete analysis report
type Report struct {
	Transactions map[string]AnalysisResult `json:"transactions"`
	Scripts      map[string]AnalysisResult `json:"scripts"`
}

// Analyzer is responsible for analyzing Cadence files
type Analyzer struct {
	Transactions map[string]AnalysisResult
	Scripts      map[string]AnalysisResult
}

// New creates a new Analyzer instance
func New() *Analyzer {
	return &Analyzer{
		Transactions: make(map[string]AnalysisResult),
		Scripts:      make(map[string]AnalysisResult),
	}
}

// GetReport returns the current analysis report
func (a *Analyzer) GetReport() Report {
	return Report{
		Transactions: a.Transactions,
		Scripts:      a.Scripts,
	}
}

// extractImports extracts imports from the code and returns the code without imports
func extractImports(content []byte) ([]Import, []byte) {
	lines := strings.Split(string(content), "\n")
	var imports []Import
	var nonImportLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 4 && parts[2] == "from" {
				imports = append(imports, Import{
					Contract: parts[1],
					Address:  parts[3],
				})
			}
		} else {
			nonImportLines = append(nonImportLines, line)
		}
	}

	return imports, []byte(strings.Join(nonImportLines, "\n"))
}

// AnalyzeFile analyzes a single Cadence file and returns its analysis result
func (a *Analyzer) AnalyzeFile(filePath string) (*AnalysisResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	imports, codeWithoutImports := extractImports(content)

	fileName := filepath.Base(filePath)
	memoryGauge := &SimpleMemoryGauge{}
	program, err := parser.ParseProgram(memoryGauge, codeWithoutImports, parser.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Check for transaction declaration
	for _, declaration := range program.Declarations() {
		if transaction, ok := declaration.(*ast.TransactionDeclaration); ok {
			params := make([]Parameter, 0)
			if transaction.ParameterList != nil {
				for _, param := range transaction.ParameterList.Parameters {
					params = append(params, Parameter{
						Name:     param.Identifier.String(),
						TypeStr:  param.TypeAnnotation.String(),
						Optional: false,
					})
				}
			}
			result := &AnalysisResult{
				FileName:   fileName,
				Type:       "transaction",
				Parameters: params,
				Imports:    imports,
			}
			a.Transactions[fileName] = *result
			return result, nil
		}
	}

	// Check for script (main function)
	for _, declaration := range program.Declarations() {
		if function, ok := declaration.(*ast.FunctionDeclaration); ok {
			if function.Identifier.String() == "main" {
				params := make([]Parameter, 0)
				if function.ParameterList != nil {
					for _, param := range function.ParameterList.Parameters {
						params = append(params, Parameter{
							Name:     param.Identifier.String(),
							TypeStr:  param.TypeAnnotation.String(),
							Optional: false,
						})
					}
				}
				result := &AnalysisResult{
					FileName:   fileName,
					Type:       "script",
					Parameters: params,
					Imports:    imports,
				}
				if function.ReturnTypeAnnotation != nil {
					result.ReturnType = function.ReturnTypeAnnotation.Type.String()
				}
				a.Scripts[fileName] = *result
				return result, nil
			}
		}
	}

	// Check for public function that could serve as the main entry point
	for _, declaration := range program.Declarations() {
		if function, ok := declaration.(*ast.FunctionDeclaration); ok {
			if function.Access != ast.AccessNotSpecified {
				params := make([]Parameter, 0)
				if function.ParameterList != nil {
					for _, param := range function.ParameterList.Parameters {
						params = append(params, Parameter{
							Name:     param.Identifier.String(),
							TypeStr:  param.TypeAnnotation.String(),
							Optional: false,
						})
					}
				}
				result := &AnalysisResult{
					FileName:   fileName,
					Type:       "script",
					Parameters: params,
					Imports:    imports,
				}
				if function.ReturnTypeAnnotation != nil {
					result.ReturnType = function.ReturnTypeAnnotation.Type.String()
				}
				a.Scripts[fileName] = *result
				return result, nil
			}
		}
	}

	return nil, fmt.Errorf("no transaction or script found in file")
}

// AnalyzeDirectory analyzes all Cadence files in a directory and its subdirectories
func (a *Analyzer) AnalyzeDirectory(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".cdc" {
			if _, err := a.AnalyzeFile(path); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to analyze %s: %v\n", path, err)
			}
		}

		return nil
	})
}
