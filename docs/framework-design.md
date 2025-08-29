# Go WASM Framework Design

## Overview

This document outlines the design for transforming the current Go WASM project into a reusable framework that enables developers to build WebAssembly applications with clean, modular architecture. The framework abstracts away WASM boilerplate while providing a plugin-based system for building domain-specific applications.

## Goals

### Primary Goals
- **Reusability**: Create a framework that can be used for any Go WASM project
- **Modularity**: Enable clean separation between framework code and application logic
- **Extensibility**: Support plugin-based architecture for different use cases
- **Developer Experience**: Provide simple, intuitive APIs with minimal boilerplate
- **Performance**: Maintain the performance benefits of Go WASM

### Non-Goals
- Not trying to replace existing web frameworks
- Not attempting to abstract away all WASM concepts
- Not building a full-stack framework (frontend remains separate)

## Architecture

### Core Components

```
wasm-framework/
├── framework/                    # Core framework package
│   ├── core/
│   │   ├── application.go      # Main application interface
│   │   ├── handler.go          # Handler base and utilities
│   │   ├── registry.go         # Function registration system
│   │   └── runtime.go          # WASM runtime management
│   ├── types/
│   │   ├── context.go          # Request context
│   │   ├── response.go         # Standard response types
│   │   ├── middleware.go       # Middleware interfaces
│   │   └── config.go           # Configuration types
│   └── utils/
│       ├── jsbridge.go         # JS<->Go conversion utilities
│       ├── validation.go       # Input validation helpers
│       └── errors.go           # Error handling utilities
├── plugins/                     # Optional plugins
│   ├── logging/                 # Structured logging
│   ├── metrics/                 # Performance metrics
│   └── tracing/                 # Distributed tracing
├── examples/                    # Example implementations
│   ├── ecs-game/               # Entity Component System game
│   ├── data-processing/        # Data processing application
│   ├── crud-api/               # CRUD operations example
│   └── minimal/                # Minimal starter template
├── templates/                   # Project generators
│   └── new-project/
└── cmd/
    └── wasm-cli/               # CLI tool for project scaffolding
```

### Core Interfaces

#### Application Interface

```go
// Application represents a WASM application
type Application interface {
    // Metadata
    Name() string
    Version() string
    
    // Lifecycle hooks
    Initialize(config Config) error
    RegisterHandlers(registry *HandlerRegistry)
    OnReady() error
    Shutdown() error
    
    // Optional: Custom configuration
    Configure() Config
}
```

#### Handler Interface

```go
// Handler processes WASM function calls
type Handler interface {
    // Handle processes the function call
    Handle(ctx *Context, inputs []js.Value) Response
    
    // Metadata
    Name() string
    Description() string
    
    // Optional: Validation
    ValidateInputs(inputs []js.Value) error
}

// HandlerFunc allows simple functions to be used as handlers
type HandlerFunc func(ctx *Context, inputs []js.Value) Response
```

#### Context

```go
// Context provides request context and utilities
type Context struct {
    // Request information
    FunctionName string
    RequestID    string
    StartTime    time.Time
    
    // Application reference
    App Application
    
    // Utilities
    Logger  Logger
    Metrics MetricsCollector
    
    // User-defined values
    values map[string]interface{}
}
```

#### Response Types

```go
// Response represents a function response
type Response interface {
    ToJS() js.Value
    IsError() bool
}

// SuccessResponse for successful operations
type SuccessResponse struct {
    Data      interface{}
    Message   string
    Metadata  map[string]interface{}
}

// ErrorResponse for errors
type ErrorResponse struct {
    Error    string
    Code     string
    Details  map[string]interface{}
}
```

### Middleware System

```go
// Middleware modifies handler behavior
type Middleware interface {
    Execute(next HandlerFunc) HandlerFunc
}

// Common middleware
type LoggingMiddleware struct{}
type MetricsMiddleware struct{}
type RecoveryMiddleware struct{}
type ValidationMiddleware struct{}
```

## Usage Examples

### Minimal Application

