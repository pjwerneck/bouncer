Atomically increment or decrement a distributed counter.

### Basic Operation
- Adds `amount` to counter value atomically
- Returns new counter value
- Negative `amount` decrements the counter
- Default `amount` is 1

### Usage Tips
- Use for distributed counting/statistics
- Safe for concurrent access
- Combine with monitoring for thresholds
- Values are 64-bit signed integers
