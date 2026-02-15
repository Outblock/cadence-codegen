package golang

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/outblock/cadence-codegen/internal/analyzer"
)

// Generator handles Go code generation
type Generator struct {
	Report  analyzer.Report
	Files   map[string]string
	BaseDir string
}

// New creates a new Go code generator
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

// typeMapping maps Cadence types to Go types
var typeMapping = map[string]string{
	"String":      "string",
	"Character":   "string",
	"Int":         "int",
	"UInt":        "uint",
	"UInt8":       "uint8",
	"UInt16":      "uint16",
	"UInt32":      "uint32",
	"UInt64":      "uint64",
	"UInt128":     "*big.Int",
	"UInt256":     "*big.Int",
	"Int8":        "int8",
	"Int16":       "int16",
	"Int32":       "int32",
	"Int64":       "int64",
	"Int128":      "*big.Int",
	"Int256":      "*big.Int",
	"Bool":        "bool",
	"Address":     "string",
	"UFix64":      "string",
	"Fix64":       "string",
	"AnyStruct":   "interface{}",
	"AnyResource": "interface{}",
	"Type":        "string",
	"StoragePath": "string",
	"PublicPath":  "string",
	"PrivatePath": "string",
	"Path":        "string",
	"Void":        "",
}

// GoFunction represents a function in the generated code
type GoFunction struct {
	Name       string
	Parameters []GoParameter
	ReturnType string
	Base64     string
	Type       string // "transaction" or "query"
}

// GoParameter represents a parameter in Go
type GoParameter struct {
	Name     string
	Type     string
	Optional bool
}

// GoStruct represents a struct in Go
type GoStruct struct {
	Name   string
	Fields []GoField
}

// GoField represents a field in a Go struct
type GoField struct {
	Name     string
	Type     string
	Optional bool
	JSONName string
}

const structTemplate = `// {{.Name}} is a generated Cadence struct.
type {{.Name}} struct {
{{- range .Fields}}
	{{.Name}} {{if .Optional}}*{{end}}{{.Type}} ` + "`" + `json:"{{.JSONName}}"` + "`" + `
{{- end}}
}
`

const functionTemplate = `// {{.Name}}Code returns the Cadence code for the {{.Name}} {{.Type}}.
func {{.Name}}Code() string {
	const encoded = "{{.Base64}}"
	code, _ := base64.StdEncoding.DecodeString(encoded)
	return string(code)
}
{{if .Parameters}}
// {{.Name}}Params holds the parameters for {{.Name}}.
type {{.Name}}Params struct {
{{- range .Parameters}}
	{{.Name}} {{if .Optional}}*{{end}}{{.Type}}
{{- end}}
}
{{end}}`

// formatFunctionName formats the filename into a valid Go exported function name (PascalCase)
func formatFunctionName(filename string) string {
	// Remove .cdc extension
	name := strings.TrimSuffix(filename, ".cdc")
	// Split by underscores or hyphens
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-'
	})

	// Capitalize each part (PascalCase for Go exports)
	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return strings.Join(parts, "")
}

// formatFieldName converts a Cadence field name to Go exported PascalCase
func formatFieldName(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-'
	})
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
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

// convertCadenceTypeToGo converts a Cadence type to its Go equivalent
func convertCadenceTypeToGo(cadenceType string) string {
	cadenceType = strings.TrimSpace(cadenceType)

	// Strip reference markers (&)
	if strings.HasPrefix(cadenceType, "&") {
		return convertCadenceTypeToGo(strings.TrimPrefix(cadenceType, "&"))
	}

	// Check if it's an optional type
	if strings.HasSuffix(cadenceType, "?") {
		baseType := strings.TrimSuffix(cadenceType, "?")
		goType := convertCadenceTypeToGo(baseType)
		// Pointer types and interfaces don't need extra *
		if strings.HasPrefix(goType, "*") || goType == "interface{}" {
			return goType
		}
		return "*" + goType
	}

	// Handle generic types like Capability<...> as interface{}
	if strings.Contains(cadenceType, "<") {
		return "interface{}"
	}

	// Check if it's an array type
	if strings.HasPrefix(cadenceType, "[") && strings.HasSuffix(cadenceType, "]") {
		elementType := strings.TrimPrefix(strings.TrimSuffix(cadenceType, "]"), "[")
		elementType = strings.TrimSpace(elementType)
		goElementType := convertCadenceTypeToGo(elementType)
		return fmt.Sprintf("[]%s", goElementType)
	}

	// Check if it's a dictionary or intersection type
	if strings.HasPrefix(cadenceType, "{") && strings.HasSuffix(cadenceType, "}") {
		inner := strings.TrimPrefix(strings.TrimSuffix(cadenceType, "}"), "{")
		parts := strings.SplitN(inner, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			keyType := strings.TrimSpace(parts[0])
			valueType := strings.TrimSpace(parts[1])
			goKeyType := convertCadenceTypeToGo(keyType)
			goValueType := convertCadenceTypeToGo(valueType)
			return fmt.Sprintf("map[%s]%s", goKeyType, goValueType)
		} else if len(parts) >= 1 {
			singleType := strings.TrimSpace(parts[0])
			return convertCadenceTypeToGo(singleType)
		}
	}

	// Use the type mapping
	goType, ok := typeMapping[cadenceType]
	if !ok {
		if strings.Contains(cadenceType, ".") {
			return flattenStructName(cadenceType)
		}
		return cadenceType
	}
	return goType
}

