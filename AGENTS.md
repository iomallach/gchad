# AGENTS.md

Guide for AI agents working in the gchad codebase.

## Project Overview

**gchad** is a WebSocket-based chat server written in Go. It implements a hub-and-spoke architecture for broadcasting messages to multiple connected clients.

- **Type**: WebSocket chat server
- **Language**: Go 1.25.3
- **Dependencies**: 
  - `github.com/gorilla/websocket` - WebSocket implementation
  - `github.com/google/uuid` - UUID generation for client IDs

## Development Environment

This project uses **Nix flakes** for development environment management. The flake provides:
- Go 1.25.3
- gotools
- golangci-lint
- delve (debugger)
- gopls (language server)

To enter the development environment:
```bash
nix develop
```

**Note**: Go commands require the Nix dev environment. If you see `"go": executable file not found`, you need to run commands inside `nix develop` or assume the user has Go installed another way.

## Essential Commands

### Building
```bash
go build ./cmd                    # Build the server binary
go build -o gchad ./cmd          # Build with custom output name
```

### Testing
```bash
go test                          # Run tests in current directory
go test -v                       # Run tests with verbose output
go test ./...                    # Run all tests recursively
```

### Running
```bash
go run ./cmd/main.go             # Run the server directly
./gchad                          # Run built binary
```

The server starts on `localhost:8080` with WebSocket endpoint at `/chat`.

### Linting
```bash
golangci-lint run                # Run linter (available in Nix env)
```

### Dependency Management
```bash
go mod tidy                      # Clean up dependencies
go mod download                  # Download dependencies
```

## Project Structure

```
gchad/
├── cmd/
│   └── main.go           # Server entry point, Hub/Client implementation
├── message.go            # Message types and serialization
├── message_test.go       # Message marshaling tests
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
└── flake.nix            # Nix development environment
```

### Architecture

**Package Organization**:
- `main` package in `cmd/main.go` - Server implementation with Hub and Client types
- `gchad` package in root - Shared message types and interfaces

**Key Components**:
1. **Hub** - Central message broker that:
   - Maintains map of connected clients
   - Handles client registration/unregistration
   - Broadcasts messages to all clients
   
2. **Client** - Represents individual WebSocket connection with:
   - Read loop (receives messages from WebSocket)
   - Write loop (sends messages to WebSocket with ping/pong)
   - Buffered send channel (256 byte capacity)
   - UUID-based identification

3. **Message System** - Type-safe message handling with:
   - `Messager` interface for all message types
   - `MessageType` enum (SystemUserJoined, SystemUserLeft, UserMsg)
   - Custom JSON marshaling/unmarshaling

## Code Conventions

### Naming Patterns
- **Types**: PascalCase (e.g., `UserMessage`, `MessageType`)
- **Functions/Methods**: camelCase (e.g., `readLoop`, `serveWs`)
- **Constants**: PascalCase or camelCase (e.g., `SystemUserJoined`, `writeWait`)
- **Package-level variables**: camelCase (e.g., `upgrader`)

### Project-Specific Patterns

**Channels**:
- Hub uses unbuffered channels for registration/unregistration
- Hub broadcast channel is unbuffered (typed as `chan gchad.Message` in Hub, but `chan []byte` in main - **inconsistency present**)
- Client send channels are buffered with capacity 256

**Timeouts**:
- Write wait: 10 seconds
- Pong wait: 60 seconds  
- Ping period: 54 seconds (9/10 of pong wait)

**WebSocket Configuration**:
- Read/Write buffer: 1024 bytes each
- CORS: Currently accepts all origins (`CheckOrigin` returns true)

**Error Handling**:
- WebSocket close errors: Only log unexpected close errors (not normal `CloseGoingAway` or `CloseAbnormalClosure`)
- JSON unmarshal errors: Logged but don't break the read loop
- Most errors use `log.Println` or `log.Printf`

### Message Type System

The codebase uses a **tagged union pattern** for messages:

```go
type Message struct {
    Inner       Messager    `json:"inner"`
    MessageType MessageType `json:"message_type"`
}
```

**Important**: 
- All message types must implement `Messager` interface (just `MessageMark()` method)
- Custom `MarshalJSON`/`UnmarshalJSON` methods handle type discrimination
- `MessageType` enum determines which concrete type to unmarshal into
- When adding new message types:
  1. Add to `MessageType` enum
  2. Create struct implementing `Messager`
  3. Update switch cases in `UnmarshalJSON`

**JSON Structure**:
```json
{
  "inner": {
    "timestamp": "2023-01-01T00:00:00Z",
    "message": "Hello"
  },
  "message_type": 2
}
```

## Testing Approach

### Test Organization
- Tests live in `*_test.go` files alongside implementation
- Package: `package gchad` (same as implementation)

