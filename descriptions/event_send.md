Triggers an event, sending a message to all waiting clients.

### Basic Operation
- Sends optional message to all waiting clients
- Returns immediately
- Can only be triggered once
- Already triggered events return 409 Conflict

### Usage Tips
- Use meaningful messages for easier debugging
- Empty message is allowed but not recommended
- All waiting clients receive the same message
