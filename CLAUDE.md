# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go WebAssembly (WASM) project that demonstrates building API-first applications that run entirely in the browser. The Go code compiles to WASM and exposes clean API interfaces to JavaScript, eliminating the need for a backend server while maintaining Go's type system and performance benefits.

## Architecture

The project follows a layered architecture with clear separation between WASM integration and business logic:

- **WASM Entry Point** (`cmd/wasm/main.go`) - Exposes Go functions to JavaScript global scope
- **API Handlers** (`internal/api/handlers.go`) - WASM-compatible function wrappers with standardized response format
- **Core Logic** (`internal/core/logic.go`) - Pure Go business logic with zero WASM dependencies
- **Frontend** (`web/`) - Vanilla JavaScript client that calls Go WASM functions
- **CLI Tools** (`cmd/cli/`) - Interactive terminal dashboard built with Bubble Tea
- **Go Server** (`cmd/server/`) - Optional HTTP server with embedded files support

## Essential Development Commands

### Setup and Development
```bash
make install        # Install Go modules and npm dependencies
make dev           # Start Vite dev server with hot reload on :5173
make serve         # Run Go HTTP server in dev mode on :8080
./bin/local dash   # Start interactive CLI dashboard (after make cli)
```

### Building
```bash
make wasm          # Build WASM module for development
make wasm-prod     # Build optimized WASM for production (-s -w flags)
make build         # Full production build (WASM + server + frontend)
make cli           # Build interactive CLI dashboard
```

### Testing and Quality
```bash
make test          # Run Go unit tests
make fmt           # Format Go code
make lint          # Run linters (golangci-lint or go vet fallback)
```

## Adding New WASM Methods

When adding new functionality exposed to JavaScript:

1. **Core Logic** (`internal/core/logic.go`):
   ```go
   func (dp *DataProcessor) NewFunction(input string) (map[string]interface{}, error) {
       // Business logic here
       return map[string]interface{}{"result": processedData}, nil
   }
   ```

2. **API Handler** (`internal/api/handlers.go`):
   ```go
   func (h *Handler) NewFunction(this js.Value, inputs []js.Value) interface{} {
       result, err := h.processor.NewFunction(inputs[0].String())
       if err != nil {
           return h.errorResponse(err.Error())
       }
       return h.successResponse(result, "Success message")
   }
   ```

3. **WASM Registration** (`cmd/wasm/main.go`):
   ```go
   goAPI.Set("newFunction", js.FuncOf(apiHandler.NewFunction))
   ```

4. **JavaScript Client** (`web/app.js`):
   ```javascript
   function newFunction() {
       const result = client.callAPI("newFunction", inputData);
       client.displayResult("resultsElementId", result);
   }
   ```

Always rebuild WASM after Go changes: `make wasm`

## Key Patterns

### Response Format
All WASM functions return standardized responses:
```go
// Success
h.successResponse(data, "message")

// Error  
h.errorResponse("error message")
```

### WebSocket Limitation
Go WASM cannot create WebSocket connections. Pattern is JavaScript handles WebSocket, passes data to Go for processing:
```javascript
ws.onmessage = (event) => {
    const processed = goAPI.processData(event.data);
    // Handle result
};
```

### Data Flow
- Keep heavy computation in Go (better performance)
- Minimize JS ↔ Go boundary crossings
- Use IndexedDB for persistent client-side storage
- Business logic stays in `internal/core/` (testable, no WASM dependencies)

## CLI Dashboard

Interactive terminal dashboard with:
- Real-time server management (start/stop/restart)
- Request monitoring with color coding
- Performance statistics
- Keyboard shortcuts (s: start, x: stop, r: restart)

Build with `make cli`, run with `./bin/local dashboard`

## File Structure Notes

- `cmd/wasm/main.go` - WASM entry point, registers functions globally
- `cmd/server/main.go` - HTTP server with conditional embedded files
- `cmd/server/embed.go` - Embedded files for production builds
- `cmd/server/no_embed.go` - Filesystem serving for development
- `internal/api/handlers.go` - WASM-compatible API layer
- `internal/core/logic.go` - Pure Go business logic
- `internal/cli/` - CLI dashboard components (Bubble Tea)
- `internal/monitoring/` - HTTP middleware for request logging
- `web/` - Frontend files (HTML, JS, CSS)
- `vite.config.js` - Vite config with WASM headers and CORP/COOP

## Dependencies

- **Go**: Standard library only for core logic
- **WASM**: `syscall/js` for JavaScript interop
- **CLI**: Bubble Tea, Cobra, Viper for terminal interface
- **Frontend**: Vite for development, no JavaScript frameworks
- **Docker**: Multi-stage builds supported

## Testing Strategy

- Unit test core business logic in `internal/core/`
- Use `make test` for Go tests
- Frontend testing via browser developer tools
- CLI dashboard has built-in request monitoring