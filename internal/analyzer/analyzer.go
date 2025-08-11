package analyzer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
// flattenStructName flattens nested struct names by removing dots
func flattenStructName(name string) string {
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		result := ""
		for _, part := range parts {
			result += part
		}
		return result
	}
	return name
}

// flattenReturnType flattens nested type references in return types
func flattenReturnType(returnType string) string {
	// Handle array types like "[FlowIDTableStaking.DelegatorInfo]?"
	if strings.HasPrefix(returnType, "[") && strings.HasSuffix(returnType, "]") {
		// Extract the inner type
		innerType := strings.TrimPrefix(strings.TrimSuffix(returnType, "]"), "[")
		// Check if it has optional marker
		hasOptional := strings.HasSuffix(innerType, "?")
		if hasOptional {
			innerType = strings.TrimSuffix(innerType, "?")
		}
		// Flatten the inner type
		flattenedInner := flattenStructName(innerType)
		// Reconstruct the type
		result := "[" + flattenedInner + "]"
		if hasOptional {
			result += "?"
		}
		return result
	}

	// Handle optional types like "FlowIDTableStaking.DelegatorInfo?"
	if strings.HasSuffix(returnType, "?") {
		baseType := strings.TrimSuffix(returnType, "?")
		flattenedBase := flattenStructName(baseType)
		return flattenedBase + "?"
	}

	// Handle simple types
	return flattenStructName(returnType)
}

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

	// Flatten struct names in the report
	flattenedStructs := make(map[string]Struct)
	for key, structDef := range a.Structs {
		flattenedKey := flattenStructName(key)
		flattenedStruct := structDef
		flattenedStruct.Name = flattenStructName(structDef.Name)
		flattenedStructs[flattenedKey] = flattenedStruct
	}

	return &Report{
		Transactions:  a.Transactions,
		Scripts:       a.Scripts,
		Structs:       flattenedStructs,
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

// FetchContractFromChain fetches contract code from Flow blockchain and analyzes its structures
func (a *Analyzer) FetchContractFromChain(contractName string, network string) error {
	// Get contract address from addresses - we need to get the report first
	report := a.GetReport()
	if report.Addresses == nil {
		return fmt.Errorf("no addresses available")
	}

	networkAddresses, ok := report.Addresses[network].(map[string]interface{})
	if !ok {
		return fmt.Errorf("network %s not found in addresses", network)
	}

	// Try to find the contract with different possible prefixes
	var contractAddress string
	var found bool

	// Try with 0x prefix first
	prefixedName := "0x" + contractName
	if contractAddress, found = networkAddresses[prefixedName].(string); found {
		contractName = prefixedName // Use the prefixed name for the API call
	} else if contractAddress, found = networkAddresses[contractName].(string); found {
		// Use original name
	} else {
		return fmt.Errorf("contract %s not found in network %s", contractName, network)
	}

	// Remove 0x prefix if present
	contractAddress = strings.TrimPrefix(contractAddress, "0x")

	// Determine API endpoint based on network
	var apiEndpoint string
	switch network {
	case "mainnet":
		apiEndpoint = "https://rest-mainnet.onflow.org"
	case "testnet":
		apiEndpoint = "https://rest-testnet.onflow.org"
	default:
		return fmt.Errorf("unsupported network: %s", network)
	}

	// Fetch contract from chain
	url := fmt.Sprintf("%s/v1/accounts/%s?expand=contracts", apiEndpoint, contractAddress)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch contract: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var accountData map[string]interface{}
	if err := json.Unmarshal(body, &accountData); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	contracts, ok := accountData["contracts"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("no contracts found in response")
	}

	// Try to find the contract with the original name (without 0x prefix)
	originalContractName := strings.TrimPrefix(contractName, "0x")
	contractCode, ok := contracts[originalContractName].(string)
	if !ok {
		return fmt.Errorf("contract %s not found in response", originalContractName)
	}

	// Decode base64 contract code
	decodedCode, err := base64.StdEncoding.DecodeString(contractCode)
	if err != nil {
		return fmt.Errorf("failed to decode contract code: %w", err)
	}

	// Analyze the decoded contract code
	return a.analyzeContractCode(string(decodedCode), originalContractName)
}

// analyzeContractCode analyzes Cadence contract code and extracts structure definitions
func (a *Analyzer) analyzeContractCode(code string, contractName string) error {
	// This is a simplified version - you might want to use a proper Cadence parser
	// For now, we'll use regex to find struct definitions

	lines := strings.Split(code, "\n")
	var currentStructName string
	var inStruct bool
	var braceCount int

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for struct definition start
		if strings.HasPrefix(line, "access(all) struct") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				structName := parts[2]
				// Create a composite name with contract prefix
				fullStructName := fmt.Sprintf("%s.%s", contractName, structName)

				// Check if we already have this struct
				if _, exists := a.Structs[fullStructName]; !exists {
					// Create new struct
					newStruct := Struct{
						Name:   fullStructName,
						Fields: []Field{},
					}
					a.Structs[fullStructName] = newStruct
					currentStructName = fullStructName
					inStruct = true
					braceCount = 0
				}
			}
		}

		// If we're inside a struct, look for field definitions
		if inStruct && currentStructName != "" {
			// Count braces to track struct boundaries
			braceCount += strings.Count(line, "{")
			braceCount -= strings.Count(line, "}")

			// Check for field definitions (simplified parsing)
			if strings.Contains(line, "let") && strings.Contains(line, ":") {
				// Extract field name and type
				fieldParts := strings.Fields(line)
				for j, part := range fieldParts {
					if part == "let" && j+2 < len(fieldParts) {
						fieldName := strings.TrimSuffix(fieldParts[j+1], ":")
						fieldType := strings.TrimSuffix(fieldParts[j+2], ";")

						// Create field
						field := Field{
							Name:     fieldName,
							TypeStr:  fieldType,
							Optional: strings.Contains(fieldType, "?"),
						}

						// Get the struct and add the field
						if structDef, exists := a.Structs[currentStructName]; exists {
							structDef.Fields = append(structDef.Fields, field)
							a.Structs[currentStructName] = structDef
						}
						break
					}
				}
			}

			// Check if we've reached the end of the struct
			if braceCount <= 0 {
				inStruct = false
				currentStructName = ""
			}
		}
	}

	return nil
}

