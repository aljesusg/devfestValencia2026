# Step 02 — Tools & schemas

**Timebox:** ~10 minutes

## Goal

Grow the server from one toy tool to a small **DevFest schedule** API:

| Tool | Purpose |
|------|---------|
| `list_sessions` | List sessions (optional `day` filter: Friday / Saturday) |
| `get_session` | Get one session by `id` |
| `find_speaker` | Find sessions by speaker name (substring match) |
| `schedule_stats` | Descriptive agenda statistics with optional day/track filters |

Data lives in [`../data/schedule.json`](../data/schedule.json) — a **local snapshot** of the real DevFest Valencia agenda (no network during the exercise).

Refresh that snapshot anytime from the live GDG API:

```bash
cd ../../demo-example && go run . -dump=../workshop/data/schedule.json
```

`mcp.AddTool` builds JSON Schema from your Go structs (`json` / `jsonschema` tags). That is the same idea production Go MCP servers rely on: typed inputs the model can call safely.

## What you will edit

The handlers and registration are implemented in `tools.go` and `main.go`.
The bonus analytics tool lives in `stats.go`.

### Try in Inspector

| Tool | Arguments |
|------|-----------|
| `list_sessions` | `{"day":"Friday"}` |
| `get_session` | `{"id":"s5"}` |
| `find_speaker` | `{"name":"Alberto"}` |
| `schedule_stats` | `{}` or `{"day":"Friday"}` |

`schedule_stats` returns counts by day, track, speaker label, and normalized
24-hour start time, plus missing-speaker, missing-track, and unparsed-time counts.
These are **agenda entries**, including breaks and check-in, not just talks.
Speaker labels are counted as stored (not split into individual people).
Missing tracks are reported, not invented. Start times are pooled across days;
filter by day to inspect a single day's distribution. This snapshot cannot tell
you attendance, popularity, or session duration.

Try asking an MCP-capable AI client: “Explore the agenda like a dataset: compare
Friday and Saturday, show the start-time distributions, and flag missing metadata.”

### Verify

```bash
go test -v ./...
go vet ./...
```

## Run

```bash
cd workshop/02-tools-schemas
go mod tidy
go run .
```

Optional: `npx @modelcontextprotocol/inspector go run .`

## Done when

- Four tools appear in the inspector (the three workshop tools plus `schedule_stats`)
- `list_sessions` with `{"day":"Friday"}` returns only Friday sessions

## Stuck?

Use [`solutions/02-tools-schemas`](../../solutions/02-tools-schemas).
