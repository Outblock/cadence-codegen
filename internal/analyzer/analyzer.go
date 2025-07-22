package analyzer

import (
	"encoding/base64"
	"encoding/json"
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

// Field represents a field in a struct
type Field struct {
	Name     string `json:"name"`
	TypeStr  string `json:"typeStr"`
	Optional bool   `json:"optional"`
	Access   string `json:"access"`
}

// Struct represents a Cadence struct declaration
type Struct struct {
	Name     string  `json:"name"`
	Fields   []Field `json:"fields"`
	Access   string  `json:"access"`
	FileName string  `json:"fileName"`
}

// AnalysisResult represents the analysis result of a single Cadence file
type AnalysisResult struct {
	FileName   string      `json:"fileName"`
	Type       string      `json:"type"`
	Parameters []Parameter `json:"parameters"`
	ReturnType string      `json:"returnType,omitempty"`
	Imports    []Import    `json:"imports"`
	Base64     string      `json:"base64,omitempty"`
	Tag        string      `json:"tag,omitempty"`
}

// Report represents the complete analysis report
type Report struct {
	Transactions  map[string]AnalysisResult `json:"transactions"`
	Scripts       map[string]AnalysisResult `json:"scripts"`
	Structs       map[string]Struct         `json:"structs"`
	Addresses     map[string]interface{}    `json:"addresses,omitempty"`
	IncludeBase64 bool                      `json:"-"`
}

// Analyzer is responsible for analyzing Cadence files
type Analyzer struct {
	Transactions  map[string]AnalysisResult
	Scripts       map[string]AnalysisResult
	Structs       map[string]Struct
	IncludeBase64 bool
	AddressesPath string // New field for storing addresses.json path
}

// New creates a new Analyzer instance
func New() *Analyzer {
	return &Analyzer{
		Transactions:  make(map[string]AnalysisResult),
		Scripts:       make(map[string]AnalysisResult),
		Structs:       make(map[string]Struct),
		IncludeBase64: false,
	}
}

// GetReport returns the current analysis report
func (a *Analyzer) GetReport() *Report {
	var addresses map[string]interface{}
	if a.AddressesPath != "" {
		if data, err := os.ReadFile(a.AddressesPath); err == nil {
			_ = json.Unmarshal(data, &addresses)
		}
	} else {
		cwd, _ := os.Getwd()
		if path, err := findAddressesJSONPath(cwd); err == nil {
			if data, err := os.ReadFile(path); err == nil {
				_ = json.Unmarshal(data, &addresses)
			}
		}
	}
	return &Report{
		Transactions:  a.Transactions,
		Scripts:       a.Scripts,
		Structs:       a.Structs,
		Addresses:     addresses,
		IncludeBase64: a.IncludeBase64,
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

	// Get full directory path as tag
	var tag string
	dir := filepath.Dir(filePath)
	if dir != "." {
		// Split the path and remove any empty parts
		parts := strings.Split(dir, string(filepath.Separator))
		var validParts []string
		for _, part := range parts {
			if part != "" && part != "." {
				validParts = append(validParts, part)
			}
		}

		// Convert to camelCase
		for i, part := range validParts {
			// Split by underscores or hyphens
			subParts := strings.FieldsFunc(part, func(r rune) bool {
				return r == '_' || r == '-'
			})
			// Capitalize each part
			for j, subPart := range subParts {
				if i == 0 && j == 0 {
					// First word starts with lowercase
					subParts[j] = strings.ToLower(subPart)
				} else {
					subParts[j] = strings.Title(strings.ToLower(subPart))
				}
			}
			validParts[i] = strings.Join(subParts, "")
		}

		// Join all parts
		tag = strings.Join(validParts, "")
		if tag != "" {
			// Ensure first character is uppercase for Swift enum
			tag = strings.Title(tag)
		}
	}

	memoryGauge := &SimpleMemoryGauge{}
	program, err := parser.ParseProgram(memoryGauge, codeWithoutImports, parser.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Create base result with common fields
	result := &AnalysisResult{
		FileName: fileName,
		Imports:  imports,
	}

	// Only set tag if it's not empty
	if tag != "" {
		result.Tag = tag
	}

	// Add base64 content if enabled
	if a.IncludeBase64 {
		result.Base64 = base64.StdEncoding.EncodeToString(content)
	}

	// Check for struct declarations
	for _, declaration := range program.Declarations() {
		if structDecl, ok := declaration.(*ast.CompositeDeclaration); ok {
			// Check if it's a struct by checking the composite kind
			if structDecl.CompositeKind == common.CompositeKindStructure {
				fields := make([]Field, 0)

				for _, member := range structDecl.Members.Declarations() {
					if field, ok := member.(*ast.FieldDeclaration); ok {
						var optional bool
						if _, isOptional := field.TypeAnnotation.Type.(*ast.OptionalType); isOptional {
							optional = true
						}

						fields = append(fields, Field{
							Name:     field.Identifier.String(),
							TypeStr:  field.TypeAnnotation.String(),
							Optional: optional,
							Access:   field.Access.String(),
						})
					}
				}

				structName := structDecl.Identifier.String()
				a.Structs[structName] = Struct{
					Name:     structName,
					Fields:   fields,
					Access:   structDecl.Access.String(),
					FileName: fileName,
				}
			}
		}
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
			result.Type = "transaction"
			result.Parameters = params
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
				result.Type = "script"
				result.Parameters = params
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
				result.Type = "script"
				result.Parameters = params
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
	if path, err := findAddressesJSONRecursive(dirPath); err == nil {
		a.AddressesPath = path
	}
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

// SetIncludeBase64 sets whether to include base64-encoded content in the analysis results
func (a *Analyzer) SetIncludeBase64(include bool) {
	a.IncludeBase64 = include
}

// 2. Helper function: search upwards for addresses.json
func findAddressesJSONPath(startPath string) (string, error) {
	dir := startPath
	for {
		candidate := filepath.Join(dir, "addresses.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("addresses.json not found")
}

// New recursive search function
func findAddressesJSONRecursive(root string) (string, error) {
	var found string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Base(path) == "addresses.json" {
			found = path
			return fmt.Errorf("found") // Use error to stop walk early
		}
		return nil
	})
	if found != "" {
		return found, nil
	}
	if err != nil && err.Error() == "found" {
		return found, nil
	}
	return "", fmt.Errorf("addresses.json not found")
}
