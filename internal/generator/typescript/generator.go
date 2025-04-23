package typescript

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/outblock/cadence-codegen/internal/analyzer"
)

// Generator handles TypeScript code generation
type Generator struct {
	Report  analyzer.Report
	Files   map[string]string
	BaseDir string
}

// New creates a new TypeScript code generator
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

// typeMapping maps Cadence types to TypeScript types
var typeMapping = map[string]string{
	"String":    "string",
	"Int":       "number",
	"UInt":      "number",
	"UInt8":     "number",
	"UInt16":    "number",
	"UInt32":    "number",
	"UInt64":    "number",
	"UInt128":   "string",
	"UInt256":   "string",
	"Int8":      "number",
	"Int16":     "number",
	"Int32":     "number",
	"Int64":     "number",
	"Int128":    "string",
	"Int256":    "string",
	"Bool":      "boolean",
	"Address":   "string",
	"UFix64":    "number",
	"Fix64":     "number",
	"AnyStruct": "any",
}

// fclTypeMapping only includes types that need special handling in FCL
var fclTypeMapping = map[string]string{
	"UInt128":   "UInt128",
	"UInt256":   "UInt256",
	"Int128":    "Int128",
	"Int256":    "Int256",
	"AnyStruct": "Any",
}

// TypeScriptFunction represents a function in the generated code
type TypeScriptFunction struct {
	Name       string
	Parameters []TypeScriptParameter
	ReturnType string
	Base64     string
	Type       string
}

// TypeScriptParameter represents a parameter in TypeScript
type TypeScriptParameter struct {
	Name     string
	Type     string
	Optional bool
	TypeStr  string // Original Cadence type string
}

// TypeScriptInterface represents an interface in TypeScript
type TypeScriptInterface struct {
	Name   string
	Fields []TypeScriptField
}

// TypeScriptField represents a field in a TypeScript interface
type TypeScriptField struct {
	Name     string
	Type     string
	Optional bool
}

const interfaceTemplate = `/** Generated Cadence interface */
export interface {{.Name}} {
{{- range .Fields}}
    {{.Name}}{{if .Optional}}?{{end}}: {{.Type}};
{{- end}}
}`

const functionTemplate = `/** Generated from Cadence files{{if .Tag}} in {{.Tag}} folder{{end}} */
{{- range $index, $func := .Functions}}
{{if $index}}

{{end}}export const {{$func.Name}} = async ({{range $index, $param := $func.Parameters}}{{if $index}}, {{end}}{{$param.Name}}{{if $param.Optional}}?{{end}}: {{$param.Type}}{{end}}){{if $func.ReturnType}}: Promise<{{$func.ReturnType}}>{{end}} => {
    const code = decodeCadence("{{$func.Base64}}");
    {{- if eq $func.Type "query"}}
    const result = await fcl.query({
        cadence: code,
        args: (arg, t) => [
            {{- range $func.Parameters}}
            arg({{.Name}}, {{getFCLType .TypeStr}}),
            {{- end}}
        ],
    });
    return result;
    {{- else}}
    const txId = await fcl.mutate({
        cadence: code,
        args: (arg, t) => [
            {{- range $func.Parameters}}
            arg({{.Name}}, {{getFCLType .TypeStr}}),
            {{- end}}
        ],
    });
    return txId;
    {{- end}}
};{{- end}}`

// formatFunctionName formats the filename into a valid TypeScript function name
func formatFunctionName(filename string) string {
	// Remove .cdc extension
	name := strings.TrimSuffix(filename, ".cdc")
	// Split by underscores or hyphens
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-'
	})

	// First convert all parts to lowercase
	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
	}

	// Then capitalize each part except the first one
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	// Join back together
	return strings.Join(parts, "")
}

// getFCLType gets the FCL type annotation for a Cadence type
func getFCLType(cadenceType string) string {
	// Check if it's an array type
	if strings.HasPrefix(cadenceType, "[") && strings.HasSuffix(cadenceType, "]") {
		// Extract element type
		elementType := strings.TrimPrefix(strings.TrimSuffix(cadenceType, "]"), "[")
		elementType = strings.TrimSpace(elementType)
		return fmt.Sprintf("t.Array(%s)", getFCLType(elementType))
	}

	// Check if it's a dictionary type
	if strings.HasPrefix(cadenceType, "{") && strings.HasSuffix(cadenceType, "}") {
		// Extract key and value types
		inner := strings.TrimPrefix(strings.TrimSuffix(cadenceType, "}"), "{")
		parts := strings.Split(inner, ":")
		if len(parts) == 2 {
			keyType := strings.TrimSpace(parts[0])
			valueType := strings.TrimSpace(parts[1])
			return fmt.Sprintf("t.Dictionary({ key: %s, value: %s })", getFCLType(keyType), getFCLType(valueType))
		}
	}

	// For special cases, use the FCL type mapping
	if fclType, ok := fclTypeMapping[cadenceType]; ok {
		return fmt.Sprintf("t.%s", fclType)
	}

	// For all other types, use the type name with t. prefix
	return fmt.Sprintf("t.%s", cadenceType)
}

