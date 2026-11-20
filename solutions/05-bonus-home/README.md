# Step 05 — bonus

See [`../../workshop/05-bonus-home/README.md`](../../workshop/05-bonus-home/README.md) for the homework menu.

A minimal “resources” sketch you can paste into Step 03’s `main.go` after creating the server:

```go
server.AddResource(&mcp.Resource{
    URI:         "devfest://schedule",
    Name:        "schedule",
    Description: "Full sample DevFest schedule JSON",
    MIMEType:    "application/json",
}, func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
    raw, err := os.ReadFile(dataPath)
    if err != nil {
        return nil, err
    }
    return &mcp.ReadResourceResult{
        Contents: []mcp.ResourceContents{
            &mcp.TextResourceContents{
                URI:      "devfest://schedule",
                MIMEType: "application/json",
                Text:     string(raw),
            },
        },
    }, nil
})
```

Confirm the exact `AddResource` / contents types against the SDK version in your `go.mod` — APIs evolve slightly between releases.
