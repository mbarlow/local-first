# ECS Game Engine in Go WASM

## Overview

This document outlines how to build an Entity Component System (ECS) game engine in Go WebAssembly with JavaScript clients. This approach leverages Go's type system and performance while using JavaScript for I/O operations and rendering.

## Architecture

### High-Level Structure

```
┌─────────────────────────────────────┐
│  JavaScript Client                  │
│  ├── WebSocket Connection           │
│  ├── Input Handling                 │
│  ├── Rendering (Canvas/WebGL)       │
│  └── Network Reconciliation         │
└─────────────────────────────────────┘
              │
              ▼ (Function Calls)
┌─────────────────────────────────────┐
│  Go WASM Game Engine                │
│  ├── Entity Component System        │
│  ├── Game Systems (Movement, AI)    │
│  ├── Game State Management          │
│  └── Physics/Logic Processing       │
└─────────────────────────────────────┘
```

### WebSocket Limitation & Solution

**Constraint**: Go WASM cannot create WebSocket connections directly.

**Solution**: JavaScript handles all networking, passes data to Go WASM for processing.

```javascript
// JavaScript manages networking
const ws = new WebSocket('ws://game-server.com/multiplayer');

ws.onmessage = (event) => {
    const serverUpdate = JSON.parse(event.data);
    // Pass to Go WASM for game logic processing
    const result = goAPI.processServerUpdate(serverUpdate);
    renderGame(result);
};

// Send player actions to server
function sendAction(action) {
    ws.send(JSON.stringify(action));
    // Immediate local prediction
    const localResult = goAPI.processLocalAction(action);
    renderGame(localResult);
}
```

## ECS Implementation in Go

### Core ECS Components

```go
// internal/core/ecs.go
package core

import (
    "fmt"
    "sync"
)

// EntityID represents a unique entity identifier
type EntityID uint32

// Component marker interface
type Component interface {
    ComponentType() string
}

// Transform component
type Transform struct {
    X, Y, Z     float64
    Rotation    float64
    Scale       float64
}

func (t Transform) ComponentType() string { return "Transform" }

// Velocity component
type Velocity struct {
    X, Y, Z float64
}

func (v Velocity) ComponentType() string { return "Velocity" }

// Health component
type Health struct {
    Current, Max int
}

func (h Health) ComponentType() string { return "Health" }

// Sprite component
type Sprite struct {
    TextureID string
    Width     int
    Height    int
    Layer     int
}

func (s Sprite) ComponentType() string { return "Sprite" }

// World manages all entities and components
type World struct {
    mu            sync.RWMutex
    nextEntityID  EntityID
    entities      map[EntityID]bool
    
    // Component storage
    transforms    map[EntityID]*Transform
    velocities    map[EntityID]*Velocity
    healths       map[EntityID]*Health
    sprites       map[EntityID]*Sprite
    
    // Systems
    systems       []System
}

// System interface
type System interface {
    Update(world *World, deltaTime float64)
    Name() string
}

// NewWorld creates a new game world
func NewWorld() *World {
    return &World{
        nextEntityID: 1,
        entities:     make(map[EntityID]bool),
        transforms:   make(map[EntityID]*Transform),
        velocities:   make(map[EntityID]*Velocity),
        healths:      make(map[EntityID]*Health),
        sprites:      make(map[EntityID]*Sprite),
        systems:      make([]System, 0),
    }
}

// CreateEntity creates a new entity and returns its ID
func (w *World) CreateEntity() EntityID {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    entityID := w.nextEntityID
    w.nextEntityID++
    w.entities[entityID] = true
    
    return entityID
}

// DeleteEntity removes an entity and all its components
func (w *World) DeleteEntity(entityID EntityID) {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    delete(w.entities, entityID)
    delete(w.transforms, entityID)
    delete(w.velocities, entityID)
    delete(w.healths, entityID)
    delete(w.sprites, entityID)
}

// AddComponent adds a component to an entity
func (w *World) AddComponent(entityID EntityID, component Component) {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    switch c := component.(type) {
    case *Transform:
        w.transforms[entityID] = c
    case *Velocity:
        w.velocities[entityID] = c
    case *Health:
        w.healths[entityID] = c
    case *Sprite:
        w.sprites[entityID] = c
    }
}

// GetComponent retrieves a component from an entity
func (w *World) GetComponent(entityID EntityID, componentType string) Component {
    w.mu.RLock()
    defer w.mu.RUnlock()
    
    switch componentType {
    case "Transform":
        return w.transforms[entityID]
    case "Velocity":
        return w.velocities[entityID]
    case "Health":
        return w.healths[entityID]
    case "Sprite":
        return w.sprites[entityID]
    }
    return nil
}

// AddSystem adds a system to the world
func (w *World) AddSystem(system System) {
    w.systems = append(w.systems, system)
}

// Update runs all systems
func (w *World) Update(deltaTime float64) {
    for _, system := range w.systems {
        system.Update(w, deltaTime)
    }
}
```

