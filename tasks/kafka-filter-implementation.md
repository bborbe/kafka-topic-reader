# Plan: Add Simple String Filtering to Kafka Topic Reader

## Overview
Add a simple `filter` query parameter to the `/read` endpoint that searches for a string in the message values only. If no filter is specified, all messages are returned as before.

## Implementation Steps

### 1. Update Handler (`pkg/handler.go`)
- Add parsing of `filter` query parameter (optional string)
- Pass filter to `changesProvider.Changes()` method

### 2. Update ChangesProvider Interface (`pkg/changes.go`)
- Modify `Changes` method signature to accept optional filter parameter
- Update `changesProvider.Changes()` implementation to apply filtering after record conversion

### 3. Add Filtering Logic (`pkg/changes.go`)
- Add simple string filtering function that searches in record values
- Apply filter after `c.converter.Convert()` but before adding to channel
- Use case-insensitive string matching (`strings.Contains` with `strings.ToLower`)

## Key Design Decisions
- **Simple approach**: Only one `filter` parameter, searches in values only
- **Case-insensitive**: Easy to use, matches most user expectations  
- **Backward compatible**: No filter = no filtering (existing behavior)
- **Filter before limit**: Only matching records count toward the limit
- **Search in converted values**: Convert to JSON first, then search in the stringified result

This keeps the implementation minimal while providing useful filtering functionality.

## Current Implementation Analysis

### Handler Structure
The current handler in `pkg/handler.go` already parses query parameters:
- `topic` (required)
- `offset` (optional, defaults based on parsing)
- `limit` (optional, defaults to 100)
- `partition` (required)

### Changes Provider
The `ChangesProvider` interface in `pkg/changes.go` currently has:
```go
Changes(
    ctx context.Context,
    topic libkafka.Topic,
    partition libkafka.Partition,
    offset libkafka.Offset,
    limit uint64,
) (Records, error)
```

### Record Structure
Records contain:
- `Key` (string)
- `Value` (interface{}) - converted from JSON or error string
- `Offset`, `Partition`, `Topic`, `Header`

## Implementation Details

### Filtering Strategy
1. Parse the `filter` parameter in the handler
2. Pass it to the changes provider
3. In the message processing loop (line 107-124 in changes.go), add filtering after conversion
4. Only increment counter and send to channel if record matches filter
5. Search in the JSON-marshaled string representation of the `Value` field

### Code Changes Preview

#### Handler changes:
```go
filter := req.FormValue("filter") // Add this line
changes, err := changesProvider.Changes(ctx, topic, *partition, *offset, limit, filter)
```

#### ChangesProvider interface:
```go
Changes(
    ctx context.Context,
    topic libkafka.Topic,
    partition libkafka.Partition,
    offset libkafka.Offset,
    limit uint64,
    filter string, // Add this parameter
) (Records, error)
```

#### Filtering function:
```go
func matchesFilter(record *Record, filter string) bool {
    if filter == "" {
        return true // No filtering
    }
    
    // Convert value to searchable string
    valueStr := ""
    if record.Value != nil {
        if jsonBytes, err := json.Marshal(record.Value); err == nil {
            valueStr = string(jsonBytes)
        }
    }
    
    return strings.Contains(strings.ToLower(valueStr), strings.ToLower(filter))
}
```
