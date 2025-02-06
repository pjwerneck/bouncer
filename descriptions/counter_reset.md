Reset counter to specified value.

### Basic Operation
- Sets counter to specified `value`
- Returns 204 No Content on success
- Default `value` is 0
- Operation is atomic

### Usage Tips
- Use for periodic resets
- Useful for time-based metrics
- Consider using delete instead of reset
- All clients see new value immediately
