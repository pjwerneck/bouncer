Reset watchdog timer to prevent expiration notification.

### Basic Operation
- Resets expiration timer to `expires` milliseconds
- Returns immediately with 204 No Content
- Zero or negative `expires` triggers immediate expiration
- Default `expires` is 60000 (one minute)

### Usage Tips
- Kick frequently enough to prevent false alarms
- Set `expires` longer than maximum task interval
- Consider network latency when setting intervals
- Kick before heavy operations, not after
