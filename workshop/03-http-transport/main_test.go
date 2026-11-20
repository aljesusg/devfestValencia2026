package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHTTP(t *testing.T) {
	schedule, err := LoadSchedule("../data/schedule.json")
	if err != nil {
		t.Fatal(err)
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "http-test", Version: "0.3.0"}, nil)
	registerTools(server, schedule)
	host := httptest.NewServer(newHTTPHandler(server))
	defer host.Close()

	response, err := http.Get(host.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	var health struct{ Status string }
	err = json.NewDecoder(response.Body).Decode(&health)
	response.Body.Close()
	if err != nil || response.StatusCode != http.StatusOK || health.Status != "ok" {
		t.Fatalf("health check failed: %+v, %v", health, err)
	}
	response, err = http.Post(host.URL+"/health", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusMethodNotAllowed || response.Header.Get("Allow") != "GET" {
		t.Fatal("health endpoint must reject POST and advertise GET")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: host.URL + "/mcp"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	tools, err := session.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 4 {
		t.Fatalf("expected four HTTP tools: %+v, %v", tools, err)
	}
	for _, tc := range []struct {
		name string
		args map[string]any
	}{
		{"list_sessions", map[string]any{"day": "Friday"}},
		{"get_session", map[string]any{"id": "s5"}},
		{"find_speaker", map[string]any{"name": "Alberto"}},
		{"schedule_stats", map[string]any{"day": "Friday"}},
	} {
		result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tc.name, Arguments: tc.args})
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if result.IsError || len(result.Content) == 0 {
			t.Fatalf("%s: unexpected result %+v", tc.name, result)
		}
		if tc.name == "schedule_stats" {
			var stats ScheduleStats
			if err := json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &stats); err != nil {
				t.Fatal(err)
			}
			if stats.TotalEntries == 0 || stats.ByDay["Friday"] != stats.TotalEntries {
				t.Fatalf("incorrect Friday statistics: %+v", stats)
			}
		}
	}
}
