# Go WASM API Project Design

## Overview

This project demonstrates building a Go application that compiles to WebAssembly (WASM) and runs entirely in the browser. The Go code exposes a clean API interface that JavaScript can call directly, eliminating the need for a backend server while maintaining the benefits of Go's type system and tooling.

## Architecture

```
Browser Environment
┌─────────────────────────────────────┐
│  JavaScript Frontend               │
│  ├── index.html                    │
│  ├── app.js (calls Go functions)   │
│  └── IndexedDB (local storage)     │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│  Go WASM Module                     │
│  ├── API Functions                 │
│  ├── Business Logic                │
│  └── Data Processing               │
└─────────────────────────────────────┘
```

## Key Benefits

- **No Backend Required**: Everything runs client-side
- **Go Development Experience**: Write in Go, deploy as web app
- **API-First Design**: Clean interfaces between Go and JavaScript
- **Free Hosting**: Deploy to GitHub Pages, Cloudflare Pages, etc.
- **Offline Capable**: Works without internet after initial load

## WebSocket Limitations

**Important**: Go WASM cannot directly create WebSocket connections. WebSockets must be managed by JavaScript, then data can be passed to Go WASM functions for processing. The pattern is:

```javascript
// JavaScript handles WebSocket
const ws = new WebSocket('wss://example.com');
ws.onmessage = (event) => {
    // Pass data to Go WASM for processing
    const result = goAPI.processWebSocketData(event.data);
    // Handle result
};
```

## Project Structure

```
go-wasm-api/
├── Makefile              # Build automation
├── README.md             # Project documentation
├── project-design.md     # This file
├── go.mod               # Go module definition
├── cmd/
│   └── wasm/
│       └── main.go      # WASM entry point
├── internal/
│   ├── api/
│   │   └── handlers.go  # API function handlers
│   └── core/
│       └── logic.go     # Business logic
├── web/
│   ├── index.html       # Frontend HTML
│   ├── app.js          # JavaScript API client
│   └── wasm_exec.js    # Go WASM runtime (copied from Go installation)
└── dist/
    └── main.wasm       # Generated WASM binary
```

## API Design Pattern

### Go Side (WASM)
```go
// Expose functions to JavaScript global scope
func main() {
    js.Global().Set("goAPI", map[string]interface{}{
        "processData":    js.FuncOf(processData),
        "validateInput":  js.FuncOf(validateInput),
        "calculateStats": js.FuncOf(calculateStats),
    })
    <-make(chan bool) // Keep WASM alive
}

func processData(this js.Value, inputs []js.Value) interface{} {
    // Convert JS input to Go types
    data := inputs[0].String()

    // Process with Go logic
    result := internal.ProcessBusinessData(data)

    // Return Go types (auto-converted to JS)
    return map[string]interface{}{
        "success": true,
        "data":    result,
    }
}
```

### JavaScript Side
```javascript
// Simple API calls to Go functions
const result = goAPI.processData("input data");
console.log(result.data);

// Async pattern for complex operations
function processAsync(data) {
    return new Promise((resolve) => {
        const result = goAPI.processData(data);
        resolve(result);
    });
}
```

## Data Flow Examples

### 1. Simple Data Processing
```
User Input → JavaScript → Go WASM → Processed Result → JavaScript → UI Update
```

### 2. WebSocket Integration
```
WebSocket Message → JavaScript Handler → Go WASM Processing → Result → JavaScript → UI/Storage
```

### 3. IndexedDB Integration
```
Go WASM Processing → JavaScript → IndexedDB Storage → Retrieval → Go WASM → Display
```

## Development Workflow

1. **Development**: Use `make dev` to build and serve locally
2. **Testing**: Go functions can be unit tested normally
3. **Building**: `make build` creates optimized WASM binary
4. **Deployment**: Deploy `web/` directory to static hosting

## Build Commands

```makefile
# Development server with hot reload
make dev

# Build optimized WASM
make build

# Clean build artifacts
make clean

# Run Go tests
make test
```

## Use Cases

This architecture works well for:
- Data processing applications
- Form validation and business logic
- Computational tools
- Offline-first applications
- APIs that don't require server-side state

Not suitable for:
- Applications requiring server-side WebSocket connections
- Server-side database operations
- Applications requiring server-side authentication
- Real-time multi-user applications

## Performance Considerations

- WASM loading is a one-time cost at page load
- Go WASM binary size is typically 2-10MB (can be optimized)
- Function calls between JS and Go have minimal overhead
- Complex data structures should stay in Go when possible
- Use IndexedDB for persistent client-side storage
