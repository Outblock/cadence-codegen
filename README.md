# Cadence Codegen

A tool for analyzing Cadence smart contracts and generating code. It extracts transaction and script information, including parameters, imports, and return types.

## Installation

### Using npm (Recommended)

```bash
# Global installation
npm install -g @outblock/cadence-codegen

# Or use in a project
npm install @outblock/cadence-codegen
npx cadence-codegen --help
```

### Using Homebrew

```bash
brew tap outblock/tap
brew install cadence-codegen
```

### Update Homebrew
```bash
brew update && brew upgrade cadence-codegen
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

### Generate TypeScript Code

Generate TypeScript code from Cadence files or JSON:

```bash
# Generate from Cadence files (outputs to cadence.generated.ts)
cadence-codegen typescript ./contracts

# Generate from Cadence files with custom output path
cadence-codegen typescript ./contracts output.ts

# Generate from previously analyzed JSON
cadence-codegen typescript analysis.json output.ts
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
  - TypeScript code with FCL integration
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

## Generated TypeScript Code

The generated TypeScript code includes:

- Type-safe functions for transactions and scripts
- FCL (Flow Client Library) integration
- Support for request and response interceptors
- Automatic type conversion from Cadence to TypeScript
- Support for async/await
- Struct definitions with proper TypeScript interfaces

Example usage of generated TypeScript code:

```typescript
import { CadenceService } from "./cadence.generated";

const service = new CadenceService();

// Add interceptors for request/response processing
service.useRequestInterceptor((config) => {
  console.log("Request config:", config);
  return config;
});

service.useResponseInterceptor((config, response) => {
  console.log("Response:", response);
  return { config, response };
});

// Execute a script
const result = await service.getAddr(flowAddress);

// Send a transaction
const txId = await service.createCoa(amount);
```

## NPM Integration

When installed via npm, the tool automatically downloads the appropriate binary for your platform (macOS, Linux, Windows) during installation. This provides a seamless experience for JavaScript/TypeScript developers who want to integrate Cadence code generation into their build processes.

### Use in package.json scripts

```json
{
  "scripts": {
    "codegen": "cadence-codegen typescript ./cadence output.ts",
    "build": "npm run codegen && tsc"
  }
}
```

### Use programmatically

```javascript
const { execSync } = require('child_process');

// Generate TypeScript code
execSync('cadence-codegen typescript ./contracts generated.ts');

// Or analyze contracts
const analysis = execSync('cadence-codegen analyze ./contracts', { encoding: 'utf8' });
console.log(JSON.parse(analysis));
```

## License

MIT License