```go
package main

import (
    "github.com/your-org/wasm-framework/framework/core"
)

type MinimalApp struct{}

func (app *MinimalApp) Name() string { return "Minimal App" }
func (app *MinimalApp) Version() string { return "1.0.0" }

func (app *MinimalApp) Initialize(config core.Config) error {
    // Setup code here
    return nil
}

func (app *MinimalApp) RegisterHandlers(registry *core.HandlerRegistry) {
    // Register a simple handler using HandlerFunc
    registry.RegisterFunc("greet", func(ctx *core.Context, inputs []js.Value) core.Response {
        name := "World"
        if len(inputs) > 0 {
            name = inputs[0].String()
        }
        
        return ctx.Success(map[string]interface{}{
            "message": fmt.Sprintf("Hello, %s!", name),
            "timestamp": time.Now().Unix(),
        })
    })
}

func main() {
    app := &MinimalApp{}
    core.RunApplication(app)
}
```

### ECS Game Engine Application

```go
package main

import (
    "github.com/your-org/wasm-framework/framework/core"
    "github.com/your-project/ecs"
)

type ECSGameApp struct {
    engine *ecs.Engine
    world  *ecs.World
}

func (app *ECSGameApp) Name() string { return "ECS Game Engine" }
func (app *ECSGameApp) Version() string { return "2.0.0" }

func (app *ECSGameApp) Initialize(config core.Config) error {
    app.engine = ecs.NewEngine()
    app.world = ecs.NewWorld()
    
    // Register systems
    app.engine.AddSystem(&ecs.MovementSystem{})
    app.engine.AddSystem(&ecs.RenderSystem{})
    app.engine.AddSystem(&ecs.CollisionSystem{})
    
    return nil
}

func (app *ECSGameApp) RegisterHandlers(registry *core.HandlerRegistry) {
    // Game tick processing
    registry.Register("processGameTick", &GameTickHandler{
        engine: app.engine,
        world:  app.world,
    })
    
    // Entity management
    registry.Register("createEntity", &CreateEntityHandler{world: app.world})
    registry.Register("deleteEntity", &DeleteEntityHandler{world: app.world})
    registry.Register("addComponent", &AddComponentHandler{world: app.world})
    
    // Player input
    registry.Register("processInput", &InputHandler{engine: app.engine})
    
    // State management
    registry.Register("saveGameState", &SaveStateHandler{world: app.world})
    registry.Register("loadGameState", &LoadStateHandler{world: app.world})
}

func main() {
    app := &ECSGameApp{}
    
    // Run with custom configuration
    core.RunApplicationWithConfig(app, core.Config{
        Debug:       true,
        MaxHandlers: 100,
        Middleware: []core.Middleware{
            &core.LoggingMiddleware{},
            &core.MetricsMiddleware{},
            &core.RecoveryMiddleware{},
        },
    })
}
```

### Custom Handler Implementation

```go
type GameTickHandler struct {
    engine *ecs.Engine
    world  *ecs.World
}

func (h *GameTickHandler) Name() string {
    return "processGameTick"
}

func (h *GameTickHandler) Description() string {
    return "Processes one game tick and returns updated state"
}

func (h *GameTickHandler) ValidateInputs(inputs []js.Value) error {
    if len(inputs) < 1 {
        return errors.New("deltaTime parameter required")
    }
    return nil
}

func (h *GameTickHandler) Handle(ctx *core.Context, inputs []js.Value) core.Response {
    deltaTime := inputs[0].Float()
    
    // Process game tick
    h.engine.Update(h.world, deltaTime)
    
    // Serialize current state
    gameState := h.world.Serialize()
    
    // Return response
    return ctx.Success(map[string]interface{}{
        "entities":  gameState.Entities,
        "timestamp": time.Now().UnixMilli(),
        "fps":       1.0 / deltaTime,
    })
}
```

## Framework Features

### 1. Automatic WASM Setup
The framework handles all WASM boilerplate:
- Function registration with JavaScript
- Keep-alive channel management
- Panic recovery and error handling
- Memory management optimizations

### 2. JavaScript Bridge Utilities
Built-in utilities for JS<->Go conversions:
```go
// Automatic type conversion
jsValue := utils.ToJS(goStruct)
goValue := utils.FromJS(jsValue, &targetStruct)

// Array and object helpers
arr := utils.JSArray([]int{1, 2, 3})
obj := utils.JSObject(map[string]interface{}{...})
```

### 3. Middleware System
Chainable middleware for cross-cutting concerns:
```go
registry.Use(
    middleware.Logger(),
    middleware.Recovery(),
    middleware.Metrics(),
    middleware.RateLimit(100),
)
```

### 4. Configuration Management
Flexible configuration system:
```go
type Config struct {
    Debug          bool
    MaxHandlers    int
    RequestTimeout time.Duration
    Middleware     []Middleware
    CustomValues   map[string]interface{}
}
```

