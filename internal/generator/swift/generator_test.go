package swift

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
		{"String", "String"},
		{"Character", "String"},
		{"Bool", "Bool"},
		{"Int", "Int"},
		{"UInt", "UInt"},
		{"Int8", "Int8"},
		{"Int16", "Int16"},
		{"Int32", "Int32"},
		{"Int64", "Int64"},
		{"UInt8", "UInt8"},
		{"UInt16", "UInt16"},
		{"UInt32", "UInt32"},
		{"UInt64", "UInt64"},
		{"Int128", "BigInt"},
		{"Int256", "BigInt"},
		{"UInt128", "BigUInt"},
		{"UInt256", "BigUInt"},
		{"UFix64", "Decimal"},
		{"Fix64", "Decimal"},
		{"Address", "Flow.Address"},
		{"AnyStruct", "AnyDecodable"},
		{"AnyResource", "AnyDecodable"},
		{"Path", "String"},
		{"StoragePath", "String"},
		{"PublicPath", "String"},
		{"PrivatePath", "String"},
		{"Type", "String"},
		{"Void", "Void"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToSwift(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToSwift(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestArrayType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"[String]", "[String]"},
		{"[UInt64]", "[UInt64]"},
		{"[Address]", "[Flow.Address]"},
		{"[Int128]", "[BigInt]"},
		{"[[String]]", "[[String]]"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToSwift(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToSwift(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestDictionaryType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"{String: UInt64}", "Dictionary<String, UInt64>"},
		{"{Address: Bool}", "Dictionary<Flow.Address, Bool>"},
		{"{String: [UInt64]}", "Dictionary<String, [UInt64]>"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToSwift(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToSwift(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestOptionalType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"String?", "String?"},
		{"UInt64?", "UInt64?"},
		{"Int128?", "BigInt?"},
		{"[String]?", "[String]?"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToSwift(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToSwift(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestReferenceType(t *testing.T) {
	result := convertCadenceTypeToSwift("&String")
	if result != "String" {
		t.Errorf("convertCadenceTypeToSwift(\"&String\") = %q, want %q", result, "String")
	}
}

func TestGenericType(t *testing.T) {
	result := convertCadenceTypeToSwift("Capability<&Something>")
	if result != "AnyDecodable" {
		t.Errorf("convertCadenceTypeToSwift(\"Capability<&Something>\") = %q, want %q", result, "AnyDecodable")
	}
}

func TestNestedType(t *testing.T) {
	result := convertCadenceTypeToSwift("FlowToken.Vault")
	if result != "FlowTokenVault" {
		t.Errorf("convertCadenceTypeToSwift(\"FlowToken.Vault\") = %q, want %q", result, "FlowTokenVault")
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

	if !strings.Contains(code, "import Flow") {
		t.Error("expected Flow import")
	}
	if !strings.Contains(code, "CadenceGen") {
		t.Error("expected CadenceGen enum")
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
		t.Error("expected transferTokens case name")
	}
	if !strings.Contains(code, "amount: Decimal") {
		t.Error("expected amount parameter with Decimal type (UFix64)")
	}
	if !strings.Contains(code, "to: Flow.Address") {
		t.Error("expected to parameter with Flow.Address type (Address)")
	}
	if !strings.Contains(code, ".transaction") {
		t.Error("expected .transaction type")
	}
	if !strings.Contains(code, b64) {
		t.Error("expected base64 encoded code")
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
		t.Error("expected getBalance case name")
	}
	if !strings.Contains(code, ".query") {
		t.Error("expected .query type")
	}
	if !strings.Contains(code, "Decimal.self") {
		t.Error("expected Decimal return type (UFix64)")
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

	if !strings.Contains(code, "struct TokenInfo") {
		t.Error("expected TokenInfo struct")
	}
	if !strings.Contains(code, "Decodable") {
		t.Error("expected Decodable conformance")
	}
	if !strings.Contains(code, "let name: String") {
		t.Error("expected name field with String type")
	}
	if !strings.Contains(code, "let balance: Decimal") {
		t.Error("expected balance field with Decimal type (UFix64)")
	}
	if !strings.Contains(code, "let owner: Flow.Address?") {
		t.Error("expected owner optional field with Flow.Address type")
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

	if !strings.Contains(code, "extension CadenceGen") {
		t.Error("expected staking tag extension")
	}
	if !strings.Contains(code, "enum staking") {
		t.Error("expected staking enum inside extension")
	}
	if !strings.Contains(code, "stake") {
		t.Error("expected stake case")
	}
	if !strings.Contains(code, "getStake") {
		t.Error("expected getStake case")
	}
}

func TestIntersectionType(t *testing.T) {
	result := convertCadenceTypeToSwift("{FungibleToken}")
	if result != "FungibleToken" {
		t.Errorf("convertCadenceTypeToSwift(\"{FungibleToken}\") = %q, want %q", result, "FungibleToken")
	}
}
