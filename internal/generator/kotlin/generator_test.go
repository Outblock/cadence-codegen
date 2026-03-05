package kotlin

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
		{"Bool", "Boolean"},
		{"Int", "Int"},
		{"UInt", "UInt"},
		{"Int8", "Byte"},
		{"Int16", "Short"},
		{"Int32", "Int"},
		{"Int64", "Long"},
		{"UInt8", "UByte"},
		{"UInt16", "UShort"},
		{"UInt32", "UInt"},
		{"UInt64", "ULong"},
		{"Int128", "BigInteger"},
		{"Int256", "BigInteger"},
		{"UInt128", "BigInteger"},
		{"UInt256", "BigInteger"},
		{"UFix64", "BigDecimal"},
		{"Fix64", "BigDecimal"},
		{"Address", "FlowAddress"},
		{"AnyStruct", "Any"},
		{"AnyResource", "Any"},
		{"Path", "String"},
		{"StoragePath", "String"},
		{"PublicPath", "String"},
		{"PrivatePath", "String"},
		{"Type", "String"},
		{"Void", "Unit"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToKotlin(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToKotlin(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestArrayType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"[String]", "List<String>"},
		{"[UInt64]", "List<ULong>"},
		{"[Address]", "List<FlowAddress>"},
		{"[Int128]", "List<BigInteger>"},
		{"[[String]]", "List<List<String>>"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToKotlin(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToKotlin(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestDictionaryType(t *testing.T) {
	tests := []struct {
		cadenceType string
		expected    string
	}{
		{"{String: UInt64}", "Map<String, ULong>"},
		{"{Address: Bool}", "Map<FlowAddress, Boolean>"},
		{"{String: [UInt64]}", "Map<String, List<ULong>>"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToKotlin(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToKotlin(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
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
		{"UInt64?", "ULong?"},
		{"Int128?", "BigInteger?"},
		{"AnyStruct?", "Any?"},
		{"[String]?", "List<String>?"},
	}

	for _, tt := range tests {
		t.Run(tt.cadenceType, func(t *testing.T) {
			result := convertCadenceTypeToKotlin(tt.cadenceType)
			if result != tt.expected {
				t.Errorf("convertCadenceTypeToKotlin(%q) = %q, want %q", tt.cadenceType, result, tt.expected)
			}
		})
	}
}

func TestReferenceType(t *testing.T) {
	result := convertCadenceTypeToKotlin("&String")
	if result != "String" {
		t.Errorf("convertCadenceTypeToKotlin(\"&String\") = %q, want %q", result, "String")
	}
}

func TestGenericType(t *testing.T) {
	result := convertCadenceTypeToKotlin("Capability<&Something>")
	if result != "Any" {
		t.Errorf("convertCadenceTypeToKotlin(\"Capability<&Something>\") = %q, want %q", result, "Any")
	}
}

func TestNestedType(t *testing.T) {
	result := convertCadenceTypeToKotlin("FlowToken.Vault")
	if result != "FlowTokenVault" {
		t.Errorf("convertCadenceTypeToKotlin(\"FlowToken.Vault\") = %q, want %q", result, "FlowTokenVault")
	}
}

func TestIntersectionType(t *testing.T) {
	result := convertCadenceTypeToKotlin("{FungibleToken}")
	if result != "FungibleToken" {
		t.Errorf("convertCadenceTypeToKotlin(\"{FungibleToken}\") = %q, want %q", result, "FungibleToken")
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

	if !strings.Contains(code, "package cadencegenerated") {
		t.Error("expected package declaration")
	}
	if !strings.Contains(code, "Code generated by cadence-codegen. DO NOT EDIT.") {
		t.Error("expected generated header comment")
	}
	if !strings.Contains(code, "import java.util.Base64") {
		t.Error("expected Base64 import")
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

	if !strings.Contains(code, "fun transferTokensCode(): String") {
		t.Error("expected transferTokensCode function")
	}
	if !strings.Contains(code, b64) {
		t.Error("expected base64 encoded code")
	}
	if !strings.Contains(code, "TransferTokensParams") {
		t.Error("expected TransferTokensParams data class")
	}
	if !strings.Contains(code, "amount: BigDecimal") {
		t.Error("expected amount field with BigDecimal type (UFix64)")
	}
	if !strings.Contains(code, "to: FlowAddress") {
		t.Error("expected to field with FlowAddress type (Address)")
	}
	if !strings.Contains(code, "import java.math.BigDecimal") {
		t.Error("expected BigDecimal import when UFix64 is used")
	}
	if !strings.Contains(code, "import org.onflow.flow.models.FlowAddress") {
		t.Error("expected FlowAddress import when Address is used")
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

	if !strings.Contains(code, "fun getBalanceCode(): String") {
		t.Error("expected getBalanceCode function")
	}
	if !strings.Contains(code, "GetBalanceParams") {
		t.Error("expected GetBalanceParams data class")
	}
	if !strings.Contains(code, "/** Returns the Cadence code for the getBalance script. */") {
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

	if !strings.Contains(code, "data class TokenInfo") {
		t.Error("expected TokenInfo data class")
	}
	if !strings.Contains(code, "name: String") {
		t.Error("expected name field")
	}
	if !strings.Contains(code, "balance: BigDecimal") {
		t.Error("expected balance field (UFix64 -> BigDecimal)")
	}
	if !strings.Contains(code, "owner: FlowAddress?") {
		t.Error("expected owner optional field")
	}
	if !strings.Contains(code, "= null") {
		t.Error("expected null default for optional field")
	}
	if !strings.Contains(code, "// --- Structs ---") {
		t.Error("expected Structs section header")
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
	if !strings.Contains(code, "stakeCode") {
		t.Error("expected stakeCode function")
	}
	if !strings.Contains(code, "getStakeCode") {
		t.Error("expected getStakeCode function")
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

	if !strings.Contains(code, "fun initAccountCode(): String") {
		t.Error("expected initAccountCode function")
	}
	// Should NOT have params data class for no-parameter functions
	if strings.Contains(code, "InitAccountParams") {
		t.Error("should not generate params data class for function with no parameters")
	}
}

func TestNeedsBigInteger(t *testing.T) {
	report := analyzer.Report{
		Transactions: make(map[string]analyzer.AnalysisResult),
		Scripts: map[string]analyzer.AnalysisResult{
			"test.cdc": {
				Parameters: []analyzer.Parameter{
					{Name: "val", TypeStr: "Int128"},
				},
			},
		},
		Structs: make(map[string]analyzer.Struct),
	}
	gen := New(report)
	if !gen.needsBigInteger() {
		t.Error("expected needsBigInteger to return true for Int128")
	}

	report2 := analyzer.Report{
		Transactions: make(map[string]analyzer.AnalysisResult),
		Scripts: map[string]analyzer.AnalysisResult{
			"test.cdc": {
				Parameters: []analyzer.Parameter{
					{Name: "val", TypeStr: "String"},
				},
			},
		},
		Structs: make(map[string]analyzer.Struct),
	}
	gen2 := New(report2)
	if gen2.needsBigInteger() {
		t.Error("expected needsBigInteger to return false for String")
	}
}

func TestNeedsBigDecimal(t *testing.T) {
	report := analyzer.Report{
		Transactions: make(map[string]analyzer.AnalysisResult),
		Scripts:      make(map[string]analyzer.AnalysisResult),
		Structs: map[string]analyzer.Struct{
			"Test": {
				Fields: []analyzer.Field{
					{Name: "val", TypeStr: "UFix64"},
				},
			},
		},
	}
	gen := New(report)
	if !gen.needsBigDecimal() {
		t.Error("expected needsBigDecimal to return true for UFix64")
	}

	report2 := analyzer.Report{
		Transactions: make(map[string]analyzer.AnalysisResult),
		Scripts:      make(map[string]analyzer.AnalysisResult),
		Structs: map[string]analyzer.Struct{
			"Test": {
				Fields: []analyzer.Field{
					{Name: "val", TypeStr: "Int"},
				},
			},
		},
	}
	gen2 := New(report2)
	if gen2.needsBigDecimal() {
		t.Error("expected needsBigDecimal to return false for Int")
	}
}

func TestNeedsFlowAddress(t *testing.T) {
	report := analyzer.Report{
		Transactions: map[string]analyzer.AnalysisResult{
			"test.cdc": {
				Parameters: []analyzer.Parameter{
					{Name: "to", TypeStr: "Address"},
				},
			},
		},
		Scripts: make(map[string]analyzer.AnalysisResult),
		Structs: make(map[string]analyzer.Struct),
	}
	gen := New(report)
	if !gen.needsFlowAddress() {
		t.Error("expected needsFlowAddress to return true for Address")
	}

	report2 := analyzer.Report{
		Transactions: map[string]analyzer.AnalysisResult{
			"test.cdc": {
				Parameters: []analyzer.Parameter{
					{Name: "val", TypeStr: "String"},
				},
			},
		},
		Scripts: make(map[string]analyzer.AnalysisResult),
		Structs: make(map[string]analyzer.Struct),
	}
	gen2 := New(report2)
	if gen2.needsFlowAddress() {
		t.Error("expected needsFlowAddress to return false for String")
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

func TestGenerateWithBigIntegerImport(t *testing.T) {
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

	if !strings.Contains(code, "import java.math.BigInteger") {
		t.Error("expected java.math.BigInteger import when Int128 is used")
	}
}
