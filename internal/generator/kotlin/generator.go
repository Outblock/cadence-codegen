package kotlin

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/outblock/cadence-codegen/internal/analyzer"
)

// Generator handles Kotlin code generation
type Generator struct {
	Report  analyzer.Report
	Files   map[string]string
	BaseDir string
}

// New creates a new Kotlin code generator
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

// typeMapping maps Cadence types to Kotlin types
var typeMapping = map[string]string{
	"String":      "String",
	"Character":   "String",
	"Int":         "Int",
	"UInt":        "UInt",
	"Int8":        "Byte",
	"Int16":       "Short",
	"Int32":       "Int",
	"Int64":       "Long",
	"UInt8":       "UByte",
	"UInt16":      "UShort",
	"UInt32":      "UInt",
	"UInt64":      "ULong",
	"Int128":      "BigInteger",
	"Int256":      "BigInteger",
	"UInt128":     "BigInteger",
	"UInt256":     "BigInteger",
	"UFix64":      "BigDecimal",
	"Fix64":       "BigDecimal",
	"Bool":        "Boolean",
	"Address":     "FlowAddress",
	"AnyStruct":   "Any",
	"AnyResource": "Any",
	"Type":        "String",
	"StoragePath": "String",
	"PublicPath":  "String",
	"PrivatePath": "String",
	"Path":        "String",
	"Void":        "Unit",
}

// KotlinFunction represents a function in the generated code
type KotlinFunction struct {
	FuncName   string
	ClassName  string
	Parameters []KotlinParameter
	ReturnType string
	Base64     string
	Type       string // "transaction" or "script"
}

// KotlinParameter represents a parameter in Kotlin
type KotlinParameter struct {
	Name     string
	Type     string
	Optional bool
}

// KotlinStruct represents a data class in Kotlin
type KotlinStruct struct {
	Name   string
	Fields []KotlinField
}

// KotlinField represents a field in a Kotlin data class
type KotlinField struct {
	Name     string
	Type     string
	Optional bool
}

const structTemplate = `data class {{.Name}}(
{{- range $i, $f := .Fields}}
    val {{$f.Name}}: {{$f.Type}}{{if $f.Optional}}?{{end}}{{if $f.Optional}} = null{{end}}{{if notLast $i (len $.Fields)}},{{end}}
{{- end}}
)
`

const functionTemplate = `/** Returns the Cadence code for the {{.FuncName}} {{.Type}}. */
fun {{.FuncName}}Code(): String {
    val encoded = "{{.Base64}}"
    return String(Base64.getDecoder().decode(encoded))
}
{{if .Parameters}}
data class {{.ClassName}}Params(
{{- range $i, $p := .Parameters}}
    val {{$p.Name}}: {{$p.Type}}{{if $p.Optional}}?{{end}}{{if $p.Optional}} = null{{end}}{{if notLast $i (len $.Parameters)}},{{end}}
{{- end}}
)
{{end}}`

// formatFunctionName formats the filename into a camelCase function name
func formatFunctionName(filename string) string {
	name := strings.TrimSuffix(filename, ".cdc")
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-'
	})

	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
		if len(parts[i]) > 0 {
			if i == 0 {
				// first part stays lowercase (camelCase)
				continue
			}
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return strings.Join(parts, "")
}

// formatClassName formats the filename into a PascalCase class name
func formatClassName(filename string) string {
	name := strings.TrimSuffix(filename, ".cdc")
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-'
	})

	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return strings.Join(parts, "")
}

