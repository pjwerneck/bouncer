Rate limiting endpoint that implements the Token Bucket algorithm.

### Basic Operation
- Each request consumes one token
- Bucket is refilled with `size` tokens every `interval` milliseconds
- Waits up to `maxwait` milliseconds for available token
- If `maxwait` is negative, waits indefinitely
- If `maxwait` is 0, returns immediately

### Usage Tips
- For N operations per second, set `size=N` and `interval=1000`
- For fractional rates, reduce the fraction:
  - 10 ops/minute: use `size=1&interval=6000`
  - Not `size=10&interval=60000` (causes long waits)
- To prevent burst behavior (thundering herd):
  - Reduce size and interval proportionally
  - Example: `size=1&interval=50` instead of `size=20&interval=1000`
  - Note: Very high rates with small intervals increase CPU load