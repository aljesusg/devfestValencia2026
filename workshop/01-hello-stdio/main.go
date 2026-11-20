package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GreetInput is the structured input for the greet tool.
type GreetInput struct {
	Name string `json:"name" jsonschema:"the person to greet"`
}

func greet(_ context.Context, _ *mcp.CallToolRequest, input GreetInput) (*mcp.CallToolResult, any, error) {
	name := strings.Join(strings.Fields(input.Name), " ")
	if name == "" {
		name = "DevFest attendee"
	}
	// Return the Kiali-inspired graph as tool content, never directly on stdout:
	// stdout belongs exclusively to the MCP transport.
	msg := fmt.Sprintf("Hola, %s! Et donem la benvinguda al DevFest València!\n\n"+
		"MCP client ──stdio──▶ devfest-mcp\n"+
		"                      └─ greet(%q) ✓", name, name)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}, nil, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "devfest-mcp", Version: "0.1.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "greet",
		Description: "Welcome a DevFest attendee in Valencian with a Kiali-inspired MCP call graph",
	}, greet)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
