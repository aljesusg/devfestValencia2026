package main

import (
	"context"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Counts describe agenda entries, including breaks, not attendance or popularity.
type ScheduleStats struct {
	Event          string         `json:"event"`
	Timezone       string         `json:"timezone"`
	TotalEntries   int            `json:"total_entries"`
	ByDay          map[string]int `json:"entries_by_day"`
	ByTrack        map[string]int `json:"entries_by_track"`
	BySpeakerLabel map[string]int `json:"entries_by_speaker_label"`
	ByStartTime    map[string]int `json:"entries_by_start_time"`
	MissingSpeaker int            `json:"missing_speaker"`
	MissingTrack   int            `json:"missing_track"`
	UnparsedTime   int            `json:"unparsed_start_time"`
}

func makeScheduleStats(schedule *Schedule) func(context.Context, *mcp.CallToolRequest, ListSessionsInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input ListSessionsInput) (*mcp.CallToolResult, any, error) {
		stats := ScheduleStats{
			Event: schedule.Event, Timezone: schedule.Timezone,
			ByDay: make(map[string]int), ByTrack: make(map[string]int),
			BySpeakerLabel: make(map[string]int), ByStartTime: make(map[string]int),
		}
		for _, session := range schedule.Filter(strings.TrimSpace(input.Day), strings.TrimSpace(input.Track)) {
			stats.TotalEntries++
			stats.ByDay[session.Day]++
			if track := strings.TrimSpace(session.Track); track != "" {
				stats.ByTrack[track]++
			} else {
				stats.MissingTrack++
			}
			if speaker := strings.TrimSpace(session.Speaker); speaker != "" {
				stats.BySpeakerLabel[speaker]++
			} else {
				stats.MissingSpeaker++
			}
			parsed := false
			for _, layout := range []string{"15:04", "3:04 PM"} {
				if start, err := time.Parse(layout, strings.TrimSpace(session.Time)); err == nil {
					stats.ByStartTime[start.Format("15:04")]++
					parsed = true
					break
				}
			}
			if !parsed {
				stats.UnparsedTime++
			}
		}
		return textResult(stats)
	}
}
