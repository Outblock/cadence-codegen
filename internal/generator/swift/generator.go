package swift

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/outblock/cadence-codegen/internal/analyzer"
)

// Generator handles Swift code generation
type Generator struct {
	Report  analyzer.Report
	Files   map[string]string
	BaseDir string
}

// New creates a new Swift code generator
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

// typeMapping maps Cadence types to Swift types
var typeMapping = map[string]string{
	"String":    "String",
	"Int":       "Int",
	"UInt":      "UInt",
	"UInt8":     "UInt8",
	"UInt16":    "UInt16",
	"UInt32":    "UInt32",
	"UInt64":    "UInt64",
	"UInt128":   "BigUInt",
	"UInt256":   "BigUInt",
	"Int8":      "Int8",
	"Int16":     "Int16",
	"Int32":     "Int32",
	"Int64":     "Int64",
	"Int128":    "BigInt",
	"Int256":    "BigInt",
	"Bool":      "Bool",
	"Address":   "Flow.Address",
	"UFix64":    "Decimal",
	"Fix64":     "Decimal",
	"AnyStruct": "Any",
}

// SwiftCase represents a case in the generated enum
type SwiftCase struct {
	Name       string
	Parameters []SwiftParameter
	ReturnType string
	Base64     string
	Type       string
}

// SwiftParameter represents a parameter in Swift
type SwiftParameter struct {
	Name     string
	Type     string
	Optional bool
}

// SwiftStruct represents a struct in Swift
type SwiftStruct struct {
	Name       string
	Fields     []SwiftField
	Implements []string
}

// SwiftField represents a field in a Swift struct
type SwiftField struct {
	Name     string
	Type     string
	Optional bool
}

const structTemplate = `
/// Generated Cadence struct
struct {{.Name}}: Decodable {
    {{- range .Fields}}
    let {{.Name}}: {{.Type}}{{if .Optional}}?{{end}}
    {{- end}}
}
`

const enumTemplate = `
/// Generated from Cadence files{{if .Tag}} in {{.Tag}} folder{{end}}
{{if .Tag}}extension CadenceGen {
    enum {{.Tag}}: CadenceTargetType, MirrorAssociated {
{{else}}enum CadenceGen: CadenceTargetType, MirrorAssociated {
{{end}}
    {{- range .Cases}}
    case {{.Name}}({{- range $index, $param := .Parameters}}{{if $index}}, {{end}}{{$param.Name}}: {{$param.Type}}{{if $param.Optional}}?{{end}}{{- end}})
    {{- end}}
    
    var cadenceBase64: String {
        switch self {
        {{- range .Cases}}
        case .{{.Name}}:
            return "{{.Base64}}"
        {{- end}}
        }
    }
    
    var type: CadenceType {
        switch self {
        {{- range .Cases}}
        case .{{.Name}}:
            return .{{.Type}}
        {{- end}}
        }
    }
    
    var arguments: [Flow.Argument] {
        associatedValues.compactMap { $0.value.toFlowValue() }.toArguments()
    }
    
    var returnType: Decodable.Type {
        if type == .transaction {
            return Flow.ID.self
        }
        
        switch self {
        {{- range .Cases}}
        case .{{.Name}}:
            {{- if .ReturnType}}
            return {{.ReturnType}}.self
            {{- else}}
            return Flow.ID.self
            {{- end}}
        {{- end}}
        }
    }
}{{if .Tag}} }{{end}}`

// formatFunctionName formats the filename into a valid Swift function name
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

// Generate generates Swift code for all transactions and scripts
func (g *Generator) Generate() (string, error) {
	var buffer bytes.Buffer
	var cases []SwiftCase
	var structs []SwiftStruct

	// Map to store cases by tag
	taggedCases := make(map[string][]SwiftCase)

	// Add header
	buffer.WriteString("import Flow\nimport BigInt\n")

	// Generate structs from composite types
	for name, composite := range g.Report.Structs {
		swiftStruct := SwiftStruct{
			Name:       name,
			Fields:     make([]SwiftField, 0),
			Implements: []string{"Decodable"},
		}

		for _, field := range composite.Fields {
			swiftType, ok := typeMapping[field.TypeStr]
			if !ok {
				swiftType = field.TypeStr
			}

			swiftStruct.Fields = append(swiftStruct.Fields, SwiftField{
				Name:     field.Name,
				Type:     swiftType,
				Optional: field.Optional,
			})
		}

		structs = append(structs, swiftStruct)
	}

	// Generate struct code
	structTmpl, err := template.New("struct").Parse(structTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse struct template: %w", err)
	}

	for _, s := range structs {
		err = structTmpl.Execute(&buffer, s)
		if err != nil {
			return "", fmt.Errorf("failed to execute struct template: %w", err)
		}
		buffer.WriteString("\n")
	}

	// Generate cases for transactions
	for filename, result := range g.Report.Transactions {
		swiftCase := SwiftCase{
			Name:       formatFunctionName(filename),
			Parameters: make([]SwiftParameter, 0),
			Base64:     result.Base64,
			Type:       "transaction",
		}

		for _, param := range result.Parameters {
			swiftType, ok := typeMapping[param.TypeStr]
			if !ok {
				swiftType = param.TypeStr
			}

			swiftCase.Parameters = append(swiftCase.Parameters, SwiftParameter{
				Name:     param.Name,
				Type:     swiftType,
				Optional: param.Optional,
			})
		}

		if result.Tag != "" {
			taggedCases[result.Tag] = append(taggedCases[result.Tag], swiftCase)
		} else {
			cases = append(cases, swiftCase)
		}
	}

	// Generate cases for scripts
	for filename, result := range g.Report.Scripts {
		swiftCase := SwiftCase{
			Name:       formatFunctionName(filename),
			Parameters: make([]SwiftParameter, 0),
			Base64:     result.Base64,
			Type:       "query",
		}

		if result.ReturnType != "" {
			swiftType, ok := typeMapping[result.ReturnType]
			if !ok {
				swiftType = result.ReturnType
			}
			swiftCase.ReturnType = swiftType
		}

		for _, param := range result.Parameters {
			swiftType, ok := typeMapping[param.TypeStr]
			if !ok {
				swiftType = param.TypeStr
			}

			swiftCase.Parameters = append(swiftCase.Parameters, SwiftParameter{
				Name:     param.Name,
				Type:     swiftType,
				Optional: param.Optional,
			})
		}

		if result.Tag != "" {
			taggedCases[result.Tag] = append(taggedCases[result.Tag], swiftCase)
		} else {
			cases = append(cases, swiftCase)
		}
	}

	// Generate enum with all cases
	tmpl, err := template.New("enum").Parse(enumTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// First generate the base CadenceGen enum
	err = tmpl.Execute(&buffer, struct {
		Cases []SwiftCase
		Tag   string
	}{
		Cases: cases,
		Tag:   "",
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Then generate tagged cases in separate extensions
	for tag, tagCases := range taggedCases {
		buffer.WriteString("\n")
		err = tmpl.Execute(&buffer, struct {
			Cases []SwiftCase
			Tag   string
		}{
			Cases: tagCases,
			Tag:   tag,
		})
		if err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}
	}

	return buffer.String(), nil
}