// convertCadenceTypeToTypeScript converts a Cadence type to its TypeScript equivalent
func convertCadenceTypeToTypeScript(cadenceType string) string {
	// Check if it's an optional type
	if strings.HasSuffix(cadenceType, "?") {
		baseType := strings.TrimSuffix(cadenceType, "?")
		tsType := convertCadenceTypeToTypeScript(baseType)
		return fmt.Sprintf("%s | undefined", tsType)
	}

	// Check if it's an array type
	if strings.HasPrefix(cadenceType, "[") && strings.HasSuffix(cadenceType, "]") {
		// Extract element type
		elementType := strings.TrimPrefix(strings.TrimSuffix(cadenceType, "]"), "[")
		elementType = strings.TrimSpace(elementType)

		// Convert element type using type mapping
		tsElementType, ok := typeMapping[elementType]
		if !ok {
			tsElementType = elementType
		}

		return fmt.Sprintf("%s[]", tsElementType)
	}

	// Check if it's a dictionary type
	if strings.HasPrefix(cadenceType, "{") && strings.HasSuffix(cadenceType, "}") {
		// Extract key and value types
		inner := strings.TrimPrefix(strings.TrimSuffix(cadenceType, "}"), "{")
		parts := strings.Split(inner, ":")
		if len(parts) == 2 {
			keyType := strings.TrimSpace(parts[0])
			valueType := strings.TrimSpace(parts[1])

			// Convert key and value types
			tsKeyType, ok := typeMapping[keyType]
			if !ok {
				tsKeyType = keyType
			}
			tsValueType, ok := typeMapping[valueType]
			if !ok {
				tsValueType = valueType
			}

			return fmt.Sprintf("Record<%s, %s>", tsKeyType, tsValueType)
		}
	}

	// For non-dictionary types, use the type mapping
	tsType, ok := typeMapping[cadenceType]
	if !ok {
		return cadenceType
	}
	return tsType
}

// Generate generates TypeScript code for all transactions and scripts
func (g *Generator) Generate() (string, error) {
	var buffer bytes.Buffer
	var functions []TypeScriptFunction
	var interfaces []TypeScriptInterface

	// Map to store functions by tag
	taggedFunctions := make(map[string][]TypeScriptFunction)

	// Add header with imports
	buffer.WriteString("import * as fcl from \"@onflow/fcl\";\n\n")
	buffer.WriteString("/** Utility function to decode Base64 Cadence code */\n")
	buffer.WriteString("const decodeCadence = (code: string): string => Buffer.from(code, 'base64').toString('utf8');\n\n")
	buffer.WriteString("/** Generated from Cadence files */\n")

	// Generate interfaces from composite types
	for name, composite := range g.Report.Structs {
		tsInterface := TypeScriptInterface{
			Name:   name,
			Fields: make([]TypeScriptField, 0),
		}

		for _, field := range composite.Fields {
			tsType := convertCadenceTypeToTypeScript(field.TypeStr)

			tsInterface.Fields = append(tsInterface.Fields, TypeScriptField{
				Name:     field.Name,
				Type:     tsType,
				Optional: field.Optional,
			})
		}

		interfaces = append(interfaces, tsInterface)
	}

	// Generate interface code
	interfaceTmpl, err := template.New("interface").Parse(interfaceTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse interface template: %w", err)
	}

	for _, i := range interfaces {
		err = interfaceTmpl.Execute(&buffer, i)
		if err != nil {
			return "", fmt.Errorf("failed to execute interface template: %w", err)
		}
		buffer.WriteString("\n")
	}

	// Generate functions for transactions
	for filename, result := range g.Report.Transactions {
		tsFunction := TypeScriptFunction{
			Name:       formatFunctionName(filename),
			Parameters: make([]TypeScriptParameter, 0),
			Base64:     result.Base64,
			Type:       "transaction",
		}

		for _, param := range result.Parameters {
			tsType := convertCadenceTypeToTypeScript(param.TypeStr)

			tsFunction.Parameters = append(tsFunction.Parameters, TypeScriptParameter{
				Name:     param.Name,
				Type:     tsType,
				Optional: param.Optional,
				TypeStr:  param.TypeStr,
			})
		}

		if result.Tag != "" {
			taggedFunctions[result.Tag] = append(taggedFunctions[result.Tag], tsFunction)
		} else {
			functions = append(functions, tsFunction)
		}
	}

	// Generate functions for scripts
	for filename, result := range g.Report.Scripts {
		tsFunction := TypeScriptFunction{
			Name:       formatFunctionName(filename),
			Parameters: make([]TypeScriptParameter, 0),
			Base64:     result.Base64,
			Type:       "query",
		}

		if result.ReturnType != "" {
			tsType := convertCadenceTypeToTypeScript(result.ReturnType)
			tsFunction.ReturnType = tsType
		}

		for _, param := range result.Parameters {
			tsType := convertCadenceTypeToTypeScript(param.TypeStr)

			tsFunction.Parameters = append(tsFunction.Parameters, TypeScriptParameter{
				Name:     param.Name,
				Type:     tsType,
				Optional: param.Optional,
				TypeStr:  param.TypeStr,
			})
		}

		if result.Tag != "" {
			taggedFunctions[result.Tag] = append(taggedFunctions[result.Tag], tsFunction)
		} else {
			functions = append(functions, tsFunction)
		}
	}

	// Generate functions
	funcMap := template.FuncMap{
		"getFCLType": getFCLType,
	}
	tmpl, err := template.New("function").Funcs(funcMap).Parse(functionTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// First generate the base functions
	err = tmpl.Execute(&buffer, struct {
		Functions []TypeScriptFunction
		Tag       string
	}{
		Functions: functions,
		Tag:       "",
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Then generate tagged functions in separate sections
	for tag, tagFunctions := range taggedFunctions {
		buffer.WriteString("\n")
		err = tmpl.Execute(&buffer, struct {
			Functions []TypeScriptFunction
			Tag       string
		}{
			Functions: tagFunctions,
			Tag:       tag,
		})
		if err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}
	}

	return buffer.String(), nil
}
