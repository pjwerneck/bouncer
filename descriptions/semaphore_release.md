Releases a previously acquired `semaphore` lock.

### Basic Operation
- Release using the key returned by acquire
- Returns immediately
- Invalid or already released keys return 409 Conflict