### Test Patterns

**Table-driven tests**: All current tests use table-driven pattern:
```go
testCases := []struct {
    expected Type
    input    Type
}{
    // test cases
}

for _, testCase := range testCases {
    // test logic
}
```

**Message Testing**:
- Marshaling tests verify JSON output matches expected string exactly
- Unmarshaling tests verify:
  1. No errors during unmarshal
  2. Type assertion succeeds
  3. All fields match expected values
  4. MessageType matches

**Test Data**:
- Uses zero-value `time.Time{}` for consistent timestamp comparisons
- Tests all three message types (UserMsg, SystemUserJoined, SystemUserLeft)

### Current Test Coverage
- Message marshaling (JSON serialization)
- Message unmarshaling (JSON deserialization)
- **Missing**: Hub logic, Client loops, WebSocket integration

## Important Gotchas

### Type Inconsistencies
⚠️ **Hub broadcast channel type mismatch**: 
- Hub struct declares: `broadcast chan gchad.Message`
- Main function initializes: `broadcast: make(chan []byte)`
- This inconsistency exists in the codebase and should be addressed

### Concurrency Patterns
- Each client spawns two goroutines (readLoop, writeLoop)
- Hub runs in its own goroutine
- Channel operations are the synchronization mechanism
- **No mutexes used** - all shared state accessed via channels

### WebSocket Quirks
- `writeLoop` batches multiple pending messages in the send channel before closing the writer
- Ping/pong heartbeat required to detect dead connections
- `SetWriteDeadline` called before each write operation
- Client cleanup happens in `defer` blocks

### JSON Marshaling
- The custom marshaling code uses `json.RawMessage` for the `Inner` field
- When unmarshaling, must check `MessageType` first to know which concrete type to create
- Struct tags (like `json:"message"`) are respected in both marshal and unmarshal

## Recent Changes (Git History)

Based on recent commits:

1. **Removed SystemWhoIsInTheRoom message type** - Simplified message types
2. **Fixed marshaling to respect struct tags** - Previously JSON field names were wrong
3. **Moved message types to separate file** - Better code organization
4. **Implemented custom marshaling** - Type-safe message serialization

## TODOs in Codebase

From code comments:

1. `cmd/main.go:27` - Hub field in Client should probably just be the broadcast channel, not the entire hub
2. `cmd/main.go:121` - Notify all clients when a client connects
3. `cmd/main.go:126` - Notify all clients when a client disconnects

## Development Workflow

When making changes:

1. **Read the relevant files** - Understand context before editing
2. **Make changes** - Follow existing patterns and naming
3. **Run tests**: `go test -v`
4. **Build**: `go build ./cmd` to verify compilation
5. **Manual test** if WebSocket logic changed (connect client, send messages)

### Adding a New Message Type

Example workflow:
1. Add constant to `MessageType` enum in `message.go`
2. Create new struct implementing `Messager` interface
3. Add JSON struct tags to fields
4. Update `UnmarshalJSON` switch statement with new case
5. Add test cases to both marshal and unmarshal tests
6. Run `go test -v` to verify

### Modifying Hub/Client Logic

- Be careful with channel operations (blocking vs non-blocking)
- Consider goroutine lifecycle and cleanup
- Test connection/disconnection scenarios manually
- Check for deadlocks and race conditions

## Dependencies

Only two external dependencies:
- `github.com/gorilla/websocket v1.5.3` - Battle-tested WebSocket library
- `github.com/google/uuid v1.6.0` - UUID generation

Keep dependencies minimal - consider carefully before adding new ones.

## Running the Server

Default configuration:
- **Port**: 8080
- **Endpoint**: `/chat`
- **Protocol**: WebSocket

To test manually:
```javascript
// In browser console
const ws = new WebSocket('ws://localhost:8080/chat');
ws.onmessage = (e) => console.log('Received:', e.data);
ws.send(JSON.stringify({
  inner: { timestamp: new Date().toISOString(), message: "Hello" },
  message_type: 2
}));
```

## Code Style Preferences

Based on existing code:

- **Inline struct definitions** for small helper structs (e.g., in MarshalJSON)
- **Early returns** for error cases
- **Defer for cleanup** (connection close, channel close, ticker stop)
- **Logging**: Use `log.Println` for info, `log.Printf` for formatted output
- **No extensive comments** - code is mostly self-documenting
- **TODOs** marked with `// TODO:` prefix

## Performance Considerations

Current implementation choices:
- Small buffer sizes (1024 bytes for WebSocket, 256 for client send channel)
- Message batching in writeLoop to reduce syscalls
- Heartbeat to avoid dead connection accumulation
- No rate limiting or message size limits (potential DOS vector)

---

*This document should be updated as the codebase evolves. When you discover new patterns, commands, or gotchas, add them here.*
