# MCP Chat Agent with Gemini Integration

A simple chat agent that integrates Gemini 2.5 Pro API with MCP (Model Context Protocol) servers for tool calling in corporate environments.

## Components

- **MCP Server** (`main.go`) - Simple MCP server with `hello_world` tool
- **Orchestrator** (`orchestrator.go`) - Core logic bridging Gemini API and MCP servers
- **Chat Interface** (`chat.go`) - CLI interface for testing

## Prerequisites

- Go 1.21 or higher
- Access to Gemini 2.5 Pro Vertex API endpoint
- Bearer token for authentication

## Build Instructions

### 1. Build MCP Server
```bash
go build -o mcp-test-server.exe main.go
```

### 2. Build Chat Agent
```bash
go build -o chat-agent.exe chat.go orchestrator.go
```

### 3. Verify Build
```bash
# Check files exist
dir *.exe
# Should show: mcp-test-server.exe and chat-agent.exe
```

## Configuration

### 1. Update Configuration in chat.go
Edit the configuration section in `chat.go`:

```go
orchestrator := NewOrchestrator(
    "https://your-vertex-endpoint.com/v1/models/gemini-2.5-pro:generateContent",  // Your endpoint
    "your-bearer-token-here",                                                     // Your token
    "./mcp-test-server.exe",                                                      // MCP server path
)
```

### 2. Test MCP Server Independently (Optional)
```bash
# Test MCP server responds to initialize
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {"tools": {}}}}' | .\mcp-test-server.exe
```

Expected response should include server info and capabilities.

## Running the Chat Agent

### 1. Start the Chat Interface
```bash
.\chat-agent.exe
```

### 2. Test Commands

**Test MCP Tool Integration:**
```
You: Hello Alice
Agent: [Should use MCP tool and respond with greeting]
```

**Test Regular Chat:**
```
You: What is 2+2?
Agent: [Regular Gemini response without tools]
```

**Test Tool Trigger Variations:**
```
You: Say hello to Bob
You: Greet John
You: Hello there, my name is Sarah
```

### 3. Exit
```
You: quit
```

## Troubleshooting

### MCP Server Issues
```bash
# Test MCP server manually
.\mcp-test-server.exe
# Then paste: {"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}
```

### Gemini API Issues
- Verify endpoint URL is correct
- Check bearer token is valid
- Ensure corporate firewall allows HTTPS traffic
- Certificate issues are handled (InsecureSkipVerify enabled)

### Build Issues
```bash
# Clean and rebuild
go clean
go mod tidy
go build -o mcp-test-server.exe main.go
go build -o chat-agent.exe chat.go orchestrator.go
```

## Expected Flow

1. **User Input** → `"Hello John"`
2. **Gemini API** → Recognizes need for `hello_world` tool
3. **MCP Server** → Executes `hello_world` with `name: "John"`
4. **Tool Result** → `"Hello, John! How are you doing today?"`
5. **Gemini API** → Formats natural response
6. **User** → Receives final response

## Adding More Tools

1. **Extend MCP Server** - Add more tools in `main.go`
2. **Rebuild MCP Server** - `go build -o mcp-test-server.exe main.go`
3. **Test** - Tools automatically available to Gemini

## Security Notes

- Certificate verification is disabled for corporate environments
- Only use in trusted corporate networks
- Bearer tokens should be kept secure
- MCP server runs locally for security

## File Structure
```
├── main.go              # MCP server implementation
├── orchestrator.go      # Core orchestration logic
├── chat.go             # CLI interface
├── go.mod              # Go module file
├── mcp-test-server.exe # Built MCP server
├── chat-agent.exe      # Built chat agent
└── README.md           # This file
```