### Game Systems

```go
// internal/core/systems.go
package core

// MovementSystem handles entity movement
type MovementSystem struct{}

func (ms *MovementSystem) Name() string {
    return "Movement"
}

func (ms *MovementSystem) Update(world *World, deltaTime float64) {
    world.mu.Lock()
    defer world.mu.Unlock()
    
    // Process all entities with both Transform and Velocity components
    for entityID := range world.entities {
        transform, hasTransform := world.transforms[entityID]
        velocity, hasVelocity := world.velocities[entityID]
        
        if hasTransform && hasVelocity {
            transform.X += velocity.X * deltaTime
            transform.Y += velocity.Y * deltaTime
            transform.Z += velocity.Z * deltaTime
        }
    }
}

// RenderSystem prepares render data (doesn't actually render)
type RenderSystem struct{}

func (rs *RenderSystem) Name() string {
    return "Render"
}

func (rs *RenderSystem) Update(world *World, deltaTime float64) {
    // This system prepares render data for JavaScript
    // Actual rendering happens in JavaScript
}

// CollisionSystem handles collision detection
type CollisionSystem struct{}

func (cs *CollisionSystem) Name() string {
    return "Collision"
}

func (cs *CollisionSystem) Update(world *World, deltaTime float64) {
    world.mu.RLock()
    defer world.mu.RUnlock()
    
    // Simple AABB collision detection
    entities := make([]EntityID, 0, len(world.entities))
    for entityID := range world.entities {
        if _, hasTransform := world.transforms[entityID]; hasTransform {
            entities = append(entities, entityID)
        }
    }
    
    // Check collisions between entities
    for i := 0; i < len(entities); i++ {
        for j := i + 1; j < len(entities); j++ {
            entity1, entity2 := entities[i], entities[j]
            if cs.checkCollision(world, entity1, entity2) {
                cs.handleCollision(world, entity1, entity2)
            }
        }
    }
}

func (cs *CollisionSystem) checkCollision(world *World, e1, e2 EntityID) bool {
    t1 := world.transforms[e1]
    t2 := world.transforms[e2]
    s1 := world.sprites[e1]
    s2 := world.sprites[e2]
    
    if t1 == nil || t2 == nil || s1 == nil || s2 == nil {
        return false
    }
    
    // Simple AABB collision
    return t1.X < t2.X + float64(s2.Width) &&
           t1.X + float64(s1.Width) > t2.X &&
           t1.Y < t2.Y + float64(s2.Height) &&
           t1.Y + float64(s1.Height) > t2.Y
}

func (cs *CollisionSystem) handleCollision(world *World, e1, e2 EntityID) {
    // Handle collision response
    // Could modify velocity, trigger events, etc.
}
```

### Game Engine Integration

