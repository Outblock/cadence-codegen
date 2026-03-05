package typescript

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/outblock/cadence-codegen/internal/analyzer"
)

func TestTypeMapping(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"String", "string"},
		{"Character", "string"},
		{"Bool", "boolean"},
		{"Int", "number"},
		{"UInt", "number"},
		{"Int8", "number"},
		{"Int16", "number"},
		{"Int32", "number"},
		{"Int64", "number"},
		{"UInt8", "number"},
		{"UInt16", "number"},
		{"UInt32", "number"},
		{"UInt64", "number"},
		{"Int128", "string"},
		{"Int256", "string"},
		{"UInt128", "string"},
		{"UInt256", "string"},
		{"UFix64", "string"},
		{"Fix64", "string"},
		{"Address", "string"},
		{"AnyStruct", "any"},
		{"AnyResource", "any"},
		{"Path", "string"},
		{"StoragePath", "string"},
		{"PublicPath", "string"},
		{"PrivatePath", "string"},
		{"Type", "string"},
		{"Void", "void"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToTypeScript(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToTypeScript(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestArrayType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"[String]", "string[]"},
		{"[UInt64]", "number[]"},
		{"[Address]", "string[]"},
		{"[Int128]", "string[]"},
		{"[[String]]", "string[][]"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToTypeScript(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToTypeScript(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestDictionaryType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"{String: UInt64}", "Record<string, number>"},
		{"{Address: Bool}", "Record<string, boolean>"},
		{"{String: [UInt64]}", "Record<string, number[]>"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToTypeScript(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToTypeScript(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestOptionalType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"String?", "string | undefined"},
		{"UInt64?", "number | undefined"},
		{"Int128?", "string | undefined"},
		{"AnyStruct?", "any | undefined"},
		{"[String]?", "string[] | undefined"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToTypeScript(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToTypeScript(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestReferenceType(t *testing.T) {
	result := convertCadenceTypeToTypeScript("&String")
	if result != "string" {
		t.Errorf("convertCadenceTypeToTypeScript(\"&String\") = %q, want %q", result, "string")
	}
}

func TestGenericType(t *testing.T) {
	result := convertCadenceTypeToTypeScript("Capability<&Something>")
	if result != "any" {
		t.Errorf("convertCadenceTypeToTypeScript(\"Capability<&Something>\") = %q, want %q", result, "any")
	}
}

func TestNestedType(t *testing.T) {
	result := convertCadenceTypeToTypeScript("FlowToken.Vault")
	if result != "FlowTokenVault" {
		t.Errorf("convertCadenceTypeToTypeScript(\"FlowToken.Vault\") = %q, want %q", result, "FlowTokenVault")
	}
}

func TestFormatFunctionName(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"transfer_tokens.cdc", "transferTokens"},
		{"get-balance.cdc", "getBalance"},
		{"simple.cdc", "simple"},
		{"my_cool_script.cdc", "myCoolScript"},
		{"UPPER_CASE.cdc", "upperCase"},
		{"already-Good.cdc", "alreadyGood"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := formatFunctionName(tt.filename)
			if result != tt.expected {
				t.Errorf("formatFunctionName(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestFlattenStructName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple", "Simple"},
		{"Flow.Token", "FlowToken"},
		{"A.B.C", "ABC"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := flattenStructName(tt.input)
			if result != tt.expected {
				t.Errorf("flattenStructName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetFCLType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"String", "t.String"},
		{"Bool", "t.Bool"},
		{"UInt64", "t.UInt64"},
		{"Address", "t.Address"},
		{"UFix64", "t.UFix64"},
		{"UInt128", "t.UInt128"},
		{"Int256", "t.Int256"},
		{"AnyStruct", "t.Any"},
		{"[String]", "t.Array(t.String)"},
		{"[UInt64]", "t.Array(t.UInt64)"},
		{"{String: UInt64}", "t.Dictionary({ key: t.String, value: t.UInt64 })"},
		{"String?", "t.String"},
		{"&String", "t.String"},
		{"{FungibleToken}", "t.FungibleToken"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := getFCLType(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("getFCLType(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestGenerateEmpty(t *testing.T) {
	report := analyzer.Report{
		Transactions: make(map[string]analyzer.AnalysisResult),
		Scripts:      make(map[string]analyzer.AnalysisResult),
		Structs:      make(map[string]analyzer.Struct),
	}

	gen := New(report)
	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if !strings.Contains(code, `import * as fcl from "@onflow/fcl"`) {
		t.Error("expected fcl import")
	}
	if !strings.Contains(code, "export class CadenceService") {
		t.Error("expected CadenceService class")
	}
}

func TestGenerateWithTransaction(t *testing.T) {
	cadenceCode := "transaction(amount: UFix64, to: Address) { execute {} }"
	b64 := base64.StdEncoding.EncodeToString([]byte(cadenceCode))

	report := analyzer.Report{
		Transactions: map[string]analyzer.AnalysisResult{
			"transfer_tokens.cdc": {
				FileName: "transfer_tokens.cdc",
				Type:     "transaction",
				Parameters: []analyzer.Parameter{
					{Name: "amount", TypeStr: "UFix64", Optional: false},
					{Name: "to", TypeStr: "Address", Optional: false},
				},
				Base64: b64,
			},
		},
		Scripts: make(map[string]analyzer.AnalysisResult),
		Structs: make(map[string]analyzer.Struct),
	}

	gen := New(report)
	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if !strings.Contains(code, "transferTokens") {
		t.Error("expected transferTokens function name")
	}
	if !strings.Contains(code, "fcl.mutate") {
		t.Error("expected fcl.mutate for transaction")
	}
	if !strings.Contains(code, "amount: string") {
		t.Error("expected amount parameter with string type (UFix64)")
	}
	if !strings.Contains(code, "to: string") {
		t.Error("expected to parameter with string type (Address)")
	}
}

func TestGenerateWithScript(t *testing.T) {
	cadenceCode := "access(all) fun main(addr: Address): UFix64 { return 0.0 }"
	b64 := base64.StdEncoding.EncodeToString([]byte(cadenceCode))

	report := analyzer.Report{
		Transactions: make(map[string]analyzer.AnalysisResult),
		Scripts: map[string]analyzer.AnalysisResult{
			"get_balance.cdc": {
				FileName:   "get_balance.cdc",
				Type:       "script",
				ReturnType: "UFix64",
				Parameters: []analyzer.Parameter{
					{Name: "addr", TypeStr: "Address", Optional: false},
				},
				Base64: b64,
			},
		},
		Structs: make(map[string]analyzer.Struct),
	}

	gen := New(report)
	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if !strings.Contains(code, "getBalance") {
		t.Error("expected getBalance function name")
	}
	if !strings.Contains(code, "fcl.query") {
		t.Error("expected fcl.query for script")
	}
	if !strings.Contains(code, "Promise<string>") {
		t.Error("expected Promise<string> return type (UFix64)")
	}
}

func TestGenerateWithStruct(t *testing.T) {
	report := analyzer.Report{
		Transactions: make(map[string]analyzer.AnalysisResult),
		Scripts:      make(map[string]analyzer.AnalysisResult),
		Structs: map[string]analyzer.Struct{
			"TokenInfo": {
				Name: "TokenInfo",
				Fields: []analyzer.Field{
					{Name: "name", TypeStr: "String", Optional: false},
					{Name: "balance", TypeStr: "UFix64", Optional: false},
					{Name: "owner", TypeStr: "Address", Optional: true},
				},
			},
		},
	}

	gen := New(report)
	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if !strings.Contains(code, "export interface TokenInfo") {
		t.Error("expected TokenInfo interface")
	}
	if !strings.Contains(code, "name: string") {
		t.Error("expected name field with string type")
	}
	if !strings.Contains(code, "balance: string") {
		t.Error("expected balance field with string type (UFix64)")
	}
	if !strings.Contains(code, "owner?: string") {
		t.Error("expected owner optional field")
	}
}

func TestGenerateWithTags(t *testing.T) {
	report := analyzer.Report{
		Transactions: map[string]analyzer.AnalysisResult{
			"stake.cdc": {
				FileName:   "stake.cdc",
				Type:       "transaction",
				Parameters: []analyzer.Parameter{},
				Base64:     base64.StdEncoding.EncodeToString([]byte("transaction {}")),
				Tag:        "staking",
			},
		},
		Scripts: map[string]analyzer.AnalysisResult{
			"get_stake.cdc": {
				FileName:   "get_stake.cdc",
				Type:       "script",
				Parameters: []analyzer.Parameter{},
				Base64:     base64.StdEncoding.EncodeToString([]byte("access(all) fun main() {}")),
				Tag:        "staking",
			},
		},
		Structs: make(map[string]analyzer.Struct),
	}

	gen := New(report)
	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if !strings.Contains(code, "Tag: staking") {
		t.Error("expected staking tag section")
	}
	if !strings.Contains(code, "stake") {
		t.Error("expected stake function")
	}
	if !strings.Contains(code, "getStake") {
		t.Error("expected getStake function")
	}
}

func TestDecodeBase64ToUTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"invalid base64", "not-valid-base64!!!", ""},
		{"valid base64", base64.StdEncoding.EncodeToString([]byte("hello world")), "hello world"},
		{"escapes backtick", base64.StdEncoding.EncodeToString([]byte("code with `backtick`")), "code with \\`backtick\\`"},
		{"escapes dollar", base64.StdEncoding.EncodeToString([]byte("cost is $100")), "cost is \\$100"},
		{"escapes backslash", base64.StdEncoding.EncodeToString([]byte("path\\to\\file")), "path\\\\to\\\\file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeBase64ToUTF8(tt.input)
			if result != tt.expected {
				t.Errorf("decodeBase64ToUTF8(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIntersectionType(t *testing.T) {
	result := convertCadenceTypeToTypeScript("{FungibleToken}")
	if result != "any" {
		t.Errorf("convertCadenceTypeToTypeScript(\"{FungibleToken}\") = %q, want %q", result, "any")
	}
}
