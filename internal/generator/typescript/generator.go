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
	Report analyzer.Report
}

// New creates a new TypeScript code generator
func New(report analyzer.Report) *Generator {
	return &Generator{
		Report: report,
	}
}

// typeMapping maps Cadence types to TypeScript types
var typeMapping = map[string]string{
	"String":    "string",
	"Int":       "number",
	"UInt":      "number",
	"UInt8":     "number",
	"UInt16":    "number",
	"UInt32":    "number",
	"UInt64":    "string", // Use string for large numbers
	"UInt128":   "string",
	"UInt256":   "string",
	"Int8":      "number",
	"Int16":     "number",
	"Int32":     "number",
	"Int64":     "string", // Use string for large numbers
	"Int128":    "string",
	"Int256":    "string",
	"Bool":      "boolean",
	"Address":   "string",
	"UFix64":    "string",
	"Fix64":     "string",
	"AnyStruct": "any",
}

const interfaceTemplate = `
export interface {{.Name}} {
    {{- range .Fields}}
    {{.Name}}: {{.Type}}{{if .Optional}} | null{{end}};
    {{- end}}
}
`

const enumTemplate = `
{{if .Tag}}
// Generated from Cadence files in {{.Tag}} folder
export namespace CadenceGen.{{.Tag}} {
{{else}}
// Generated from Cadence files
export namespace CadenceGen {
{{end}}
    export type CadenceType = 'query' | 'transaction';

    export interface CadenceTargetType {
        cadenceBase64: string;
        type: CadenceType;
        args: any[];
    }

    {{- range .Cases}}
    export interface {{.Name}}Args {
        {{- range .Parameters}}
        {{.Name}}: {{.Type}}{{if .Optional}} | null{{end}};
        {{- end}}
    }

    export interface {{.Name}}Target extends CadenceTargetType {
        type: {{.Type}};
        {{- if .ReturnType}}
        returnType: {{.ReturnType}};
        {{- end}}
        args: [{{.Name}}Args];
    }
    {{- end}}

    export const targets = {
        {{- range .Cases}}
        {{.Name}}: (args: {{.Name}}Args): {{.Name}}Target => ({
            cadenceBase64: "{{.Base64}}",
            type: "{{.Type}}",
            {{- if .ReturnType}}
            returnType: "{{.ReturnType}}",
            {{- end}}
            args: [args]
        }),
        {{- end}}
    };
}
`

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

// Generate generates TypeScript code for all transactions and scripts
func (g *Generator) Generate() (string, error) {
	var buffer bytes.Buffer
	var cases []struct {
		Name       string
		Parameters []struct {
			Name     string
			Type     string
			Optional bool
		}
		ReturnType string
		Base64     string
		Type       string
	}
	var interfaces []struct {
		Name   string
		Fields []struct {
			Name     string
			Type     string
			Optional bool
		}
	}

	// Generate interfaces from structs
	for name, composite := range g.Report.Structs {
		iface := struct {
			Name   string
			Fields []struct {
				Name     string
				Type     string
				Optional bool
			}
		}{
			Name: name,
			Fields: make([]struct {
				Name     string
				Type     string
				Optional bool
			}, 0),
		}

		for _, field := range composite.Fields {
			tsType, ok := typeMapping[field.TypeStr]
			if !ok {
				tsType = field.TypeStr
			}

			iface.Fields = append(iface.Fields, struct {
				Name     string
				Type     string
				Optional bool
			}{
				Name:     field.Name,
				Type:     tsType,
				Optional: field.Optional,
			})
		}

		interfaces = append(interfaces, iface)
	}

	// Generate interface code
	interfaceTmpl, err := template.New("interface").Parse(interfaceTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse interface template: %w", err)
	}

	for _, iface := range interfaces {
		err = interfaceTmpl.Execute(&buffer, iface)
		if err != nil {
			return "", fmt.Errorf("failed to execute interface template: %w", err)
		}
		buffer.WriteString("\n")
	}

	// Map to store cases by tag
	taggedCases := make(map[string][]struct {
		Name       string
		Parameters []struct {
			Name     string
			Type     string
			Optional bool
		}
		ReturnType string
		Base64     string
		Type       string
	})

	// Generate cases for transactions
	for filename, result := range g.Report.Transactions {
		c := struct {
			Name       string
			Parameters []struct {
				Name     string
				Type     string
				Optional bool
			}
			ReturnType string
			Base64     string
			Type       string
		}{
			Name: formatFunctionName(filename),
			Parameters: make([]struct {
				Name     string
				Type     string
				Optional bool
			}, 0),
			Base64: result.Base64,
			Type:   "transaction",
		}

		for _, param := range result.Parameters {
			tsType, ok := typeMapping[param.TypeStr]
			if !ok {
				tsType = param.TypeStr
			}

			c.Parameters = append(c.Parameters, struct {
				Name     string
				Type     string
				Optional bool
			}{
				Name:     param.Name,
				Type:     tsType,
				Optional: param.Optional,
			})
		}

		if result.Tag != "" {
			taggedCases[result.Tag] = append(taggedCases[result.Tag], c)
		} else {
			cases = append(cases, c)
		}
	}

	// Generate cases for scripts
	for filename, result := range g.Report.Scripts {
		c := struct {
			Name       string
			Parameters []struct {
				Name     string
				Type     string
				Optional bool
			}
			ReturnType string
			Base64     string
			Type       string
		}{
			Name: formatFunctionName(filename),
			Parameters: make([]struct {
				Name     string
				Type     string
				Optional bool
			}, 0),
			Base64: result.Base64,
			Type:   "query",
		}

		if result.ReturnType != "" {
			tsType, ok := typeMapping[result.ReturnType]
			if !ok {
				tsType = result.ReturnType
			}
			c.ReturnType = tsType
		}

		for _, param := range result.Parameters {
			tsType, ok := typeMapping[param.TypeStr]
			if !ok {
				tsType = param.TypeStr
			}

			c.Parameters = append(c.Parameters, struct {
				Name     string
				Type     string
				Optional bool
			}{
				Name:     param.Name,
				Type:     tsType,
				Optional: param.Optional,
			})
		}

		if result.Tag != "" {
			taggedCases[result.Tag] = append(taggedCases[result.Tag], c)
		} else {
			cases = append(cases, c)
		}
	}

	// Generate enum with all cases
	tmpl, err := template.New("enum").Parse(enumTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// First generate the base CadenceGen namespace
	err = tmpl.Execute(&buffer, struct {
		Cases []struct {
			Name       string
			Parameters []struct {
				Name     string
				Type     string
				Optional bool
			}
			ReturnType string
			Base64     string
			Type       string
		}
		Tag string
	}{
		Cases: cases,
		Tag:   "",
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Then generate tagged cases in separate namespaces
	for tag, tagCases := range taggedCases {
		buffer.WriteString("\n")
		err = tmpl.Execute(&buffer, struct {
			Cases []struct {
				Name       string
				Parameters []struct {
					Name     string
					Type     string
					Optional bool
				}
				ReturnType string
				Base64     string
				Type       string
			}
			Tag string
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