```go
// internal/core/game_engine.go
package core

import (
    "encoding/json"
    "time"
)

// GameEngine manages the overall game state and loop
type GameEngine struct {
    world        *World
    lastTick     time.Time
    isRunning    bool
    tickRate     float64 // ticks per second
}

// NewGameEngine creates a new game engine instance
func NewGameEngine() *GameEngine {
    engine := &GameEngine{
        world:     NewWorld(),
        lastTick:  time.Now(),
        tickRate:  60.0, // 60 FPS
    }
    
    // Add default systems
    engine.world.AddSystem(&MovementSystem{})
    engine.world.AddSystem(&CollisionSystem{})
    engine.world.AddSystem(&RenderSystem{})
    
    return engine
}

// ProcessTick updates the game world
func (ge *GameEngine) ProcessTick() map[string]interface{} {
    now := time.Now()
    deltaTime := now.Sub(ge.lastTick).Seconds()
    ge.lastTick = now
    
    // Update all systems
    ge.world.Update(deltaTime)
    
    // Return serialized game state for JavaScript
    return ge.SerializeGameState()
}

// ProcessPlayerInput handles player input
func (ge *GameEngine) ProcessPlayerInput(input map[string]interface{}) error {
    // Parse input and apply to entities
    if playerID, ok := input["playerID"].(float64); ok {
        entityID := EntityID(playerID)
        
        if velocity := ge.world.velocities[entityID]; velocity != nil {
            if dx, ok := input["dx"].(float64); ok {
                velocity.X = dx
            }
            if dy, ok := input["dy"].(float64); ok {
                velocity.Y = dy
            }
        }
    }
    
    return nil
}

// SerializeGameState returns the current game state for JavaScript
func (ge *GameEngine) SerializeGameState() map[string]interface{} {
    ge.world.mu.RLock()
    defer ge.world.mu.RUnlock()
    
    entities := make([]map[string]interface{}, 0)
    
    for entityID := range ge.world.entities {
        entity := map[string]interface{}{
            "id": entityID,
        }
        
        if transform := ge.world.transforms[entityID]; transform != nil {
            entity["transform"] = map[string]interface{}{
                "x":        transform.X,
                "y":        transform.Y,
                "z":        transform.Z,
                "rotation": transform.Rotation,
                "scale":    transform.Scale,
            }
        }
        
        if sprite := ge.world.sprites[entityID]; sprite != nil {
            entity["sprite"] = map[string]interface{}{
                "textureId": sprite.TextureID,
                "width":     sprite.Width,
                "height":    sprite.Height,
                "layer":     sprite.Layer,
            }
        }
        
        if health := ge.world.healths[entityID]; health != nil {
            entity["health"] = map[string]interface{}{
                "current": health.Current,
                "max":     health.Max,
            }
        }
        
        entities = append(entities, entity)
    }
    
    return map[string]interface{}{
        "entities": entities,
        "timestamp": time.Now().UnixMilli(),
    }
}

// CreatePlayer creates a player entity
func (ge *GameEngine) CreatePlayer(x, y float64) EntityID {
    entityID := ge.world.CreateEntity()
    
    ge.world.AddComponent(entityID, &Transform{
        X: x, Y: y, Z: 0,
        Rotation: 0, Scale: 1,
    })
    
    ge.world.AddComponent(entityID, &Velocity{
        X: 0, Y: 0, Z: 0,
    })
    
    ge.world.AddComponent(entityID, &Health{
        Current: 100, Max: 100,
    })
    
    ge.world.AddComponent(entityID, &Sprite{
        TextureID: "player",
        Width:     32,
        Height:    32,
        Layer:     1,
    })
    
    return entityID
}
```

## WASM API Integration

### API Handlers

