package main

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Demo 🚀",
		"1.0.0",
	)

	// Add tool
	tool := mcp.NewTool(
		"hello_world",
		mcp.WithDescription("Say hello to someone"),
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("Name of the person to greet"),
		),
	)

	// Add tool handler
	s.AddTool(tool, func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		// Get the name parameter from arguments
		nameArg, exists := arguments["name"]
		if !exists {
			return mcp.NewToolResultError("name parameter is required"), nil
		}

		name, ok := nameArg.(string)
		if !ok {
			return mcp.NewToolResultError("name must be a string"), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Hello, %s! %s", name,"How are you doing today?")), nil
	})

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}