// ResolveNestedTypes resolves nested type references by fetching contracts from chain
func (a *Analyzer) ResolveNestedTypes(network string) error {
	// Collect all nested type references that are actually used
	nestedTypes := make(map[string]map[string]bool) // contract -> set of struct names

	// Helper function to extract nested types from a type string
	extractNestedTypes := func(typeStr string) {
		if strings.Contains(typeStr, ".") {
			// Remove array brackets and optional markers first
			cleanType := strings.TrimSuffix(strings.TrimSuffix(typeStr, "?"), "]")
			cleanType = strings.TrimPrefix(cleanType, "[")
			if strings.Contains(cleanType, ".") {
				parts := strings.Split(cleanType, ".")
				if len(parts) == 2 {
					contractName := parts[0]
					structName := parts[1]
					if nestedTypes[contractName] == nil {
						nestedTypes[contractName] = make(map[string]bool)
					}
					nestedTypes[contractName][structName] = true
				}
			}
		}
	}

	// Check in scripts for nested references
	for _, script := range a.Scripts {
		extractNestedTypes(script.ReturnType)
	}

	// Check in transactions for nested references
	for _, transaction := range a.Transactions {
		extractNestedTypes(transaction.ReturnType)
	}

	// Check in structs for nested references
	for _, structDef := range a.Structs {
		for _, field := range structDef.Fields {
			extractNestedTypes(field.TypeStr)
		}
	}

	fmt.Printf("Found nested types to resolve: %v\n", nestedTypes)

	// Fetch each contract and analyze only the used structures
	for contractName, structNames := range nestedTypes {
		if err := a.FetchContractFromChainSelective(contractName, network, structNames); err != nil {
			fmt.Printf("Warning: failed to fetch contract %s: %v\n", contractName, err)
			// Continue with other contracts even if one fails
		}
	}

	return nil
}