```go
// internal/api/game_handlers.go
package api

import (
    "encoding/json"
    "fmt"
    "syscall/js"
    
    "github.com/mbarlow/local-first/internal/core"
)

// GameHandler contains game-specific API endpoints
type GameHandler struct {
    engine *core.GameEngine
}

// NewGameHandler creates a new game handler instance
func NewGameHandler() *GameHandler {
    return &GameHandler{
        engine: core.NewGameEngine(),
    }
}

// ProcessGameTick handles game tick updates from JavaScript
func (gh *GameHandler) ProcessGameTick(this js.Value, inputs []js.Value) interface{} {
    gameState := gh.engine.ProcessTick()
    
    return toJSValue(map[string]interface{}{
        "success":   true,
        "data":      gameState,
        "message":   "Game tick processed",
        "timestamp": js.Global().Get("Date").New().Call("getTime").Int(),
    })
}

// ProcessPlayerInput handles player input from JavaScript
func (gh *GameHandler) ProcessPlayerInput(this js.Value, inputs []js.Value) interface{} {
    if len(inputs) == 0 {
        return gh.errorResponse("No input provided")
    }
    
    // Parse JavaScript input object
    inputJS := inputs[0]
    input := make(map[string]interface{})
    
    // Convert JS object to Go map
    keys := js.Global().Get("Object").Call("keys", inputJS)
    length := keys.Get("length").Int()
    
    for i := 0; i < length; i++ {
        key := keys.Index(i).String()
        value := inputJS.Get(key)
        
        switch value.Type() {
        case js.TypeNumber:
            input[key] = value.Float()
        case js.TypeString:
            input[key] = value.String()
        case js.TypeBoolean:
            input[key] = value.Bool()
        }
    }
    
    err := gh.engine.ProcessPlayerInput(input)
    if err != nil {
        return gh.errorResponse(err.Error())
    }
    
    return gh.successResponse(nil, "Input processed")
}

// CreatePlayer creates a new player entity
func (gh *GameHandler) CreatePlayer(this js.Value, inputs []js.Value) interface{} {
    x := 0.0
    y := 0.0
    
    if len(inputs) >= 2 {
        x = inputs[0].Float()
        y = inputs[1].Float()
    }
    
    playerID := gh.engine.CreatePlayer(x, y)
    
    return gh.successResponse(map[string]interface{}{
        "playerID": playerID,
    }, "Player created")
}

// ProcessServerUpdate handles server reconciliation
func (gh *GameHandler) ProcessServerUpdate(this js.Value, inputs []js.Value) interface{} {
    if len(inputs) == 0 {
        return gh.errorResponse("No server update provided")
    }
    
    // Parse server update and reconcile with local state
    // This is where client-side prediction reconciliation happens
    
    return gh.successResponse(nil, "Server update processed")
}

// Helper methods
func (gh *GameHandler) successResponse(data interface{}, message string) js.Value {
    response := map[string]interface{}{
        "success":   true,
        "data":      data,
        "message":   message,
        "timestamp": js.Global().Get("Date").New().Call("getTime").Int(),
    }
    return toJSValue(response)
}

func (gh *GameHandler) errorResponse(message string) js.Value {
    response := map[string]interface{}{
        "success":   false,
        "error":     message,
        "timestamp": js.Global().Get("Date").New().Call("getTime").Int(),
    }
    return toJSValue(response)
}
```

### WASM Entry Point Integration

```go
// cmd/wasm/main.go - Add to existing main function
func main() {
    fmt.Println("Go WASM API loaded successfully!")

    // Create API handler instance
    apiHandler := api.NewHandler()
    gameHandler := api.NewGameHandler() // Add this

    // Create a JavaScript object to hold our API functions
    goAPI := js.Global().Get("Object").New()
    
    // Existing functions...
    goAPI.Set("processData", js.FuncOf(apiHandler.ProcessData))
    goAPI.Set("validateInput", js.FuncOf(apiHandler.ValidateInput))
    goAPI.Set("calculateStats", js.FuncOf(apiHandler.CalculateStats))
    goAPI.Set("formatJSON", js.FuncOf(apiHandler.FormatJSON))
    goAPI.Set("generateID", js.FuncOf(apiHandler.GenerateID))
    goAPI.Set("getVersion", js.FuncOf(apiHandler.GetVersion))
    
    // Game functions
    goAPI.Set("processGameTick", js.FuncOf(gameHandler.ProcessGameTick))
    goAPI.Set("processPlayerInput", js.FuncOf(gameHandler.ProcessPlayerInput))
    goAPI.Set("createPlayer", js.FuncOf(gameHandler.CreatePlayer))
    goAPI.Set("processServerUpdate", js.FuncOf(gameHandler.ProcessServerUpdate))
    
    // Set the goAPI object on the global window
    js.Global().Set("goAPI", goAPI)

    fmt.Println("Game API functions registered")
    fmt.Println("Available game functions: processGameTick, processPlayerInput, createPlayer, processServerUpdate")

    // Keep the Go program alive
    <-make(chan bool)
}
```