// formatFieldName converts a Cadence field name to camelCase for Kotlin
func formatFieldName(name string) string {
	// If it contains separators, split and camelCase
	if strings.ContainsAny(name, "_-") {
		parts := strings.FieldsFunc(name, func(r rune) bool {
			return r == '_' || r == '-'
		})
		for i := range parts {
			if i == 0 {
				parts[i] = strings.ToLower(parts[i])
			} else if len(parts[i]) > 0 {
				parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
			}
		}
		return strings.Join(parts, "")
	}
	// Already camelCase or single word - keep as-is
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

// convertCadenceTypeToKotlin converts a Cadence type to its Kotlin equivalent
func convertCadenceTypeToKotlin(cadenceType string) string {
	cadenceType = strings.TrimSpace(cadenceType)

	// Strip reference markers (&)
	if strings.HasPrefix(cadenceType, "&") {
		return convertCadenceTypeToKotlin(strings.TrimPrefix(cadenceType, "&"))
	}

	// Check if it's an optional type
	if strings.HasSuffix(cadenceType, "?") {
		baseType := strings.TrimSuffix(cadenceType, "?")
		ktType := convertCadenceTypeToKotlin(baseType)
		return ktType + "?"
	}

	// Handle generic types like Capability<...> as Any
	if strings.Contains(cadenceType, "<") {
		return "Any"
	}

	// Check if it's an array type
	if strings.HasPrefix(cadenceType, "[") && strings.HasSuffix(cadenceType, "]") {
		elementType := strings.TrimPrefix(strings.TrimSuffix(cadenceType, "]"), "[")
		elementType = strings.TrimSpace(elementType)
		ktElementType := convertCadenceTypeToKotlin(elementType)
		return fmt.Sprintf("List<%s>", ktElementType)
	}

	// Check if it's a dictionary or intersection type
	if strings.HasPrefix(cadenceType, "{") && strings.HasSuffix(cadenceType, "}") {
		inner := strings.TrimPrefix(strings.TrimSuffix(cadenceType, "}"), "{")
		parts := strings.SplitN(inner, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			keyType := strings.TrimSpace(parts[0])
			valueType := strings.TrimSpace(parts[1])
			ktKeyType := convertCadenceTypeToKotlin(keyType)
			ktValueType := convertCadenceTypeToKotlin(valueType)
			return fmt.Sprintf("Map<%s, %s>", ktKeyType, ktValueType)
		} else if len(parts) >= 1 {
			singleType := strings.TrimSpace(parts[0])
			return convertCadenceTypeToKotlin(singleType)
		}
	}

	// Use the type mapping
	ktType, ok := typeMapping[cadenceType]
	if !ok {
		if strings.Contains(cadenceType, ".") {
			return flattenStructName(cadenceType)
		}
		return cadenceType
	}
	return ktType
}

// needsBigInteger checks if any type in the report requires java.math.BigInteger
func (g *Generator) needsBigInteger() bool {
	for _, s := range g.Report.Structs {
		for _, f := range s.Fields {
			if ktType := convertCadenceTypeToKotlin(f.TypeStr); strings.Contains(ktType, "BigInteger") {
				return true
			}
		}
	}
	for _, r := range g.Report.Transactions {
		for _, p := range r.Parameters {
			if ktType := convertCadenceTypeToKotlin(p.TypeStr); strings.Contains(ktType, "BigInteger") {
				return true
			}
		}
	}
	for _, r := range g.Report.Scripts {
		for _, p := range r.Parameters {
			if ktType := convertCadenceTypeToKotlin(p.TypeStr); strings.Contains(ktType, "BigInteger") {
				return true
			}
		}
		if ktType := convertCadenceTypeToKotlin(r.ReturnType); strings.Contains(ktType, "BigInteger") {
			return true
		}
	}
	return false
}

// needsBigDecimal checks if any type in the report requires java.math.BigDecimal
func (g *Generator) needsBigDecimal() bool {
	for _, s := range g.Report.Structs {
		for _, f := range s.Fields {
			if ktType := convertCadenceTypeToKotlin(f.TypeStr); strings.Contains(ktType, "BigDecimal") {
				return true
			}
		}
	}
	for _, r := range g.Report.Transactions {
		for _, p := range r.Parameters {
			if ktType := convertCadenceTypeToKotlin(p.TypeStr); strings.Contains(ktType, "BigDecimal") {
				return true
			}
		}
	}
	for _, r := range g.Report.Scripts {
		for _, p := range r.Parameters {
			if ktType := convertCadenceTypeToKotlin(p.TypeStr); strings.Contains(ktType, "BigDecimal") {
				return true
			}
		}
		if ktType := convertCadenceTypeToKotlin(r.ReturnType); strings.Contains(ktType, "BigDecimal") {
			return true
		}
	}
	return false
}

// needsFlowAddress checks if any type in the report requires FlowAddress import
func (g *Generator) needsFlowAddress() bool {
	for _, s := range g.Report.Structs {
		for _, f := range s.Fields {
			if ktType := convertCadenceTypeToKotlin(f.TypeStr); strings.Contains(ktType, "FlowAddress") {
				return true
			}
		}
	}
	for _, r := range g.Report.Transactions {
		for _, p := range r.Parameters {
			if ktType := convertCadenceTypeToKotlin(p.TypeStr); strings.Contains(ktType, "FlowAddress") {
				return true
			}
		}
	}
	for _, r := range g.Report.Scripts {
		for _, p := range r.Parameters {
			if ktType := convertCadenceTypeToKotlin(p.TypeStr); strings.Contains(ktType, "FlowAddress") {
				return true
			}
		}
		if ktType := convertCadenceTypeToKotlin(r.ReturnType); strings.Contains(ktType, "FlowAddress") {
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

// Generate generates Kotlin code for all transactions and scripts
func (g *Generator) Generate() (string, error) {
	var buffer bytes.Buffer

	funcMap := template.FuncMap{
		"notLast": func(i, length int) bool {
			return i < length-1
		},
		"len": func(s interface{}) int {
			switch v := s.(type) {
			case []KotlinField:
				return len(v)
			case []KotlinParameter:
				return len(v)
			default:
				return 0
			}
		},
	}

	// Header
	buffer.WriteString("// Code generated by cadence-codegen. DO NOT EDIT.\n\n")
	buffer.WriteString("package cadencegenerated\n\n")

	// Build imports
	var imports []string
	imports = append(imports, "import java.util.Base64")
	if g.needsBigInteger() {
		imports = append(imports, "import java.math.BigInteger")
	}
	if g.needsBigDecimal() {
		imports = append(imports, "import java.math.BigDecimal")
	}
	if g.needsFlowAddress() {
		imports = append(imports, "import org.onflow.flow.models.FlowAddress")
	}
	sort.Strings(imports)

	for _, imp := range imports {
		buffer.WriteString(imp + "\n")
	}

	// Generate structs
	structTmpl, err := template.New("struct").Funcs(funcMap).Parse(structTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse struct template: %w", err)
	}

	var structNames []string
	for name := range g.Report.Structs {
		structNames = append(structNames, name)
	}
	sort.Strings(structNames)

	if len(structNames) > 0 {
		buffer.WriteString("\n// --- Structs ---\n")
	}

	for _, name := range structNames {
		composite := g.Report.Structs[name]
		ktStruct := KotlinStruct{
			Name:   flattenStructName(name),
			Fields: make([]KotlinField, 0),
		}
		for _, field := range composite.Fields {
			ktType := convertCadenceTypeToKotlin(field.TypeStr)
			ktStruct.Fields = append(ktStruct.Fields, KotlinField{
				Name:     formatFieldName(field.Name),
				Type:     ktType,
				Optional: field.Optional,
			})
		}

		buffer.WriteString("\n")
		err = structTmpl.Execute(&buffer, ktStruct)
		if err != nil {
			return "", fmt.Errorf("failed to execute struct template: %w", err)
		}
	}

	// Collect functions
	var functions []KotlinFunction
	taggedFunctions := make(map[string][]KotlinFunction)

	// Transactions
	var txFilenames []string
	for filename := range g.Report.Transactions {
		txFilenames = append(txFilenames, filename)
	}
	sort.Strings(txFilenames)

	for _, filename := range txFilenames {
		result := g.Report.Transactions[filename]
		ktFunc := KotlinFunction{
			FuncName:   formatFunctionName(filename),
			ClassName:  formatClassName(filename),
			Parameters: make([]KotlinParameter, 0),
			Base64:     result.Base64,
			Type:       "transaction",
		}
		for _, param := range result.Parameters {
			ktType := convertCadenceTypeToKotlin(param.TypeStr)
			ktFunc.Parameters = append(ktFunc.Parameters, KotlinParameter{
				Name:     formatFieldName(param.Name),
				Type:     ktType,
				Optional: param.Optional,
			})
		}
		if result.Tag != "" {
			taggedFunctions[result.Tag] = append(taggedFunctions[result.Tag], ktFunc)
		} else {
			functions = append(functions, ktFunc)
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
		ktFunc := KotlinFunction{
			FuncName:   formatFunctionName(filename),
			ClassName:  formatClassName(filename),
			Parameters: make([]KotlinParameter, 0),
			Base64:     result.Base64,
			Type:       "script",
		}
		if result.ReturnType != "" {
			ktFunc.ReturnType = convertCadenceTypeToKotlin(result.ReturnType)
		}
		for _, param := range result.Parameters {
			ktType := convertCadenceTypeToKotlin(param.TypeStr)
			ktFunc.Parameters = append(ktFunc.Parameters, KotlinParameter{
				Name:     formatFieldName(param.Name),
				Type:     ktType,
				Optional: param.Optional,
			})
		}
		if result.Tag != "" {
			taggedFunctions[result.Tag] = append(taggedFunctions[result.Tag], ktFunc)
		} else {
			functions = append(functions, ktFunc)
		}
	}

	// Sort functions by name
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].FuncName < functions[j].FuncName
	})

	// Generate functions
	funcTmpl, err := template.New("function").Funcs(funcMap).Parse(functionTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse function template: %w", err)
	}

	if len(functions) > 0 {
		hasTx := false
		hasScript := false
		for _, f := range functions {
			if f.Type == "transaction" {
				hasTx = true
			} else {
				hasScript = true
			}
		}
		if hasTx {
			buffer.WriteString("\n// --- Transactions ---\n")
		}
		for _, f := range functions {
			if f.Type == "transaction" {
				buffer.WriteString("\n")
				err = funcTmpl.Execute(&buffer, f)
				if err != nil {
					return "", fmt.Errorf("failed to execute function template: %w", err)
				}
			}
		}
		if hasScript {
			buffer.WriteString("\n// --- Scripts ---\n")
		}
		for _, f := range functions {
			if f.Type == "script" {
				buffer.WriteString("\n")
				err = funcTmpl.Execute(&buffer, f)
				if err != nil {
					return "", fmt.Errorf("failed to execute function template: %w", err)
				}
			}
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
			return tagFuncs[i].FuncName < tagFuncs[j].FuncName
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
