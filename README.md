# local-first

A complete example of building API-first applications using Go compiled to WebAssembly. This project demonstrates how to create powerful client-side applications with Go's type system and performance while maintaining free hosting through static site deployment.

## 🚀 Quick Start

### Prerequisites
- **Go 1.21+** - [Download Go](https://golang.org/dl/)
- **Node.js 18+** - [Download Node](https://nodejs.org/)
- **Git** - [Download Git](https://git-scm.com/)

### 1️⃣ Fastest Start (Recommended)

```bash
# Clone and setup
git clone https://github.com/mbarlow/local-first.git
cd local-first

# Install all dependencies (Go + Node)
make install

# Start development server with hot reload
make dev

# Open http://localhost:5173 in your browser
```

That's it! The app will automatically rebuild when you change Go or JavaScript files.

### 2️⃣ Alternative: Go Server Mode

If you prefer using the Go server instead of Vite:

```bash
# Build WASM and start Go server
make wasm
make serve

# Open http://localhost:8080 in your browser
```

### 3️⃣ Alternative: Docker Development

For a fully containerized environment:

```bash
# Build and run with Docker
make docker-build
make docker-dev

# Open http://localhost:5173 in your browser
```

### 4️⃣ Manual Setup (Advanced)

If you want to understand what's happening under the hood:

```bash
# Install Node dependencies
npm install

# Build the WASM module
GOOS=js GOARCH=wasm go build -o web/main.wasm cmd/wasm/main.go

# Copy WASM runtime support file
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/

# Option A: Run with Vite (hot reload)
npm run dev  # http://localhost:5173

# Option B: Run with Go server
go run cmd/server/main.go -dev  # http://localhost:8080
```

## 💡 What You'll See

Once running, you'll have access to a web app that demonstrates:
- Text processing and analysis
- Statistical calculations  
- Input validation (email, URL, JSON)
- ID generation (UUID, timestamps)
- All running locally in your browser via Go WASM!

## 🔍 Common Issues & Solutions

**Problem:** `make: command not found`
- **Solution:** Install make or run commands directly from package.json/Makefile

**Problem:** "Go WASM API not ready yet"
- **Solution:** Run `make wasm` first to build the WASM module, then start the server

**Problem:** "Function returned undefined"
- **Solution:** This was a known issue that's now fixed. If you see it, try `make clean && make wasm`

**Problem:** WASM file not loading/corrupted
- **Solution:** The Makefile now auto-downloads `wasm_exec.js` if missing or corrupted

**Problem:** Port already in use
- **Solution:** Change port in vite.config.js or use `-port` flag with Go server

**Problem:** Browser shows CORS errors
- **Solution:** Use the provided servers (Vite or Go) - they include proper WASM headers

## 📋 System Requirements

- **Go 1.21+** - For building WASM modules
- **Node.js 18+** - For Vite development server  
- **Modern Browser** - Chrome 88+, Firefox 89+, Safari 15+, Edge 88+
- **Docker** (optional) - For containerized development
- **Make** (optional) - For convenient command shortcuts

## 🏗️ Project Structure

```
/
├── Makefile              # Build automation and commands
├── Dockerfile           # Multi-stage Docker build
├── package.json         # Node.js dependencies and scripts
├── vite.config.js       # Vite configuration
├── README.md            # Project documentation
├── go.mod               # Go module definition
├── cmd/
│   ├── wasm/
│   │   └── main.go      # WASM entry point - exposes Go functions to JS
│   └── server/
│       ├── main.go      # Go HTTP server with embedded files
│       ├── embed.go     # Embedded file system (production)
│       └── no_embed.go  # File system mode (development)
├── internal/
│   ├── api/
│   │   └── handlers.go  # API endpoint handlers with JS integration
│   └── core/
│       └── logic.go     # Business logic and data processing
├── web/
│   ├── index.html       # Frontend demo application
│   ├── app.js          # JavaScript API client with IndexedDB integration
│   ├── main.wasm       # Compiled WASM binary (generated)
│   └── wasm_exec.js    # Go WASM runtime (auto-copied)
├── bin/                 # Compiled binaries (generated)
└── dist/               # Production build output (generated)
```

## 🔧 Development Commands

### Essential Commands
| Command | Description | Port |
|---------|-------------|------|
| `make install` | Install all dependencies (run this first!) | - |
| `make dev` | Start Vite dev server with hot reload | 5173 |
| `make serve` | Run Go server in development mode | 8080 |
| `make build` | Build everything for production | - |

### Build Commands
| Command | Description |
|---------|-------------|
| `make wasm` | Build WASM module (development) |
| `make wasm-prod` | Build optimized WASM (production) |
| `make server` | Build Go server binary |
| `make server-embed` | Build server with embedded static files |

### Docker Commands
| Command | Description | Port |
|---------|-------------|------|
| `make docker-build` | Build Docker image | - |
| `make docker-dev` | Run development container | 5173 |
| `make docker-prod` | Run production container | 8080 |

### Utility Commands
| Command | Description |
|---------|-------------|
| `make test` | Run Go unit tests |
| `make fmt` | Format Go code |
| `make lint` | Run linters |
| `make clean` | Remove all build artifacts |
| `make help` | Show available commands |

## 🎯 API Functions

The Go WASM module exposes these functions to JavaScript:

### Data Processing
- **`processData(text)`** - Analyzes text for word count, readability, frequency
- **`calculateStats(numbers)`** - Computes mean, median, std dev, quartiles
- **`validateInput(input, type)`** - Validates emails, URLs, phone numbers, JSON

### Utilities
- **`formatJSON(jsonString)`** - Pretty-prints and validates JSON
- **`generateID(type)`** - Creates UUIDs, short IDs, timestamps
- **`getVersion()`** - Returns API version and build information

## 💻 Usage Examples

### JavaScript Integration

```javascript
// Simple function call
const result = goAPI.processData("Hello, World!");
console.log(result.data.wordCount); // 2

// Statistics calculation
const stats = goAPI.calculateStats([1, 2, 3, 4, 5]);
console.log(stats.data.mean); // 3

// Input validation
const validation = goAPI.validateInput("user@example.com", "email");
console.log(validation.data.valid); // true

// ID generation
const uuid = goAPI.generateID("uuid");
console.log(uuid.data.id); // "550e8400-e29b-41d4-a716-446655440000"
```

### WebSocket Integration Pattern

Since Go WASM cannot directly create WebSocket connections, handle them in JavaScript and pass data to Go for processing:

```javascript
const ws = new WebSocket('wss://api.example.com');
ws.onmessage = (event) => {
    // Process WebSocket data with Go WASM
    const processed = goAPI.processData(event.data);
    handleProcessedData(processed);
};
```

### IndexedDB Integration

The demo includes IndexedDB integration for persistent client-side storage:

```javascript
// Results are automatically saved to IndexedDB
const localStorage = new LocalStorage();
await localStorage.saveResult(processedData);
const recentResults = await localStorage.getResults(10);
```

## 🚀 Production Deployment

### Build for Production
```bash
# Build optimized WASM and server with embedded files
make build

# This creates:
# - Optimized WASM binary (smaller size)
# - Go server with embedded static files
# - Production-ready assets in dist/
```

### Deploy as Static Site

Since everything runs in the browser, you can deploy to any static hosting:

**GitHub Pages:**
```bash
make build
# Commit the web/ directory
# Enable GitHub Pages pointing to web/
```

**Netlify / Vercel / Cloudflare Pages:**
- Build command: `make wasm-prod`
- Publish directory: `web`
- No server needed!

### Deploy with Go Server

For more control over headers and routing:

```bash
# Build production server
make server-embed

# Run the server
./bin/server -port 8080

# Or use Docker
make docker-prod
```

## 🔍 Development Tips

### Hot Reload with Vite
When using `make dev`, Vite will:
- ✅ Auto-reload on JavaScript/HTML/CSS changes
- ⚠️ Require manual WASM rebuild for Go changes

**For automatic Go rebuilds** (optional):
```bash
# Terminal 1: Watch and rebuild WASM automatically
find . -name "*.go" | entr make wasm

# Terminal 2: Run Vite dev server
make dev
```

**Quick workflow:**
```bash
# Make Go changes, then:
make wasm  # Rebuild WASM
# Refresh browser - changes will be visible
```

### Debugging
- **Go output:** Check browser console for `fmt.Println` output
- **JavaScript errors:** Appear in browser developer console
- **Go panics:** Show as JavaScript exceptions with stack traces
- **Network:** Check Network tab for WASM loading issues

### Performance Tips
- Keep heavy computation in Go (it's faster than JS)
- Batch operations to minimize JS ↔ Go boundary crossings
- Use `make wasm-prod` for 30-40% smaller WASM files
- Enable gzip/brotli compression on your web server

## 📦 Dependencies

### Go Dependencies
This project uses **zero external Go dependencies** - only Go standard library:
- `syscall/js` - JavaScript interop for WASM
- `encoding/json` - JSON processing  
- `regexp` - Pattern matching for validation
- `crypto/rand` - Secure ID generation
- `math` - Statistical calculations
- `embed` - Static file embedding for production builds

### Frontend Dependencies
- **Vite** - Fast development server and build tool
- **TailwindCSS** - Utility-first CSS (loaded from CDN)
- **No framework dependencies** - Vanilla JavaScript for maximum compatibility

## 🤝 API Design Philosophy

- **Clean Interfaces**: Functions accept simple types, return structured responses
- **Error Handling**: All functions return `{success: bool, data: any, message: string}`
- **Type Safety**: Go's type system ensures reliable data processing
- **Testability**: Core logic separated from WASM integration layer

## 📈 Performance Characteristics

- **Bundle Size**: ~2-8MB WASM binary (can be cached)
- **Load Time**: ~100-500ms initial WASM compilation
- **Execution**: Near-native performance for computational tasks
- **Memory**: Efficient Go garbage collector in WASM environment

## 🔗 Useful Links

- [Go WebAssembly Documentation](https://github.com/golang/go/wiki/WebAssembly)
- [WASM Browser Support](https://caniuse.com/wasm)
- [TinyGo](https://tinygo.org/) - Alternative Go compiler for smaller WASM binaries
- [WebAssembly Studio](https://webassembly.studio/) - Online WASM development

## 📄 License

MIT License - see LICENSE file for details

## 🙋‍♂️ Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make changes and test: `make test && make build`
4. Submit a pull request

---

**Built with ❤️ using Go and WebAssembly**
