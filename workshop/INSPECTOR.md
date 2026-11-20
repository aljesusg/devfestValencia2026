# Four steps, one Inspector

From the repository root (`mcp/`):

```bash
python3 workshop/launch_inspector.py
```

Or from its parent: `python3 mcp/workshop/launch_inspector.py`.

Requires Go, Python 3, Node.js and npx. The launcher builds the servers, starts a
loopback-only HTTP backend, and opens Inspector with **four named server entries**.
Select an entry, connect it, then open **Tools** and list the tools if necessary.
Inspector has server entries, not tool folders; the matching code stays in each
numbered workshop folder. Step 04 deliberately reuses Step 03's HTTP backend.

Default ports: UI **6294**, sandbox **6295**, proxy **6297** (for Inspector versions
that use one), app origin **6298**, MCP backend **8093**.
Use the browser URL printed by Inspector, including any authentication token.
Leave the terminal running; Ctrl+C stops the session and its HTTP backend.
If ports are busy:

```bash
python3 workshop/launch_inspector.py --client-port 6394 --http-port 8193
```

The backend port is independent from Inspector's UI port. Change the client URL
below to match `--http-port` when overriding it.

## 01 — Hello MCP · Valencian greeting

**Folder:** `01-hello-stdio/`

**Purpose:** typed Go function → discovered MCP tool → response over stdio.

Select `greet`. In the **name** field enter `Héctor` (plain text, without JSON
quotes). If using a JSON arguments editor, enter:

```json
{"name":"Héctor"}
```

Click **Run Tool**. Expect the Valencian welcome and Kiali-inspired call graph.
The graph is illustrative, not real Kiali telemetry.

## 02 — Tools & schemas · Explore the dataset

**Folder:** `02-tools-schemas/`

**Purpose:** query the offline agenda using structured input schemas.

Select a tool and enter the following into its JSON arguments editor, or put each
value into the corresponding named form field:

| Tool | JSON arguments | What to look for |
|---|---|---|
| `list_sessions` | `{"day":"Friday"}` | Only Friday entries; optional `track` can stay empty |
| `get_session` | `{"id":"s5"}` | Alberto's MCP workshop |
| `find_speaker` | `{"name":"Alberto"}` | Case-insensitive partial-name search |
| `schedule_stats` | `{}` | Dataset-wide distributions and missing-data counts |
| `schedule_stats` | `{"day":"Friday"}` | Filtered exploratory analysis |
| `get_session` | `{"id":"does-not-exist"}` | A clear tool error rather than a server crash |

Do not invent a track filter: inspect the data first; absent tracks are missing
metadata, not evidence that the agenda has a single track.

## 03 — HTTP transport · Agenda analytics

**Folder:** `03-http-transport/`

**Purpose:** call the same four agenda tools through Streamable HTTP.

The launcher configures the URL automatically. For a manual connection choose
**Streamable HTTP** (not legacy SSE) and enter this in **URL**:

```text
http://127.0.0.1:8093/mcp
```

Connect, list tools, and run `schedule_stats` with `{"day":"Saturday"}`.
Compare it to Friday. Try `list_sessions` with `{"day":"Saturday"}` to inspect the
underlying rows. This demonstrates that the schema and tools stay the same when
the transport changes.

Health check, from another terminal:

```bash
curl http://127.0.0.1:8093/health
```

Times are normalized to 24-hour `HH:MM`; counts include breaks/check-in.
Speaker labels are not a deduplicated count of people. Missing metadata and
unparseable times are reported explicitly. These data cannot establish session
duration, attendance, popularity, or causal conclusions.

## 04 — Connect a client · Guided investigation

**Folder:** `04-connect-cursor/`

**Purpose:** connect an AI host to the HTTP server, rather than create a fourth
implementation. Its Inspector entry is an independent client connection to the
same backend as Step 03.

First use the Step 04 entry in Inspector to run `get_session` with `{"id":"s5"}`.
Then, in **Cursor**, merge this entry into your project's `.cursor/mcp.json`
(`mcpServers` object). Do not overwrite existing entries:

```json
{
  "mcpServers": {
    "devfest-step-04": {
      "url": "http://127.0.0.1:8093/mcp"
    }
  }
}
```

Keep the launcher running, reload Cursor's MCP connections, enable this server,
and enter these prompts into **Cursor's AI chat**, not Inspector's arguments box:

> Use devfest-step-04 to find Alberto's sessions. Show their IDs, days, and times.

> Explore the agenda as a dataset. Call schedule_stats separately for Friday and
> Saturday. Compare entry counts and start-time distributions. Flag missing
> metadata and explain what this snapshot cannot tell us.

> Use the tools to plan my Friday around session s5. List the actual agenda
> entries in time order. Do not assume durations or invent talks.

Inspector manually exercises tools; an AI host chooses tools based on these
natural-language prompts. A successful Step 04 answer includes actual tool calls.

## Verify without opening Inspector

```bash
python3 workshop/launch_inspector.py --check
```

This builds all three implementations and checks HTTP health. Each Go module
also has tests: run `go test -v ./...` inside its numbered folder.
