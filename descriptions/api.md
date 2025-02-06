A lightweight RPC service for distributed application control. Provides primitives for rate limiting, resource synchronization, and process coordination.

### General Concepts
- Endpoints use GET method with query parameters
- Clients block until the operation is completed or `maxwait` is reached
- Resources are created automatically on first use
- All time values are in milliseconds
- All numeric parameters are integers

### Quick Tips
- Test endpoints easily with `curl`, `ab` or your browser
- Use `maxwait=0` to test resource availability without blocking
- Monitor resource usage with the `/stats` endpoints
- Check server readiness at `/.well-known/ready`

### Status Codes
- `204 No Content`: Operation completed successfully
- `200 OK`: Operation completed with data returned
- `408 Request Timeout`: The `maxwait` time was exceeded
- `409 Conflict`: Operation conflicts with current state


