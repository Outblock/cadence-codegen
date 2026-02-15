package golang

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
		{"Bool", "bool"},
		{"Int", "int"},
		{"UInt", "uint"},
		{"Int8", "int8"},
		{"Int16", "int16"},
		{"Int32", "int32"},
		{"Int64", "int64"},
		{"UInt8", "uint8"},
		{"UInt16", "uint16"},
		{"UInt32", "uint32"},
		{"UInt64", "uint64"},
		{"Int128", "*big.Int"},
		{"Int256", "*big.Int"},
		{"UInt128", "*big.Int"},
		{"UInt256", "*big.Int"},
		{"UFix64", "string"},
		{"Fix64", "string"},
		{"Address", "string"},
		{"AnyStruct", "interface{}"},
		{"AnyResource", "interface{}"},
		{"Path", "string"},
		{"StoragePath", "string"},
		{"PublicPath", "string"},
		{"PrivatePath", "string"},
		{"Type", "string"},
		{"Void", ""},
		{"Character", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToGo(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToGo(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestArrayType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"[String]", "[]string"},
		{"[UInt64]", "[]uint64"},
		{"[Address]", "[]string"},
		{"[Int128]", "[]*big.Int"},
		{"[[String]]", "[][]string"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToGo(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToGo(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestDictionaryType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"{String: UInt64}", "map[string]uint64"},
		{"{Address: Bool}", "map[string]bool"},
		{"{String: [UInt64]}", "map[string][]uint64"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToGo(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToGo(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestOptionalType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"String?", "*string"},
		{"UInt64?", "*uint64"},
		{"Int128?", "*big.Int"},       // already a pointer
		{"AnyStruct?", "interface{}"}, // interface{} already nullable
		{"[String]?", "*[]string"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToGo(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToGo(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestReferenceType(t *testing.T) {
	result := convertCadenceTypeToGo("&String")
	if result != "string" {
		t.Errorf("convertCadenceTypeToGo(\"&String\") = %q, want %q", result, "string")
	}
}

func TestGenericType(t *testing.T) {
	result := convertCadenceTypeToGo("Capability<&Something>")
	if result != "interface{}" {
		t.Errorf("convertCadenceTypeToGo(\"Capability<&Something>\") = %q, want %q", result, "interface{}")
	}
}

func TestNestedType(t *testing.T) {
	result := convertCadenceTypeToGo("FlowToken.Vault")
	if result != "FlowTokenVault" {
		t.Errorf("convertCadenceTypeToGo(\"FlowToken.Vault\") = %q, want %q", result, "FlowTokenVault")
	}
}

func TestFormatFunctionName(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"transfer_tokens.cdc", "TransferTokens"},
		{"get-balance.cdc", "GetBalance"},
		{"simple.cdc", "Simple"},
		{"my_cool_script.cdc", "MyCoolScript"},
		{"UPPER_CASE.cdc", "UpperCase"},
		{"already-Good.cdc", "AlreadyGood"},
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

func TestFormatFieldName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"name", "Name"},
		{"token_balance", "TokenBalance"},
		{"myField", "MyField"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatFieldName(tt.input)
			if result != tt.expected {
				t.Errorf("formatFieldName(%q) = %q, want %q", tt.input, result, tt.expected)
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

	if !strings.Contains(code, "package cadence_generated") {
		t.Error("expected package declaration")
	}
	if !strings.Contains(code, "Code generated by cadence-codegen. DO NOT EDIT.") {
		t.Error("expected generated header comment")
	}
	// Should have base64 import blank identifier to suppress unused warning
	if !strings.Contains(code, "var _ = base64.StdEncoding") {
		t.Error("expected base64 blank identifier for empty report")
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

	if !strings.Contains(code, "func TransferTokensCode() string") {
		t.Error("expected TransferTokensCode function")
	}
	if !strings.Contains(code, b64) {
		t.Error("expected base64 encoded code")
	}
	if !strings.Contains(code, "TransferTokensParams") {
		t.Error("expected TransferTokensParams struct")
	}
	if !strings.Contains(code, "Amount string") {
		t.Error("expected Amount field with string type (UFix64)")
	}
	if !strings.Contains(code, "To string") {
		t.Error("expected To field with string type (Address)")
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

	if !strings.Contains(code, "func GetBalanceCode() string") {
		t.Error("expected GetBalanceCode function")
	}
	if !strings.Contains(code, "GetBalanceParams") {
		t.Error("expected GetBalanceParams struct")
	}
	if !strings.Contains(code, "// GetBalanceCode returns the Cadence code for the GetBalance script.") {
		t.Error("expected script type in comment")
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

	if !strings.Contains(code, "type TokenInfo struct") {
		t.Error("expected TokenInfo struct")
	}
	if !strings.Contains(code, "Name string") {
		t.Error("expected Name field")
	}
	if !strings.Contains(code, "Balance string") {
		t.Error("expected Balance field (UFix64 -> string)")
	}
	if !strings.Contains(code, "Owner *string") {
		t.Error("expected Owner optional field")
	}
	if !strings.Contains(code, `json:"name"`) {
		t.Error("expected JSON tag for name field")
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

	if !strings.Contains(code, "// --- staking ---") {
		t.Error("expected staking tag section")
	}
	if !strings.Contains(code, "StakeCode") {
		t.Error("expected StakeCode function")
	}
	if !strings.Contains(code, "GetStakeCode") {
		t.Error("expected GetStakeCode function")
	}
}

func TestGenerateWithBigIntImport(t *testing.T) {
	report := analyzer.Report{
		Transactions: make(map[string]analyzer.AnalysisResult),
		Scripts: map[string]analyzer.AnalysisResult{
			"get_big.cdc": {
				FileName: "get_big.cdc",
				Type:     "script",
				Parameters: []analyzer.Parameter{
					{Name: "val", TypeStr: "Int128", Optional: false},
				},
				Base64: base64.StdEncoding.EncodeToString([]byte("access(all) fun main(val: Int128) {}")),
			},
		},
		Structs: make(map[string]analyzer.Struct),
	}

	gen := New(report)
	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	if !strings.Contains(code, `"math/big"`) {
		t.Error("expected math/big import when Int128 is used")
	}
}

func TestGenerateNoParamsFunction(t *testing.T) {
	report := analyzer.Report{
		Transactions: map[string]analyzer.AnalysisResult{
			"init_account.cdc": {
				FileName:   "init_account.cdc",
				Type:       "transaction",
				Parameters: []analyzer.Parameter{},
				Base64:     base64.StdEncoding.EncodeToString([]byte("transaction { execute {} }")),
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

	if !strings.Contains(code, "func InitAccountCode() string") {
		t.Error("expected InitAccountCode function")
	}
	// Should NOT have params struct for no-parameter functions
	if strings.Contains(code, "InitAccountParams") {
		t.Error("should not generate params struct for function with no parameters")
	}
}

func TestSetBaseDir(t *testing.T) {
	report := analyzer.Report{}
	gen := New(report)
	gen.SetBaseDir("/some/dir")
	if gen.BaseDir != "/some/dir" {
		t.Errorf("SetBaseDir did not set BaseDir correctly")
	}
}

func TestIntersectionType(t *testing.T) {
	result := convertCadenceTypeToGo("{FungibleToken}")
	if result != "FungibleToken" {
		t.Errorf("convertCadenceTypeToGo(\"{FungibleToken}\") = %q, want %q", result, "FungibleToken")
	}
}
