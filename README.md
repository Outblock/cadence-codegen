# Cadence Codegen

A tool for analyzing Cadence smart contracts and generating code. It extracts transaction and script information, including parameters, imports, and return types.

## Features

- Analyzes Cadence files (.cdc)
- Extracts transaction and script information
- Parses import statements
- Generates structured JSON output

## Installation

```bash
go install github.com/outblock/cadence-codegen@latest
```

## Usage

```go
analyzer := analyzer.New()

// Analyze a single file
result, err := analyzer.AnalyzeFile("path/to/contract.cdc")

// Analyze a directory
err := analyzer.AnalyzeDirectory("path/to/contracts")

// Get the complete report
report := analyzer.GetReport()
```

## Output Format

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
      ]
    }
  },
  "scripts": {
    // Similar structure for scripts
  }
}
```

## License

MIT License 