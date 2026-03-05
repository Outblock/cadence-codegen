package python

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/outblock/cadence-codegen/internal/analyzer"
)

// Generator handles Python code generation
type Generator struct {
	Report  analyzer.Report
	Files   map[string]string
	BaseDir string
}

// New creates a new Python code generator
func New(report analyzer.Report) *Generator {
	return &Generator{
		Report:  report,
		Files:   make(map[string]string),
		BaseDir: "",
	}
}

// SetBaseDir sets the base directory for reading files
func (g *Generator) SetBaseDir(dir string) {
	g.BaseDir = dir
}

// typeMapping maps Cadence types to Python types
var typeMapping = map[string]string{
	"String":      "str",
	"Character":   "str",
	"Int":         "int",
	"UInt":        "int",
	"UInt8":       "int",
	"UInt16":      "int",
	"UInt32":      "int",
	"UInt64":      "int",
	"UInt128":     "int",
	"UInt256":     "int",
	"Int8":        "int",
	"Int16":       "int",
	"Int32":       "int",
	"Int64":       "int",
	"Int128":      "int",
	"Int256":      "int",
	"Bool":        "bool",
	"Address":     "str",
	"UFix64":      "Decimal",
	"Fix64":       "Decimal",
	"AnyStruct":   "Any",
	"AnyResource": "Any",
	"Type":        "str",
	"StoragePath": "str",
	"PublicPath":  "str",
	"PrivatePath": "str",
	"Path":        "str",
	"Void":        "None",
}

// PyFunction represents a function in the generated code
type PyFunction struct {
	Name       string
	ClassName  string
	Parameters []PyParameter
	ReturnType string
	Base64     string
	Type       string // "transaction" or "script"
}

// PyParameter represents a parameter in Python
type PyParameter struct {
	Name     string
	Type     string
	Optional bool
}

// PyStruct represents a dataclass in Python
type PyStruct struct {
	Name           string
	RequiredFields []PyField
	OptionalFields []PyField
}

// PyField represents a field in a Python dataclass
type PyField struct {
	Name     string
	Type     string
	Optional bool
}

const structTemplate = `@dataclass
class {{.Name}}:
{{- range .RequiredFields}}
    {{.Name}}: {{.Type}}
{{- end}}
{{- range .OptionalFields}}
    {{.Name}}: Optional[{{.Type}}] = None
{{- end}}
`

const functionTemplate = `
def {{.Name}}_code() -> str:
    """Returns the Cadence code for the {{.Name}} {{.Type}}."""
    encoded = "{{.Base64}}"
    return base64.b64decode(encoded).decode("utf-8")

{{- if .Parameters}}


@dataclass
class {{.ClassName}}Params:
{{- range .Parameters}}
{{- if .Optional}}
    {{.Name}}: Optional[{{.Type}}] = None
{{- else}}
    {{.Name}}: {{.Type}}
{{- end}}
{{- end}}
{{- end}}
`

// formatFunctionName formats the filename into a valid Python function name (snake_case)
func formatFunctionName(filename string) string {
	// Remove .cdc extension
	name := strings.TrimSuffix(filename, ".cdc")
	// Split by underscores or hyphens
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-'
	})

	// Lowercase each part and join with underscores
	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
	}

	return strings.Join(parts, "_")
}

// formatClassName formats the filename into a valid Python class name (PascalCase)
func formatClassName(filename string) string {
	// Remove .cdc extension
	name := strings.TrimSuffix(filename, ".cdc")
	// Split by underscores or hyphens
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-'
	})

	// Capitalize each part (PascalCase)
	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return strings.Join(parts, "")
}

// formatFieldName converts a Cadence field name to Python snake_case
func formatFieldName(name string) string {
	// Cadence field names are typically already snake_case or camelCase
	// For simplicity, just return the name as-is (Cadence convention is camelCase/snake_case)
	return name
}

// flattenStructName removes dots from nested struct names
func flattenStructName(name string) string {
	if strings.Contains(name, ".") {
		result := ""
		for _, part := range strings.Split(name, ".") {
			result += part
		}
		return result
	}
	return name
}