## JavaScript Client Implementation

### Game Client Class

```javascript
// web/game.js
class GameClient {
    constructor(canvasId) {
        this.canvas = document.getElementById(canvasId);
        this.ctx = this.canvas.getContext('2d');
        this.lastTime = 0;
        this.gameState = {};
        this.isRunning = false;
        this.keys = {};
        this.playerId = null;
        this.ws = null;
        
        // Networking
        this.serverUrl = 'ws://localhost:8080/ws';
        this.connected = false;
        
        this.initInput();
        this.initWebSocket();
    }
    
    initWebSocket() {
        try {
            this.ws = new WebSocket(this.serverUrl);
            
            this.ws.onopen = () => {
                console.log('Connected to game server');
                this.connected = true;
                
                // Request to join game
                this.sendMessage({
                    type: 'join_game',
                    timestamp: Date.now()
                });
            };
            
            this.ws.onmessage = (event) => {
                const message = JSON.parse(event.data);
                this.handleServerMessage(message);
            };
            
            this.ws.onclose = () => {
                console.log('Disconnected from game server');
                this.connected = false;
                
                // Attempt to reconnect after 3 seconds
                setTimeout(() => this.initWebSocket(), 3000);
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };
        } catch (error) {
            console.error('Failed to connect to game server:', error);
            // Continue with local-only mode
        }
    }
    
    sendMessage(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        }
    }
    
    handleServerMessage(message) {
        switch (message.type) {
            case 'game_state':
                // Server authoritative update
                const result = goAPI.processServerUpdate(message.data);
                if (result.success) {
                    // Server reconciliation handled in Go WASM
                }
                break;
                
            case 'player_joined':
                console.log('Player joined:', message.playerId);
                if (message.playerId === this.playerId) {
                    console.log('You joined the game');
                }
                break;
                
            case 'player_assigned':
                this.playerId = message.playerId;
                console.log('Assigned player ID:', this.playerId);
                break;
        }
    }
    
    initInput() {
        document.addEventListener('keydown', (e) => {
            this.keys[e.code] = true;
            this.handleInput();
        });
        
        document.addEventListener('keyup', (e) => {
            this.keys[e.code] = false;
            this.handleInput();
        });
    }
    
    handleInput() {
        const input = {
            playerID: this.playerId || 1, // Default to 1 for local play
            dx: 0,
            dy: 0
        };
        
        // Calculate movement based on keys
        if (this.keys['ArrowLeft'] || this.keys['KeyA']) input.dx = -100;
        if (this.keys['ArrowRight'] || this.keys['KeyD']) input.dx = 100;
        if (this.keys['ArrowUp'] || this.keys['KeyW']) input.dy = -100;
        if (this.keys['ArrowDown'] || this.keys['KeyS']) input.dy = 100;
        
        // Process input locally (client-side prediction)
        const result = goAPI.processPlayerInput(input);
        
        // Send to server if connected
        if (this.connected) {
            this.sendMessage({
                type: 'player_input',
                input: input,
                timestamp: Date.now()
            });
        }
    }
    
    start() {
        // Create player if not exists
        if (!this.playerId) {
            const result = goAPI.createPlayer(100, 100);
            if (result.success) {
                this.playerId = result.data.playerID;
                console.log('Created local player:', this.playerId);
            }
        }
        
        this.isRunning = true;
        this.gameLoop();
    }
    
    stop() {
        this.isRunning = false;
    }
    
    gameLoop(currentTime = 0) {
        if (!this.isRunning) return;
        
        const deltaTime = (currentTime - this.lastTime) / 1000;
        this.lastTime = currentTime;
        
        // Process game tick in Go WASM
        const result = goAPI.processGameTick();
        
        if (result.success) {
            this.gameState = result.data;
            this.render();
        }
        
        requestAnimationFrame((time) => this.gameLoop(time));
    }
    
    render() {
        // Clear canvas
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        
        if (!this.gameState.entities) return;
        
        // Sort entities by layer for proper rendering order
        const entities = [...this.gameState.entities].sort((a, b) => {
            const layerA = a.sprite ? a.sprite.layer : 0;
            const layerB = b.sprite ? b.sprite.layer : 0;
            return layerA - layerB;
        });
        
        // Render entities
        entities.forEach(entity => {
            if (entity.transform && entity.sprite) {
                this.renderEntity(entity);
            }
        });
        
        // Render UI
        this.renderUI();
    }
    
    renderEntity(entity) {
        const { transform, sprite } = entity;
        
        this.ctx.save();
        this.ctx.translate(transform.x + sprite.width/2, transform.y + sprite.height/2);
        this.ctx.rotate(transform.rotation);
        this.ctx.scale(transform.scale, transform.scale);
        
        // Simple colored rectangle for now
        // In a real game, you'd load and draw textures
        switch (sprite.textureId) {
            case 'player':
                this.ctx.fillStyle = '#4CAF50';
                break;
            case 'enemy':
                this.ctx.fillStyle = '#F44336';
                break;
            default:
                this.ctx.fillStyle = '#2196F3';
        }
        
        this.ctx.fillRect(-sprite.width/2, -sprite.height/2, sprite.width, sprite.height);
        
        // Health bar if entity has health
        if (entity.health) {
            const healthPercent = entity.health.current / entity.health.max;
            const barWidth = sprite.width;
            const barHeight = 4;
            
            this.ctx.fillStyle = '#333';
            this.ctx.fillRect(-barWidth/2, -sprite.height/2 - 10, barWidth, barHeight);
            
            this.ctx.fillStyle = healthPercent > 0.5 ? '#4CAF50' : 
                                 healthPercent > 0.25 ? '#FF9800' : '#F44336';
            this.ctx.fillRect(-barWidth/2, -sprite.height/2 - 10, 
                             barWidth * healthPercent, barHeight);
        }
        
        this.ctx.restore();
    }
    
    renderUI() {
        // Render game UI (score, health, etc.)
        this.ctx.font = '16px Arial';
        this.ctx.fillStyle = '#000';
        this.ctx.fillText(`Player ID: ${this.playerId || 'Local'}`, 10, 25);
        this.ctx.fillText(`Connected: ${this.connected ? 'Yes' : 'No'}`, 10, 45);
        this.ctx.fillText(`Entities: ${this.gameState.entities ? this.gameState.entities.length : 0}`, 10, 65);
        
        // Instructions
        this.ctx.font = '12px Arial';
        this.ctx.fillText('Use WASD or Arrow keys to move', 10, this.canvas.height - 20);
    }
}

// Initialize game when WASM is ready
let gameClient;

// Wait for Go WASM to load
function initGame() {
    if (typeof goAPI !== 'undefined') {
        gameClient = new GameClient('gameCanvas');
        gameClient.start();
        console.log('Game initialized');
    } else {
        setTimeout(initGame, 100);
    }
}

// Auto-start when page loads
document.addEventListener('DOMContentLoaded', initGame);
```