### 5. Development Tools
CLI tool for project management:
```bash
# Create new project
wasm-cli new my-project --template minimal

# Add handler boilerplate
wasm-cli generate handler MyHandler

# Build and optimize
wasm-cli build --optimize

# Development server
wasm-cli serve --hot-reload
```

## Migration Path

### Phase 1: Core Framework (Week 1)
1. Extract core interfaces and types
2. Implement handler registration system
3. Create JavaScript bridge utilities
4. Build minimal working example

### Phase 2: Current Features as Plugins (Week 2)
1. Refactor data processing into example
2. Move API handlers to plugin architecture
3. Update build system for framework structure
4. Create comprehensive tests

### Phase 3: Advanced Features (Week 3-4)
1. Implement middleware system
2. Add configuration management
3. Create CLI tool
4. Build additional examples (ECS, CRUD, etc.)

### Phase 4: Documentation & Polish (Week 5)
1. Write comprehensive documentation
2. Create tutorial series
3. Build project templates
4. Performance optimization

## Benefits

### For Framework Users
- **Quick Start**: Get a WASM app running in minutes
- **Best Practices**: Built-in patterns for common tasks
- **Type Safety**: Leverage Go's type system
- **Performance**: Optimized WASM output
- **Flexibility**: Plugin architecture for any use case

### For Framework Maintainers
- **Clean Separation**: Framework vs application code
- **Testability**: Each component can be tested independently
- **Extensibility**: Easy to add new features
- **Community**: Enable ecosystem of plugins and examples

## Performance Considerations

### Optimization Strategies
1. **Minimize JS Bridge Crossings**: Batch operations when possible
2. **Efficient Serialization**: Use optimized JSON/binary formats
3. **Memory Pooling**: Reuse objects to reduce GC pressure
4. **Lazy Loading**: Load handlers on demand
5. **WASM Size**: Tree-shaking and dead code elimination

### Benchmarking
Framework includes built-in benchmarking:
```go
func BenchmarkHandler(b *testing.B) {
    ctx := core.NewTestContext()
    handler := &MyHandler{}
    
    for i := 0; i < b.N; i++ {
        handler.Handle(ctx, []js.Value{...})
    }
}
```

## Security Considerations

### Built-in Security Features
1. **Input Validation**: Automatic input sanitization
2. **Panic Recovery**: Graceful error handling
3. **Rate Limiting**: Prevent abuse
4. **Context Isolation**: Separate execution contexts
5. **Secure Defaults**: Safe configuration out of the box

## Testing Strategy

### Unit Testing
```go
func TestHandler(t *testing.T) {
    app := &MyApp{}
    ctx := core.NewTestContext(app)
    
    response := myHandler.Handle(ctx, []js.Value{
        core.JSValue("test input"),
    })
    
    assert.True(t, response.IsSuccess())
    assert.Equal(t, "expected", response.Data["result"])
}
```

### Integration Testing
```go
func TestApplication(t *testing.T) {
    app := &MyApp{}
    tester := core.NewApplicationTester(app)
    
    // Test initialization
    err := tester.Initialize()
    assert.NoError(t, err)
    
    // Test handler registration
    handlers := tester.GetRegisteredHandlers()
    assert.Contains(t, handlers, "myHandler")
    
    // Test handler execution
    result := tester.CallHandler("myHandler", "input")
    assert.Equal(t, "expected", result)
}
```

## Future Enhancements

### Planned Features
1. **Hot Module Reload**: Update handlers without restart
2. **Distributed Tracing**: OpenTelemetry integration
3. **GraphQL Support**: GraphQL handler generator
4. **WebSocket Support**: Bidirectional communication patterns
5. **Service Worker**: Offline-first capabilities
6. **WASM Threads**: Parallel processing support

### Community Features
1. **Plugin Registry**: Central repository for plugins
2. **Template Gallery**: Starter templates for common use cases
3. **Performance Dashboard**: Real-time metrics visualization
4. **Documentation Generator**: Auto-generate API docs from handlers

## Conclusion

This framework design provides a solid foundation for building Go WASM applications while maintaining flexibility and performance. The plugin-based architecture ensures that developers can use the framework for any type of application, from games to data processing to CRUD APIs.

The clean separation between framework and application code makes it easy to maintain and evolve both independently. By providing comprehensive examples and tools, we lower the barrier to entry for developers new to WebAssembly while still providing the power and flexibility needed for complex applications.