package main

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func fixture() *Schedule {
	return &Schedule{
		Event: "DevFest", Timezone: "CET",
		Sessions: []Session{
			{ID: "s1", Day: "Friday", DayLocal: "Viernes", Track: "Cloud", Time: "4:00 PM", Speaker: "Alberto"},
			{ID: "s2", Day: "Friday", DayLocal: "Viernes", Track: "Cloud", Time: "16:00", Speaker: "Ada"},
			{ID: "s3", Day: "Saturday", DayLocal: "Sabado", Time: "TBA"},
		},
	}
}

func TestScheduleTools(t *testing.T) {
	ctx := context.Background()
	schedule := fixture()
	_, value, err := makeListSessions(schedule)(ctx, nil, ListSessionsInput{Day: " viernes ", Track: "cloud"})
	if err != nil || len(value.([]Session)) != 2 {
		t.Fatalf("day/track filtering: value=%v err=%v", value, err)
	}
	_, value, err = makeFindSpeaker(schedule)(ctx, nil, FindSpeakerInput{Name: " ALB "})
	if err != nil || len(value.([]Session)) != 1 || value.([]Session)[0].ID != "s1" {
		t.Fatalf("speaker matching: value=%v err=%v", value, err)
	}
	_, value, err = makeGetSession(schedule)(ctx, nil, GetSessionInput{ID: " s1 "})
	if err != nil || value.(Session).ID != "s1" {
		t.Fatalf("get session: value=%v err=%v", value, err)
	}
	for _, id := range []string{"", "missing"} {
		if _, _, err := makeGetSession(schedule)(ctx, nil, GetSessionInput{ID: id}); err == nil {
			t.Fatalf("expected error for id %q", id)
		}
	}
	if _, _, err := makeFindSpeaker(schedule)(ctx, nil, FindSpeakerInput{Name: " "}); err == nil {
		t.Fatal("blank speaker must fail")
	}
	result, _, err := makeListSessions(schedule)(ctx, nil, ListSessionsInput{Day: "Sunday"})
	if err != nil || result.Content[0].(*mcp.TextContent).Text != "[]" {
		t.Fatalf("expected empty JSON array, got %+v, %v", result, err)
	}
}

func TestStats(t *testing.T) {
	stats := summarize(fixture(), ListSessionsInput{})
	if stats.TotalEntries != 3 || stats.MissingSpeaker != 1 || stats.MissingTrack != 1 || stats.UnparsedTime != 1 {
		t.Fatalf("unexpected totals: %+v", stats)
	}
	if !reflect.DeepEqual(stats.ByStartTime, map[string]int{"16:00": 2}) ||
		stats.ByDay["Friday"] != 2 || stats.ByTrack["Cloud"] != 2 || stats.BySpeakerLabel["Alberto"] != 1 {
		t.Fatalf("unexpected distributions: %+v", stats)
	}
	if got := summarize(fixture(), ListSessionsInput{Day: " viernes ", Track: "cloud"}); got.TotalEntries != 2 {
		t.Fatalf("filtered statistics: %+v", got)
	}
	if got := summarize(fixture(), ListSessionsInput{Day: "Sunday"}); got.TotalEntries != 0 || got.ByDay == nil {
		t.Fatalf("empty statistics: %+v", got)
	}
}

func TestStdio(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	binary := filepath.Join(t.TempDir(), "devfest-mcp")
	if output, err := exec.CommandContext(ctx, "go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "workshop-test", Version: "0.2.0"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: exec.CommandContext(ctx, binary)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tool := range tools.Tools {
		names[tool.Name] = true
	}
	if !reflect.DeepEqual(names, map[string]bool{"list_sessions": true, "get_session": true, "find_speaker": true, "schedule_stats": true}) {
		t.Fatalf("unexpected tools: %v", names)
	}
	for _, tc := range []struct {
		name    string
		args    map[string]any
		isError bool
	}{
		{"list_sessions", map[string]any{"day": "Friday"}, false},
		{"get_session", map[string]any{"id": "s5"}, false},
		{"find_speaker", map[string]any{"name": "Alberto"}, false},
		{"schedule_stats", map[string]any{"day": "Friday"}, false},
		{"get_session", map[string]any{"id": "missing"}, true},
		{"get_session", map[string]any{}, true},
		{"find_speaker", map[string]any{"name": ""}, true},
	} {
		t.Run(tc.name+":"+toJSON(tc.args), func(t *testing.T) {
			result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tc.name, Arguments: tc.args})
			if err != nil {
				t.Fatal(err)
			}
			if result.IsError != tc.isError {
				t.Fatalf("unexpected result: %+v", result)
			}
			if !tc.isError {
				text := result.Content[0].(*mcp.TextContent).Text
				if !json.Valid([]byte(text)) {
					t.Fatalf("invalid JSON: %s", text)
				}
			}
		})
	}
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
