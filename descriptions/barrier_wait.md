Synchronize multiple clients at a common barrier point.

### Basic Operation
- Blocks until `size` clients reach the barrier
- Returns 204 No Content when barrier triggers
- Returns 408 Request Timeout on `maxwait`
- Returns 409 Conflict if barrier already triggered
- If `maxwait` is negative, waits indefinitely
- If `maxwait` is 0, returns immediately

### Usage Tips
- Default size is 2 clients
- All waiting clients are released simultaneously
- Use for multi-party synchronization
- Consider network latency when setting timeouts
- Barriers cannot be reused after triggering