### HTML Game Interface

```html
<!-- web/game.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go WASM ECS Game</title>
    <style>
        body {
            margin: 0;
            padding: 20px;
            font-family: Arial, sans-serif;
            background: #f0f0f0;
        }
        
        .game-container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        
        canvas {
            border: 2px solid #333;
            background: #fff;
            display: block;
            margin: 0 auto;
        }
        
        .controls {
            text-align: center;
            margin-top: 20px;
        }
        
        button {
            padding: 10px 20px;
            margin: 5px;
            font-size: 16px;
            border: none;
            border-radius: 4px;
            background: #4CAF50;
            color: white;
            cursor: pointer;
        }
        
        button:hover {
            background: #45a049;
        }
        
        button:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        
        .status {
            text-align: center;
            margin: 10px 0;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="game-container">
        <h1>Go WASM ECS Game Engine</h1>
        
        <div class="status">
            <div id="statusIndicator" class="status-dot loading"></div>
            <span id="statusText">Loading Go WASM API...</span>
        </div>
        
        <canvas id="gameCanvas" width="800" height="600"></canvas>
        
        <div class="controls">
            <button onclick="gameClient && gameClient.start()">Start Game</button>
            <button onclick="gameClient && gameClient.stop()">Stop Game</button>
            <button onclick="location.reload()">Restart</button>
        </div>
        
        <div style="margin-top: 20px; font-size: 12px; color: #666;">
            <p><strong>Controls:</strong></p>
            <ul>
                <li>WASD or Arrow keys: Move player</li>
                <li>Game runs at 60 FPS with ECS architecture</li>
                <li>Multiplayer support via WebSocket (if server available)</li>
            </ul>
        </div>
    </div>

    <script src="wasm_exec.js"></script>
    <script src="app.js"></script>
    <script src="game.js"></script>
</body>
</html>
```

