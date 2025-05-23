# Cadence Codegen

A tool for analyzing Cadence smart contracts and generating code. It extracts transaction and script information, including parameters, imports, and return types.

## Installation

### Using Homebrew

```bash
brew tap outblock/tap
brew install cadence-codegen
```

### Using Go

```bash
go install github.com/outblock/cadence-codegen@latest
```

## Usage

### Analyze Cadence Files

Analyze Cadence files and generate a JSON report:

```bash
# Analyze a directory of Cadence files
cadence-codegen analyze ./contracts

# Analyze with custom output path
cadence-codegen analyze ./contracts output.json

# Analyze without base64 encoding
cadence-codegen analyze ./contracts --base64=false
```

### Generate Swift Code

Generate Swift code from Cadence files or JSON:

```bash
# Generate from Cadence files (outputs to CadenceGen.swift)
cadence-codegen swift ./contracts

# Generate from Cadence files with custom output path
cadence-codegen swift ./contracts output.swift

# Generate from previously analyzed JSON
cadence-codegen swift analysis.json output.swift
```

## Features

- Analyzes Cadence files (.cdc)
- Extracts:
  - Transaction parameters and types
  - Script parameters and return types
  - Import statements
  - Struct definitions
- Generates:
  - Structured JSON output
  - Swift code with type-safe wrappers
- Supports folder-based tagging for better organization
- Base64 encoding of Cadence files (optional)

## JSON Output Format

```json
{
  "transactions": {
    "filename.cdc": {
      "fileName": "filename.cdc",
      "type": "transaction",
      "parameters": [
        {
          "name": "amount",
          "typeStr": "UFix64",
          "optional": false
        }
      ],
      "imports": [
        {
          "contract": "FungibleToken",
          "address": "0xFungibleToken"
        }
      ],
      "tag": "TokenTransfer"
    }
  },
  "scripts": {
    // Similar structure for scripts
  },
  "structs": {
    // Struct definitions
  }
}
```

## Generated Swift Code

The generated Swift code includes:
- Type-safe enums for transactions and scripts
  - Separate enums for each folder (e.g., `CadenceGen.EVM` for files in the EVM folder)
  - Main `CadenceGen` enum for files in the root directory
- Struct definitions with proper Swift types
- Automatic Flow SDK integration
- Support for async/await
- Error handling

Example usage of generated Swift code:

```swift
// Execute a script from the EVM folder
let result: String? = try await flow.query(CadenceGen.EVM.getAddr(flowAddress: address))

// Execute a script from the root directory
let result: String? = try await flow.query(CadenceGen.getAddr(flowAddress: address))

// Send a transaction
let txId = try await flow.sendTx(
    CadenceGen.EVM.createCoa(amount: amount),
    singers: [signer]
) { 
    // Transaction build options
}
```

## License

MIT License 