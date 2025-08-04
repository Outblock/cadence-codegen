package typescript

import (
	"bytes"
	"encoding/json"
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
	"UFix64":    "string",
	"Fix64":     "string",
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

const functionTemplate = `{{- range $index, $func := .Functions}}
{{if $index}}

{{end}}  public async {{$func.Name}}({{range $index, $param := $func.Parameters}}{{if $index}}, {{end}}{{$param.Name}}{{if $param.Optional}}?{{end}}: {{$param.Type}}{{end}}){{if $func.ReturnType}}: Promise<{{$func.ReturnType}}>{{end}} {
    const code = decodeCadence("{{$func.Base64}}");
    {{- if eq $func.Type "query"}}
    let config = {
      cadence: code,
      name: "{{$func.Name}}",
      type: "script",
      args: (arg: any, t: any) => [
        {{- range $func.Parameters}}
        arg({{.Name}}, {{getFCLType .TypeStr}}),
        {{- end}}
      ],
	  	limit: 9999,
    };
    config = await this.runRequestInterceptors(config);
    let response = await fcl.query(config);
    response = await this.runResponseInterceptors(config, response);
    return response;
    {{- else}}
    let config = {
      cadence: code,
      name: "{{$func.Name}}",
      type: "transaction",
      args: (arg: any, t: any) => [
        {{- range $func.Parameters}}
        arg({{.Name}}, {{getFCLType .TypeStr}}),
        {{- end}}
      ],
	  limit: 9999,
    };
    config = await this.runRequestInterceptors(config);
    let txId = await fcl.mutate(config);
    txId = await this.runResponseInterceptors(config, txId);
    return txId;
    {{- end}}
  }
{{- end}}`

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
	// Check if it's an optional type - strip the ? for FCL type
	if strings.HasSuffix(cadenceType, "?") {
		baseType := strings.TrimSuffix(cadenceType, "?")
		return getFCLType(baseType)
	}

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

// Add a function to flatten nested struct names for interface naming
func flattenStructName(name string) string {
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		// Handle multiple dots by joining all parts
		result := ""
		for _, part := range parts {
			result += part
		}
		return result
	}
	return name
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

		// Convert element type recursively
		tsElementType := convertCadenceTypeToTypeScript(elementType)

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

			// Convert key and value types recursively
			tsKeyType := convertCadenceTypeToTypeScript(keyType)
			tsValueType := convertCadenceTypeToTypeScript(valueType)

			return fmt.Sprintf("Record<%s, %s>", tsKeyType, tsValueType)
		}
	}

	// For non-dictionary types, use the type mapping
	tsType, ok := typeMapping[cadenceType]
	if !ok {
		// New: If it's a nested name, flatten it
		if strings.Contains(cadenceType, ".") {
			return flattenStructName(cadenceType)
		}
		return cadenceType
	}
	return tsType
}

