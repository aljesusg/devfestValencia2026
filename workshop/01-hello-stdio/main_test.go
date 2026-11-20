package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestGreet(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		{name: "Ada", want: "Ada"},
		{name: "  Héctor  ", want: "Héctor"},
		{name: " \n\t ", want: "DevFest attendee"},
		{name: "Ada\nLovelace", want: "Ada Lovelace"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := greet(context.Background(), nil, GreetInput{Name: tc.name})
			if err != nil {
				t.Fatal(err)
			}
			if result.IsError || len(result.Content) != 1 {
				t.Fatalf("unexpected result: %+v", result)
			}
			text, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Fatalf("expected text content, got %T", result.Content[0])
			}
			if !strings.Contains(text.Text, "Hola, "+tc.want+"!") ||
				!strings.Contains(text.Text, "MCP client ──stdio──▶ devfest-mcp") ||
				!strings.Contains(text.Text, "greet(") {
				t.Fatalf("unexpected greeting: %s", text.Text)
			}
		})
	}
}

func TestStdio(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	binary := filepath.Join(t.TempDir(), "devfest-mcp")
	if output, err := exec.CommandContext(ctx, "go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build server: %v\n%s", err, output)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "workshop-test", Version: "0.1.0"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{
		Command: exec.CommandContext(ctx, binary),
	}, nil)
	if err != nil {
		t.Fatalf("MCP handshake: %v", err)
	}
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools.Tools) != 1 || tools.Tools[0].Name != "greet" {
		t.Fatalf("expected only greet, got %+v", tools.Tools)
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "greet", Arguments: map[string]any{"name": "Ada"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || len(result.Content) != 1 {
		t.Fatalf("unexpected tool result: %+v", result)
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok || !strings.Contains(text.Text, `greet("Ada") ✓`) {
		t.Fatalf("unexpected greeting content: %+v", result.Content)
	}
	t.Log(text.Text)
}
