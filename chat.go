package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Configuration - update these with your actual values
	orchestrator := NewOrchestrator(
		"",     // Gemini Flash endpoint
		"",               // Replace with your token
		"./mcp-test-server.exe",    // Path to your MCP server
	)

	fmt.Println("🤖 MCP Chat Agent Started!")
	fmt.Println("Type 'quit' to exit")
	fmt.Println("Available commands:")
	fmt.Println("  - Hello [name] (will use MCP hello_world tool)")
	fmt.Println("  - Any other message (regular chat)")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "quit" {
			fmt.Println("Goodbye!")
			break
		}
		
		if input == "" {
			continue
		}

		fmt.Print("Agent: ")
		response, err := orchestrator.ProcessMessage(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("%s\n", response)
		}
		fmt.Println()
	}
}