// needsBigInt checks if any type in the report requires math/big
func (g *Generator) needsBigInt() bool {
	for _, s := range g.Report.Structs {
		for _, f := range s.Fields {
			if goType := convertCadenceTypeToGo(f.TypeStr); strings.Contains(goType, "big.Int") {
				return true
			}
		}
	}
	for _, r := range g.Report.Transactions {
		for _, p := range r.Parameters {
			if goType := convertCadenceTypeToGo(p.TypeStr); strings.Contains(goType, "big.Int") {
				return true
			}
		}
	}
	for _, r := range g.Report.Scripts {
		for _, p := range r.Parameters {
			if goType := convertCadenceTypeToGo(p.TypeStr); strings.Contains(goType, "big.Int") {
				return true
			}
		}
		if goType := convertCadenceTypeToGo(r.ReturnType); strings.Contains(goType, "big.Int") {
			return true
		}
	}
	return false
}

// decodeBase64ToUTF8 decodes base64 string to UTF-8 for embedding in comments
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

// Generate generates Go code for all transactions and scripts
func (g *Generator) Generate() (string, error) {
	var buffer bytes.Buffer

	// Header
	buffer.WriteString("// Code generated by cadence-codegen. DO NOT EDIT.\n")
	buffer.WriteString("package cadence_generated\n\n")

	// Build imports
	imports := []string{`"encoding/base64"`}
	if g.needsBigInt() {
		imports = append(imports, `"math/big"`)
	}
	sort.Strings(imports)

	buffer.WriteString("import (\n")
	for _, imp := range imports {
		buffer.WriteString("\t" + imp + "\n")
	}
	buffer.WriteString(")\n")

	// Suppress unused import warning for base64 when there are only structs
	hasCode := len(g.Report.Transactions) > 0 || len(g.Report.Scripts) > 0
	if !hasCode {
		buffer.WriteString("\nvar _ = base64.StdEncoding\n")
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

	for _, name := range structNames {
		composite := g.Report.Structs[name]
		goStruct := GoStruct{
			Name:   flattenStructName(name),
			Fields: make([]GoField, 0),
		}
		for _, field := range composite.Fields {
			goType := convertCadenceTypeToGo(field.TypeStr)
			goStruct.Fields = append(goStruct.Fields, GoField{
				Name:     formatFieldName(field.Name),
				Type:     goType,
				Optional: field.Optional,
				JSONName: field.Name,
			})
		}

		buffer.WriteString("\n")
		err = structTmpl.Execute(&buffer, goStruct)
		if err != nil {
			return "", fmt.Errorf("failed to execute struct template: %w", err)
		}
	}

	// Collect functions
	var functions []GoFunction
	taggedFunctions := make(map[string][]GoFunction)

	// Transactions
	var txFilenames []string
	for filename := range g.Report.Transactions {
		txFilenames = append(txFilenames, filename)
	}
	sort.Strings(txFilenames)

	for _, filename := range txFilenames {
		result := g.Report.Transactions[filename]
		goFunc := GoFunction{
			Name:       formatFunctionName(filename),
			Parameters: make([]GoParameter, 0),
			Base64:     result.Base64,
			Type:       "transaction",
		}
		for _, param := range result.Parameters {
			goType := convertCadenceTypeToGo(param.TypeStr)
			goFunc.Parameters = append(goFunc.Parameters, GoParameter{
				Name:     formatFieldName(param.Name),
				Type:     goType,
				Optional: param.Optional,
			})
		}
		if result.Tag != "" {
			taggedFunctions[result.Tag] = append(taggedFunctions[result.Tag], goFunc)
		} else {
			functions = append(functions, goFunc)
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
		goFunc := GoFunction{
			Name:       formatFunctionName(filename),
			Parameters: make([]GoParameter, 0),
			Base64:     result.Base64,
			Type:       "script",
		}
		if result.ReturnType != "" {
			goFunc.ReturnType = convertCadenceTypeToGo(result.ReturnType)
		}
		for _, param := range result.Parameters {
			goType := convertCadenceTypeToGo(param.TypeStr)
			goFunc.Parameters = append(goFunc.Parameters, GoParameter{
				Name:     formatFieldName(param.Name),
				Type:     goType,
				Optional: param.Optional,
			})
		}
		if result.Tag != "" {
			taggedFunctions[result.Tag] = append(taggedFunctions[result.Tag], goFunc)
		} else {
			functions = append(functions, goFunc)
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

		buffer.WriteString(fmt.Sprintf("\n// --- %s ---\n", tag))
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
