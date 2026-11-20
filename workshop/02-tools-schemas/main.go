package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	dataPath := os.Getenv("SCHEDULE_PATH")
	if dataPath == "" {
		dataPath = filepath.Join("..", "data", "schedule.json")
	}

	schedule, err := LoadSchedule(dataPath)
	if err != nil {
		log.Fatalf("load schedule: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "devfest-mcp",
		Version: "0.2.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{Name: "list_sessions", Description: "List DevFest sessions, optionally filtered by day (Friday/Saturday or the local day name) and track"}, makeListSessions(schedule))
	mcp.AddTool(server, &mcp.Tool{Name: "get_session", Description: "Get a single DevFest session by its exact id; use list_sessions to discover ids"}, makeGetSession(schedule))
	mcp.AddTool(server, &mcp.Tool{Name: "find_speaker", Description: "Find DevFest sessions by speaker name (case-insensitive substring match)"}, makeFindSpeaker(schedule))
	mcp.AddTool(server, &mcp.Tool{Name: "schedule_stats", Description: "Explore the local agenda: session counts by day, track, speaker label and start time, plus missing metadata counts. Includes non-talk agenda entries; speaker labels are not individual people. Optional day/track filters."}, makeScheduleStats(schedule))

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