// FetchContractFromChainSelective fetches contract code and analyzes only specific structures
func (a *Analyzer) FetchContractFromChainSelective(contractName string, network string, structNames map[string]bool) error {
	// Get contract address from addresses - we need to get the report first
	report := a.GetReport()
	if report.Addresses == nil {
		return fmt.Errorf("no addresses available")
	}

	networkAddresses, ok := report.Addresses[network].(map[string]interface{})
	if !ok {
		return fmt.Errorf("network %s not found in addresses", network)
	}

	// Try to find the contract with different possible prefixes
	var contractAddress string
	var found bool

	// Try with 0x prefix first
	prefixedName := "0x" + contractName
	if contractAddress, found = networkAddresses[prefixedName].(string); found {
		contractName = prefixedName // Use the prefixed name for the API call
	} else if contractAddress, found = networkAddresses[contractName].(string); found {
		// Use original name
	} else {
		return fmt.Errorf("contract %s not found in network %s", contractName, network)
	}

	// Remove 0x prefix if present
	contractAddress = strings.TrimPrefix(contractAddress, "0x")

	// Determine API endpoint based on network
	var apiEndpoint string
	switch network {
	case "mainnet":
		apiEndpoint = "https://rest-mainnet.onflow.org"
	case "testnet":
		apiEndpoint = "https://rest-testnet.onflow.org"
	default:
		return fmt.Errorf("unsupported network: %s", network)
	}

	// Fetch contract from chain
	url := fmt.Sprintf("%s/v1/accounts/%s?expand=contracts", apiEndpoint, contractAddress)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch contract: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var accountData map[string]interface{}
	if err := json.Unmarshal(body, &accountData); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	contracts, ok := accountData["contracts"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("no contracts found in response")
	}

	// Try to find the contract with the original name (without 0x prefix)
	originalContractName := strings.TrimPrefix(contractName, "0x")
	contractCode, ok := contracts[originalContractName].(string)
	if !ok {
		return fmt.Errorf("contract %s not found in response", originalContractName)
	}

	// Decode base64 contract code
	decodedCode, err := base64.StdEncoding.DecodeString(contractCode)
	if err != nil {
		return fmt.Errorf("failed to decode contract code: %w", err)
	}

	// Analyze the decoded contract code, but only for the specified structures
	return a.analyzeContractCodeSelective(string(decodedCode), originalContractName, structNames)
}

// analyzeContractCodeSelective analyzes Cadence contract code and extracts only specified structure definitions
func (a *Analyzer) analyzeContractCodeSelective(code string, contractName string, structNames map[string]bool) error {
	// This is a simplified version - you might want to use a proper Cadence parser
	// For now, we'll use regex to find struct definitions

	lines := strings.Split(code, "\n")
	var currentStructName string
	var inStruct bool
	var braceCount int

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for struct definition start
		if strings.HasPrefix(line, "access(all) struct") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				structName := parts[2]
				// Only process if this struct is in our list of needed structs
				if structNames[structName] {
					// Create a composite name with contract prefix
					fullStructName := fmt.Sprintf("%s.%s", contractName, structName)

					// Check if we already have this struct
					if _, exists := a.Structs[fullStructName]; !exists {
						// Create new struct
						newStruct := Struct{
							Name:   fullStructName,
							Fields: []Field{},
						}
						a.Structs[fullStructName] = newStruct
						currentStructName = fullStructName
						inStruct = true
						braceCount = 0
					}
				}
			}
		}

		// If we're inside a struct, look for field definitions
		if inStruct && currentStructName != "" {
			// Count braces to track struct boundaries
			braceCount += strings.Count(line, "{")
			braceCount -= strings.Count(line, "}")

			// Check for field definitions (simplified parsing)
			if strings.Contains(line, "let") && strings.Contains(line, ":") {
				// Extract field name and type
				fieldParts := strings.Fields(line)
				for j, part := range fieldParts {
					if part == "let" && j+2 < len(fieldParts) {
						fieldName := strings.TrimSuffix(fieldParts[j+1], ":")
						fieldType := strings.TrimSuffix(fieldParts[j+2], ";")

						// Create field
						field := Field{
							Name:     fieldName,
							TypeStr:  fieldType,
							Optional: strings.Contains(fieldType, "?"),
						}

						// Get the struct and add the field
						if structDef, exists := a.Structs[currentStructName]; exists {
							structDef.Fields = append(structDef.Fields, field)
							a.Structs[currentStructName] = structDef
						}
						break
					}
				}
			}

			// Check if we've reached the end of the struct
			if braceCount <= 0 {
				inStruct = false
				currentStructName = ""
			}
		}
	}

	return nil
}
