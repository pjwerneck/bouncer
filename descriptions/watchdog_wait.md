Monitor a periodic task by waiting for its watchdog to expire.

### Basic Operation
- Blocks until watchdog expires or timeout occurs
- Returns 204 No Content when watchdog expires
- Returns 408 Request Timeout on `maxwait`
- If `maxwait` is negative, waits indefinitely
- If `maxwait` is 0, returns immediately

### Usage Tips
- Multiple clients can monitor same watchdog
- All waiting clients are notified on expiration
- Use reasonable `maxwait` for monitoring tasks
- Combine with alerts/monitoring systems
