package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

// Gemini API structures
type GeminiRequest struct {
	Contents []Content `json:"contents"`
	Tools    []Tool    `json:"tools,omitempty"`
}

type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Text             string            `json:"text,omitempty"`
	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
}

type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type FunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"function_declarations"`
}

type FunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content      Content `json:"content"`
	FinishReason string  `json:"finishReason"`
}

// MCP structures
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type MCPToolsListResult struct {
	Tools []MCPTool `json:"tools"`
}

type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type MCPCallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type MCPCallToolResult struct {
	Content []MCPContent `json:"content"`
}

type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Orchestrator
type Orchestrator struct {
	geminiEndpoint string
	geminiToken    string
	mcpServerPath  string
}

func NewOrchestrator(geminiEndpoint, geminiToken, mcpServerPath string) *Orchestrator {
	return &Orchestrator{
		geminiEndpoint: geminiEndpoint,
		geminiToken:    geminiToken,
		mcpServerPath:  mcpServerPath,
	}
}

func (o *Orchestrator) callGemini(request GeminiRequest) (*GeminiResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", o.geminiEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.geminiToken)

	// Create HTTP client with disabled certificate verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var geminiResp GeminiResponse
	err = json.Unmarshal(body, &geminiResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v, body: %s", err, string(body))
	}

	return &geminiResp, nil
}

func (o *Orchestrator) callMCPServer(method string, params interface{}) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(o.mcpServerPath)
	cmd.Stdin = bytes.NewBuffer(jsonData)
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("MCP server error: %v", err)
	}

	// Debug: print raw output
	fmt.Printf("DEBUG: MCP Server Output: %s\n", string(output))

	if len(output) == 0 {
		return nil, fmt.Errorf("MCP server returned empty response")
	}

	var mcpResp MCPResponse
	err = json.Unmarshal(output, &mcpResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MCP response: %v, output: %s", err, string(output))
	}

	return &mcpResp, nil
}

func (o *Orchestrator) getAvailableTools() ([]FunctionDeclaration, error) {
	// Initialize MCP server first
	initParams := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
	}
	
	_, err := o.callMCPServer("initialize", initParams)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MCP server: %v", err)
	}

	// Get tools from MCP server
	resp, err := o.callMCPServer("tools/list", nil)
	if err != nil {
		return nil, err
	}

	// Convert MCP response to tools list
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}

	var toolsList MCPToolsListResult
	err = json.Unmarshal(resultBytes, &toolsList)
	if err != nil {
		return nil, err
	}

	// Convert to Gemini format
	var functions []FunctionDeclaration
	for _, tool := range toolsList.Tools {
		functions = append(functions, FunctionDeclaration{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.InputSchema,
		})
	}

	return functions, nil
}

func (o *Orchestrator) executeTool(functionCall *FunctionCall) (string, error) {
	params := MCPCallToolParams{
		Name:      functionCall.Name,
		Arguments: functionCall.Args,
	}

	resp, err := o.callMCPServer("tools/call", params)
	if err != nil {
		return "", err
	}

	// Parse tool result
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return "", err
	}

	var toolResult MCPCallToolResult
	err = json.Unmarshal(resultBytes, &toolResult)
	if err != nil {
		return "", err
	}

	// Extract text from result
	if len(toolResult.Content) > 0 && toolResult.Content[0].Type == "text" {
		return toolResult.Content[0].Text, nil
	}

	return "Tool executed successfully", nil
}

func (o *Orchestrator) ProcessMessage(userMessage string) (string, error) {
	// Get available tools
	tools, err := o.getAvailableTools()
	if err != nil {
		return "", fmt.Errorf("failed to get tools: %v", err)
	}

	// Initial request to Gemini
	request := GeminiRequest{
		Contents: []Content{
			{
				Role:  "user",
				Parts: []Part{{Text: userMessage}},
			},
		},
		Tools: []Tool{{FunctionDeclarations: tools}},
	}

	resp, err := o.callGemini(request)
	if err != nil {
		return "", fmt.Errorf("Gemini API error: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	candidate := resp.Candidates[0]
	
	// Check if Gemini wants to call a function
	if len(candidate.Content.Parts) > 0 && candidate.Content.Parts[0].FunctionCall != nil {
		functionCall := candidate.Content.Parts[0].FunctionCall
		
		// Execute the tool
		toolResult, err := o.executeTool(functionCall)
		if err != nil {
			return "", fmt.Errorf("tool execution error: %v", err)
		}

		// Send tool result back to Gemini
		followUpRequest := GeminiRequest{
			Contents: []Content{
				{Role: "user", Parts: []Part{{Text: userMessage}}},
				{Role: "model", Parts: []Part{{FunctionCall: functionCall}}},
				{Role: "function", Parts: []Part{{
					FunctionResponse: &FunctionResponse{
						Name:     functionCall.Name,
						Response: map[string]interface{}{"result": toolResult},
					},
				}}},
			},
			Tools: []Tool{{FunctionDeclarations: tools}},
		}

		finalResp, err := o.callGemini(followUpRequest)
		if err != nil {
			return "", fmt.Errorf("Gemini follow-up error: %v", err)
		}

		if len(finalResp.Candidates) > 0 && len(finalResp.Candidates[0].Content.Parts) > 0 {
			return finalResp.Candidates[0].Content.Parts[0].Text, nil
		}
	}

	// Direct text response
	if len(candidate.Content.Parts) > 0 {
		return candidate.Content.Parts[0].Text, nil
	}

	return "No response generated", nil
}

