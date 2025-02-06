A `semaphore` can be used to control concurrent access to shared resources.

### Basic Operation
- Up to `size` locks can be held simultaneously
- Each acquire returns a unique release key
- Waits up to `maxwait` milliseconds for an available lock
- If `maxwait` is negative, waits indefinitely
- If `maxwait` is 0, returns immediately

### Usage Tips
- Locks expire automatically after `expires` milliseconds
- Set reasonable `expires` time to prevent orphaned locks
- Use `size>1` for resource pools