// convertCadenceTypeToPython converts a Cadence type to its Python equivalent
func convertCadenceTypeToPython(cadenceType string) string {
	cadenceType = strings.TrimSpace(cadenceType)

	// Strip reference markers (&)
	if strings.HasPrefix(cadenceType, "&") {
		return convertCadenceTypeToPython(strings.TrimPrefix(cadenceType, "&"))
	}

	// Check if it's an optional type
	if strings.HasSuffix(cadenceType, "?") {
		baseType := strings.TrimSuffix(cadenceType, "?")
		pyType := convertCadenceTypeToPython(baseType)
		return fmt.Sprintf("Optional[%s]", pyType)
	}

	// Handle generic types like Capability<...> as Any
	if strings.Contains(cadenceType, "<") {
		return "Any"
	}

	// Check if it's an array type
	if strings.HasPrefix(cadenceType, "[") && strings.HasSuffix(cadenceType, "]") {
		elementType := strings.TrimPrefix(strings.TrimSuffix(cadenceType, "]"), "[")
		elementType = strings.TrimSpace(elementType)
		pyElementType := convertCadenceTypeToPython(elementType)
		return fmt.Sprintf("list[%s]", pyElementType)
	}

	// Check if it's a dictionary or intersection type
	if strings.HasPrefix(cadenceType, "{") && strings.HasSuffix(cadenceType, "}") {
		inner := strings.TrimPrefix(strings.TrimSuffix(cadenceType, "}"), "{")
		parts := strings.SplitN(inner, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			keyType := strings.TrimSpace(parts[0])
			valueType := strings.TrimSpace(parts[1])
			pyKeyType := convertCadenceTypeToPython(keyType)
			pyValueType := convertCadenceTypeToPython(valueType)
			return fmt.Sprintf("dict[%s, %s]", pyKeyType, pyValueType)
		} else if len(parts) >= 1 {
			singleType := strings.TrimSpace(parts[0])
			return convertCadenceTypeToPython(singleType)
		}
	}

	// Use the type mapping
	pyType, ok := typeMapping[cadenceType]
	if !ok {
		if strings.Contains(cadenceType, ".") {
			return flattenStructName(cadenceType)
		}
		return cadenceType
	}
	return pyType
}

// needsDecimal checks if any type in the report requires Decimal
func (g *Generator) needsDecimal() bool {
	for _, s := range g.Report.Structs {
		for _, f := range s.Fields {
			if pyType := convertCadenceTypeToPython(f.TypeStr); strings.Contains(pyType, "Decimal") {
				return true
			}
		}
	}
	for _, r := range g.Report.Transactions {
		for _, p := range r.Parameters {
			if pyType := convertCadenceTypeToPython(p.TypeStr); strings.Contains(pyType, "Decimal") {
				return true
			}
		}
	}
	for _, r := range g.Report.Scripts {
		for _, p := range r.Parameters {
			if pyType := convertCadenceTypeToPython(p.TypeStr); strings.Contains(pyType, "Decimal") {
				return true
			}
		}
		if pyType := convertCadenceTypeToPython(r.ReturnType); strings.Contains(pyType, "Decimal") {
			return true
		}
	}
	return false
}

// needsOptional checks if any type in the report uses Optional
func (g *Generator) needsOptional() bool {
	for _, s := range g.Report.Structs {
		for _, f := range s.Fields {
			if f.Optional {
				return true
			}
		}
	}
	for _, r := range g.Report.Transactions {
		for _, p := range r.Parameters {
			if p.Optional {
				return true
			}
		}
	}
	for _, r := range g.Report.Scripts {
		for _, p := range r.Parameters {
			if p.Optional {
				return true
			}
		}
	}
	return false
}

// needsAny checks if any type in the report uses Any
func (g *Generator) needsAny() bool {
	for _, s := range g.Report.Structs {
		for _, f := range s.Fields {
			if pyType := convertCadenceTypeToPython(f.TypeStr); pyType == "Any" || strings.Contains(pyType, "Any") {
				return true
			}
		}
	}
	for _, r := range g.Report.Transactions {
		for _, p := range r.Parameters {
			if pyType := convertCadenceTypeToPython(p.TypeStr); pyType == "Any" || strings.Contains(pyType, "Any") {
				return true
			}
		}
	}
	for _, r := range g.Report.Scripts {
		for _, p := range r.Parameters {
			if pyType := convertCadenceTypeToPython(p.TypeStr); pyType == "Any" || strings.Contains(pyType, "Any") {
				return true
			}
		}
		if pyType := convertCadenceTypeToPython(r.ReturnType); pyType == "Any" || strings.Contains(pyType, "Any") {
			return true
		}
	}
	return false
}

// decodeBase64ToUTF8 decodes base64 string to UTF-8
func decodeBase64ToUTF8(base64Str string) string {
	if base64Str == "" {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(decoded))
}

