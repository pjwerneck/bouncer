Wait for an event to be triggered.

### Basic Operation
- Blocks until event is triggered or timeout
- Returns message from triggering client
- Returns 200 OK with message in body
- If `maxwait` is negative, waits indefinitely
- If `maxwait` is 0, returns immediately

### Usage Tips
- Set reasonable `maxwait` to avoid indefinite blocking
- Check response body for event message
- Multiple clients can wait for same event
- All waiting clients receive same message