## Performance Considerations

### Efficiency Factors

**High Performance:**
- Go WASM provides near-native performance for game logic
- ECS architecture is cache-friendly and scales well
- Minimal JavaScript ↔ WASM boundary crossings
- Game logic runs in compiled Go code

**Optimization Strategies:**
- Batch entity updates to reduce JS calls
- Use object pools for frequently created/destroyed entities
- Implement spatial partitioning for collision detection
- Serialize only visible/relevant entities to JavaScript

### Memory Management

- Go's garbage collector handles memory in WASM
- Use sync.Pool for temporary objects
- Avoid excessive allocations in hot paths
- Monitor memory usage with browser dev tools

## Multiplayer Architecture

### Client-Side Prediction Pattern

```go
// Immediate local response
func (gh *GameHandler) ProcessPlayerInput(input map[string]interface{}) {
    // Apply input immediately for responsive feel
    gh.engine.ApplyInput(input)
}

// Server reconciliation
func (gh *GameHandler) ProcessServerUpdate(serverState map[string]interface{}) {
    // Compare server state with local state
    // Correct any discrepancies
    // Re-apply any inputs that occurred after server state timestamp
}
```

### Server Options

1. **Extend Current Go Server**: Add WebSocket support to `cmd/server/main.go`
2. **Dedicated Game Server**: Separate authoritative server process  
3. **Hybrid Approach**: Local for single-player, server for multiplayer

## Development Workflow

### Adding New Features

1. **Define Components**: Add to `internal/core/ecs.go`
2. **Create Systems**: Implement in `internal/core/systems.go`  
3. **Add WASM API**: Create handlers in `internal/api/game_handlers.go`
4. **Register Functions**: Update `cmd/wasm/main.go`
5. **Update Client**: Modify JavaScript in `web/game.js`
6. **Build & Test**: `make wasm && make dev`

### Testing Strategy

- Unit test core systems in pure Go
- Integration test WASM functions  
- Browser testing for rendering and input
- Load test multiplayer with multiple clients

## Next Steps

1. **Start Simple**: Implement basic ECS with Transform/Velocity
2. **Add Rendering**: Sprite system with texture loading
3. **Implement Input**: Keyboard/mouse handling
4. **Add Networking**: WebSocket integration for multiplayer
5. **Optimize**: Profile and optimize hot paths
6. **Scale**: Add more complex systems (AI, physics, audio)

This architecture provides a solid foundation for building high-performance games in Go WASM while leveraging JavaScript for I/O operations and rendering.