// Generate generates Python code for all transactions and scripts
func (g *Generator) Generate() (string, error) {
	var buffer bytes.Buffer

	// Header
	buffer.WriteString("# Code generated by cadence-codegen. DO NOT EDIT.\n\n")
	buffer.WriteString("from __future__ import annotations\n\n")
	buffer.WriteString("import base64\n")
	buffer.WriteString("from dataclasses import dataclass\n")

	if g.needsDecimal() {
		buffer.WriteString("from decimal import Decimal\n")
	}

	// Build typing imports
	var typingImports []string
	if g.needsAny() {
		typingImports = append(typingImports, "Any")
	}
	if g.needsOptional() {
		typingImports = append(typingImports, "Optional")
	}
	sort.Strings(typingImports)
	if len(typingImports) > 0 {
		buffer.WriteString(fmt.Sprintf("from typing import %s\n", strings.Join(typingImports, ", ")))
	}

	// Generate structs
	structTmpl, err := template.New("struct").Parse(structTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse struct template: %w", err)
	}

	var structNames []string
	for name := range g.Report.Structs {
		structNames = append(structNames, name)
	}
	sort.Strings(structNames)

	if len(structNames) > 0 {
		buffer.WriteString("\n\n# --- Structs ---\n")
	}

	for _, name := range structNames {
		composite := g.Report.Structs[name]
		pyStruct := PyStruct{
			Name:           flattenStructName(name),
			RequiredFields: make([]PyField, 0),
			OptionalFields: make([]PyField, 0),
		}
		for _, field := range composite.Fields {
			pyType := convertCadenceTypeToPython(field.TypeStr)
			f := PyField{
				Name:     formatFieldName(field.Name),
				Type:     pyType,
				Optional: field.Optional,
			}
			if field.Optional {
				pyStruct.OptionalFields = append(pyStruct.OptionalFields, f)
			} else {
				pyStruct.RequiredFields = append(pyStruct.RequiredFields, f)
			}
		}

		buffer.WriteString("\n")
		err = structTmpl.Execute(&buffer, pyStruct)
		if err != nil {
			return "", fmt.Errorf("failed to execute struct template: %w", err)
		}
	}

	// Collect functions
	var functions []PyFunction
	taggedFunctions := make(map[string][]PyFunction)

	// Transactions
	var txFilenames []string
	for filename := range g.Report.Transactions {
		txFilenames = append(txFilenames, filename)
	}
	sort.Strings(txFilenames)

	for _, filename := range txFilenames {
		result := g.Report.Transactions[filename]
		pyFunc := PyFunction{
			Name:       formatFunctionName(filename),
			ClassName:  formatClassName(filename),
			Parameters: make([]PyParameter, 0),
			Base64:     result.Base64,
			Type:       "transaction",
		}
		for _, param := range result.Parameters {
			pyType := convertCadenceTypeToPython(param.TypeStr)
			pyFunc.Parameters = append(pyFunc.Parameters, PyParameter{
				Name:     formatFieldName(param.Name),
				Type:     pyType,
				Optional: param.Optional,
			})
		}
		if result.Tag != "" {
			taggedFunctions[result.Tag] = append(taggedFunctions[result.Tag], pyFunc)
		} else {
			functions = append(functions, pyFunc)
		}
	}

	// Scripts
	var scriptFilenames []string
	for filename := range g.Report.Scripts {
		scriptFilenames = append(scriptFilenames, filename)
	}
	sort.Strings(scriptFilenames)

	for _, filename := range scriptFilenames {
		result := g.Report.Scripts[filename]
		pyFunc := PyFunction{
			Name:       formatFunctionName(filename),
			ClassName:  formatClassName(filename),
			Parameters: make([]PyParameter, 0),
			Base64:     result.Base64,
			Type:       "script",
		}
		if result.ReturnType != "" {
			pyFunc.ReturnType = convertCadenceTypeToPython(result.ReturnType)
		}
		for _, param := range result.Parameters {
			pyType := convertCadenceTypeToPython(param.TypeStr)
			pyFunc.Parameters = append(pyFunc.Parameters, PyParameter{
				Name:     formatFieldName(param.Name),
				Type:     pyType,
				Optional: param.Optional,
			})
		}
		if result.Tag != "" {
			taggedFunctions[result.Tag] = append(taggedFunctions[result.Tag], pyFunc)
		} else {
			functions = append(functions, pyFunc)
		}
	}

	// Sort functions by name
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Name < functions[j].Name
	})

	// Generate functions
	funcTmpl, err := template.New("function").Parse(functionTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse function template: %w", err)
	}

	hasTransactionsOrScripts := len(g.Report.Transactions) > 0 || len(g.Report.Scripts) > 0

	if hasTransactionsOrScripts && len(functions) > 0 {
		// Determine section header
		hasTx := false
		hasScript := false
		for _, f := range functions {
			if f.Type == "transaction" {
				hasTx = true
			} else {
				hasScript = true
			}
		}
		_ = hasTx
		_ = hasScript
	}

	for _, f := range functions {
		buffer.WriteString("\n")
		err = funcTmpl.Execute(&buffer, f)
		if err != nil {
			return "", fmt.Errorf("failed to execute function template: %w", err)
		}
	}

	// Generate tagged functions grouped by tag
	var tagNames []string
	for tag := range taggedFunctions {
		tagNames = append(tagNames, tag)
	}
	sort.Strings(tagNames)

	for _, tag := range tagNames {
		tagFuncs := taggedFunctions[tag]
		sort.Slice(tagFuncs, func(i, j int) bool {
			return tagFuncs[i].Name < tagFuncs[j].Name
		})

		buffer.WriteString(fmt.Sprintf("\n\n# --- %s ---\n", tag))
		for _, f := range tagFuncs {
			buffer.WriteString("\n")
			err = funcTmpl.Execute(&buffer, f)
			if err != nil {
				return "", fmt.Errorf("failed to execute function template: %w", err)
			}
		}
	}

	return buffer.String(), nil
}