// Generate generates TypeScript code for all transactions and scripts
func (g *Generator) Generate() (string, error) {
	var buffer bytes.Buffer
	var functions []TypeScriptFunction
	// Remove interfaces variable declaration
	// var interfaces []TypeScriptInterface

	// Map to store functions by tag
	taggedFunctions := make(map[string][]TypeScriptFunction)

	// Add header with imports
	buffer.WriteString("import * as fcl from \"@onflow/fcl\";\n")
	buffer.WriteString("import { Buffer } from 'buffer';\n\n")
	buffer.WriteString("/** Utility function to decode Base64 Cadence code */\n")
	buffer.WriteString("const decodeCadence = (code: string): string => Buffer.from(code, 'base64').toString('utf8');\n\n")
	buffer.WriteString("/** Generated from Cadence files */\n")

	// 1. Output all interfaces/types (including composite types)
	// Add FlowSigner interface
	buffer.WriteString("/** Flow Signer interface for transaction signing */\n")
	buffer.WriteString("export interface FlowSigner {\n")
	buffer.WriteString("  address: string;\n")
	buffer.WriteString("  keyIndex: number;\n")
	buffer.WriteString("  sign(signableData: Uint8Array): Promise<Uint8Array>;\n")
	buffer.WriteString("  authzFunc: (account: any) => Promise<any>;\n")
	buffer.WriteString("}\n\n")

	// Add CompositeSignature and AuthorizationAccount interfaces
	buffer.WriteString("export interface CompositeSignature {\n")
	buffer.WriteString("  addr: string;\n")
	buffer.WriteString("  keyId: number;\n")
	buffer.WriteString("  signature: string;\n")
	buffer.WriteString("}\n\n")
	buffer.WriteString("export interface AuthorizationAccount extends Record<string, any> {\n")
	buffer.WriteString("  tempId: string;\n")
	buffer.WriteString("  addr: string;\n")
	buffer.WriteString("  keyId: number;\n")
	buffer.WriteString("  signingFunction: (signable: { message: string }) => Promise<CompositeSignature>;\n")
	buffer.WriteString("}\n\n")
	buffer.WriteString("export type AuthorizationFunction = (account: any) => Promise<AuthorizationAccount>;\n\n")

	// Export addresses if available
	if g.Report.Addresses != nil {
		buffer.WriteString("/** Network addresses for contract imports */\n")
		buffer.WriteString("export const addresses = ")
		// Convert addresses to JSON string
		addressesJSON, err := json.Marshal(g.Report.Addresses)
		if err != nil {
			return "", fmt.Errorf("failed to marshal addresses: %w", err)
		}
		buffer.WriteString(string(addressesJSON))
		buffer.WriteString(";\n\n")
	}

	// Generate interfaces from composite types
	interfaceTmpl, err := template.New("interface").Parse(interfaceTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse interface template: %w", err)
	}

	// Group structs by contract name for nested types
	contractStructs := make(map[string][]analyzer.Struct)
	regularStructs := make([]analyzer.Struct, 0)

	for _, composite := range g.Report.Structs {
		if strings.Contains(composite.Name, ".") {
			// This is a nested type, group by contract
			parts := strings.Split(composite.Name, ".")
			if len(parts) == 2 {
				contractName := parts[0]
				contractStructs[contractName] = append(contractStructs[contractName], composite)
			}
		} else {
			// This is a regular struct
			regularStructs = append(regularStructs, composite)
		}
	}

	// Generate regular struct interfaces
	for _, composite := range regularStructs {
		tsInterface := TypeScriptInterface{
			Name:   flattenStructName(composite.Name),
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
		err = interfaceTmpl.Execute(&buffer, tsInterface)
		if err != nil {
			return "", fmt.Errorf("failed to execute interface template: %w", err)
		}
		buffer.WriteString("\n\n")
	}

	// Generate nested type interfaces as namespaces
	for _, structs := range contractStructs {
		for _, composite := range structs {
			tsInterface := TypeScriptInterface{
				Name:   flattenStructName(composite.Name),
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
			err = interfaceTmpl.Execute(&buffer, tsInterface)
			if err != nil {
				return "", fmt.Errorf("failed to execute interface template: %w", err)
			}
			buffer.WriteString("\n\n")
		}
	}

	// 2. Output class header and interceptor related code
	buffer.WriteString("type RequestInterceptor = (config: any) => any | Promise<any>;\n")
	buffer.WriteString("type ResponseInterceptor = (config: any, response: any) => any | Promise<any>;\n\n")
	buffer.WriteString("export class CadenceService {\n")
	buffer.WriteString("  private requestInterceptors: RequestInterceptor[] = [];\n")
	buffer.WriteString("  private responseInterceptors: ResponseInterceptor[] = [];\n\n")

	// Insert constructor
	buffer.WriteString("  constructor() {\n")
	buffer.WriteString("  }\n\n")

	buffer.WriteString("  useRequestInterceptor(interceptor: RequestInterceptor) {\n    this.requestInterceptors.push(interceptor);\n  }\n\n")
	buffer.WriteString("  useResponseInterceptor(interceptor: ResponseInterceptor) {\n    this.responseInterceptors.push(interceptor);\n  }\n\n")
	buffer.WriteString("  private async runRequestInterceptors(config: any) {\n    let c = config;\n    for (const interceptor of this.requestInterceptors) {\n      c = await interceptor(c);\n    }\n    return c;\n  }\n\n")
	buffer.WriteString("  private async runResponseInterceptors(config: any, response: any) {\n    let r = response;\n    for (const interceptor of this.responseInterceptors) {\n      r = await interceptor(config, r);\n    }\n    return r;\n  }\n\n")

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
			// Replace all function return type references
			if strings.HasPrefix(tsType, "[") && strings.HasSuffix(tsType, "]") {
				// 形如 [FlowIDTableStaking.DelegatorInfo] -> FlowIDTableStakingDelegatorInfo[]
				inner := strings.TrimPrefix(strings.TrimSuffix(tsType, "]"), "[")
				inner = flattenStructName(strings.TrimSpace(inner))
				tsType = inner + "[]"
			} else if strings.HasSuffix(tsType, "| undefined") {
				// 形如 FlowIDTableStaking.DelegatorInfo | undefined
				base := strings.TrimSuffix(tsType, "| undefined")
				base = flattenStructName(strings.TrimSpace(base))
				tsType = base + " | undefined"
			} else {
				tsType = flattenStructName(tsType)
			}
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
	// 4. Close class with single '}'
	buffer.WriteString("}\n")

	return buffer.String(), nil
}
