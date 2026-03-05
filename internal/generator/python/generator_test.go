package python

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
		{"String", "str"},
		{"Character", "str"},
		{"Bool", "bool"},
		{"Int", "int"},
		{"UInt", "int"},
		{"Int8", "int"},
		{"Int16", "int"},
		{"Int32", "int"},
		{"Int64", "int"},
		{"UInt8", "int"},
		{"UInt16", "int"},
		{"UInt32", "int"},
		{"UInt64", "int"},
		{"Int128", "int"},
		{"Int256", "int"},
		{"UInt128", "int"},
		{"UInt256", "int"},
		{"UFix64", "Decimal"},
		{"Fix64", "Decimal"},
		{"Address", "str"},
		{"AnyStruct", "Any"},
		{"AnyResource", "Any"},
		{"Path", "str"},
		{"StoragePath", "str"},
		{"PublicPath", "str"},
		{"PrivatePath", "str"},
		{"Type", "str"},
		{"Void", "None"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToPython(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToPython(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestArrayType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"[String]", "list[str]"},
		{"[UInt64]", "list[int]"},
		{"[Address]", "list[str]"},
		{"[Int128]", "list[int]"},
		{"[[String]]", "list[list[str]]"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToPython(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToPython(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestDictionaryType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"{String: UInt64}", "dict[str, int]"},
		{"{Address: Bool}", "dict[str, bool]"},
		{"{String: [UInt64]}", "dict[str, list[int]]"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToPython(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToPython(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestOptionalType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"String?", "Optional[str]"},
		{"UInt64?", "Optional[int]"},
		{"Int128?", "Optional[int]"},
		{"AnyStruct?", "Optional[Any]"},
		{"[String]?", "Optional[list[str]]"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToPython(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToPython(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestReferenceType(t *testing.T) {
	result := convertCadenceTypeToPython("&String")
	if result != "str" {
		t.Errorf("convertCadenceTypeToPython(\"&String\") = %q, want %q", result, "str")
	}
}

func TestGenericType(t *testing.T) {
	result := convertCadenceTypeToPython("Capability<&Something>")
	if result != "Any" {
		t.Errorf("convertCadenceTypeToPython(\"Capability<&Something>\") = %q, want %q", result, "Any")
	}
}

func TestNestedType(t *testing.T) {
	result := convertCadenceTypeToPython("FlowToken.Vault")
	if result != "FlowTokenVault" {
		t.Errorf("convertCadenceTypeToPython(\"FlowToken.Vault\") = %q, want %q", result, "FlowTokenVault")
	}
}

func TestIntersectionType(t *testing.T) {
	result := convertCadenceTypeToPython("{FungibleToken}")
	if result != "FungibleToken" {
		t.Errorf("convertCadenceTypeToPython(\"{FungibleToken}\") = %q, want %q", result, "FungibleToken")
	}
}

func TestFormatFunctionName(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"transfer_tokens.cdc", "transfer_tokens"},
		{"get-balance.cdc", "get_balance"},
		{"simple.cdc", "simple"},
		{"my_cool_script.cdc", "my_cool_script"},
		{"UPPER_CASE.cdc", "upper_case"},
		{"already-Good.cdc", "already_good"},
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

func TestFormatClassName(t *testing.T) {
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
			result := formatClassName(tt.filename)
			if result != tt.expected {
				t.Errorf("formatClassName(%q) = %q, want %q", tt.filename, result, tt.expected)
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

	if !strings.Contains(code, "# Code generated by cadence-codegen. DO NOT EDIT.") {
		t.Error("expected generated header comment")
	}
	if !strings.Contains(code, "from __future__ import annotations") {
		t.Error("expected future annotations import")
	}
	if !strings.Contains(code, "import base64") {
		t.Error("expected base64 import")
	}
	if !strings.Contains(code, "from dataclasses import dataclass") {
		t.Error("expected dataclass import")
	}
	// Should NOT have Decimal import when not needed
	if strings.Contains(code, "from decimal import Decimal") {
		t.Error("should not import Decimal when not needed")
	}
	// Should NOT have Optional import when not needed
	if strings.Contains(code, "Optional") {
		t.Error("should not import Optional when not needed")
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

	if !strings.Contains(code, "def transfer_tokens_code() -> str:") {
		t.Error("expected transfer_tokens_code function")
	}
	if !strings.Contains(code, b64) {
		t.Error("expected base64 encoded code")
	}
	if !strings.Contains(code, "class TransferTokensParams:") {
		t.Error("expected TransferTokensParams dataclass")
	}
	if !strings.Contains(code, "amount: Decimal") {
		t.Error("expected amount field with Decimal type (UFix64)")
	}
	if !strings.Contains(code, "to: str") {
		t.Error("expected to field with str type (Address)")
	}
	if !strings.Contains(code, "from decimal import Decimal") {
		t.Error("expected Decimal import when UFix64 is used")
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

	if !strings.Contains(code, "def get_balance_code() -> str:") {
		t.Error("expected get_balance_code function")
	}
	if !strings.Contains(code, "class GetBalanceParams:") {
		t.Error("expected GetBalanceParams dataclass")
	}
	if !strings.Contains(code, `Returns the Cadence code for the get_balance script.`) {
		t.Error("expected script type in docstring")
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

	if !strings.Contains(code, "class TokenInfo:") {
		t.Error("expected TokenInfo dataclass")
	}
	if !strings.Contains(code, "name: str") {
		t.Error("expected name field")
	}
	if !strings.Contains(code, "balance: Decimal") {
		t.Error("expected balance field (UFix64 -> Decimal)")
	}
	if !strings.Contains(code, "owner: Optional[str] = None") {
		t.Error("expected owner optional field with default None")
	}
	if !strings.Contains(code, "# --- Structs ---") {
		t.Error("expected Structs section header")
	}
	if !strings.Contains(code, "from decimal import Decimal") {
		t.Error("expected Decimal import")
	}
	if !strings.Contains(code, "Optional") {
		t.Error("expected Optional import")
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

	if !strings.Contains(code, "# --- staking ---") {
		t.Error("expected staking tag section")
	}
	if !strings.Contains(code, "stake_code") {
		t.Error("expected stake_code function")
	}
	if !strings.Contains(code, "get_stake_code") {
		t.Error("expected get_stake_code function")
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

	if !strings.Contains(code, "def init_account_code() -> str:") {
		t.Error("expected init_account_code function")
	}
	// Should NOT have params dataclass for no-parameter functions
	if strings.Contains(code, "InitAccountParams") {
		t.Error("should not generate params dataclass for function with no parameters")
	}
}

func TestNeedsDecimal(t *testing.T) {
	// Report with UFix64 should need Decimal
	report := analyzer.Report{
		Transactions: make(map[string]analyzer.AnalysisResult),
		Scripts: map[string]analyzer.AnalysisResult{
			"test.cdc": {
				FileName: "test.cdc",
				Type:     "script",
				Parameters: []analyzer.Parameter{
					{Name: "val", TypeStr: "UFix64", Optional: false},
				},
				Base64: base64.StdEncoding.EncodeToString([]byte("access(all) fun main(val: UFix64) {}")),
			},
		},
		Structs: make(map[string]analyzer.Struct),
	}

	gen := New(report)
	if !gen.needsDecimal() {
		t.Error("expected needsDecimal to return true for UFix64")
	}

	// Report with Fix64 should also need Decimal
	report2 := analyzer.Report{
		Transactions: make(map[string]analyzer.AnalysisResult),
		Scripts:      make(map[string]analyzer.AnalysisResult),
		Structs: map[string]analyzer.Struct{
			"Info": {
				Name: "Info",
				Fields: []analyzer.Field{
					{Name: "rate", TypeStr: "Fix64", Optional: false},
				},
			},
		},
	}

	gen2 := New(report2)
	if !gen2.needsDecimal() {
		t.Error("expected needsDecimal to return true for Fix64")
	}

	// Report without fixed-point types should not need Decimal
	report3 := analyzer.Report{
		Transactions: make(map[string]analyzer.AnalysisResult),
		Scripts: map[string]analyzer.AnalysisResult{
			"test.cdc": {
				FileName: "test.cdc",
				Type:     "script",
				Parameters: []analyzer.Parameter{
					{Name: "val", TypeStr: "UInt64", Optional: false},
				},
				Base64: base64.StdEncoding.EncodeToString([]byte("access(all) fun main(val: UInt64) {}")),
			},
		},
		Structs: make(map[string]analyzer.Struct),
	}

	gen3 := New(report3)
	if gen3.needsDecimal() {
		t.Error("expected needsDecimal to return false for UInt64")
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
