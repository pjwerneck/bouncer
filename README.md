# Bouncer

[![Go Report Card](https://goreportcard.com/badge/github.com/pjwerneck/bouncer)](https://goreportcard.com/report/github.com/pjwerneck/bouncer)
[![Build Status](https://github.com/pjwerneck/bouncer/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/pjwerneck/bouncer/actions/workflows/go.yml?branch=master)


![bouncer logo](https://s3.amazonaws.com/www.pedrowerneck.com/images/bouncer-sm.png)


A lightweight RPC service for distributed application control - replaces
makeshift solutions using memcached or Redis.

## Quick Start

Run the server:
```bash
docker run -p 5505:5505 pjwerneck/bouncer:latest
```

View the API documentation:
```bash
# Open Swagger UI in your browser
http://localhost:5505/docs
```

### Examples

#### *"I want to limit something to only one operation per second"*
```bash
# Use a token bucket
curl http://localhost:5505/tokenbucket/myapp/acquire
```

#### *"How to increase the limit to twenty per second?"*
```bash
# Set size=20
curl http://localhost:5505/tokenbucket/myapp/acquire?size=20
```

#### *"But I can't have all twenty starting at the same time!"*
```bash
# Set interval to 1000/rate for smoother distribution
curl http://localhost:5505/tokenbucket/myapp/acquire?interval=50
```

#### *"What if I have a resource that can be used by only one client at a time?"*
```bash
# Use a semaphore
KEY=$(curl http://localhost:5505/semaphore/myapp/acquire)
# ... do work ...
curl http://localhost:5505/semaphore/myapp/release?key=$KEY
```

#### *"Now I need to limit it to ten concurrent clients."*
```bash
# Use a semaphore with size=10
KEY=$(curl http://localhost:5505/semaphore/myapp/acquire?size=10)
# ... do work ...
curl http://localhost:5505/semaphore/myapp/release?key=$KEY
```

#### *"I have some clients that must wait for something else to finish."*
```bash
# Waiting clients
curl http://localhost:5505/event/myapp/wait

# Triggering client
curl http://localhost:5505/event/myapp/send?message=wakeup
```

#### *"I need to count how many times something happens across multiple services"*
```bash
# Increment counter by 1
curl http://localhost:5505/counter/myapp/count

# Increment by specific amount
curl http://localhost:5505/counter/myapp/count?amount=5

# Get current value
curl http://localhost:5505/counter/myapp/value

# Reset counter
curl http://localhost:5505/counter/myapp/reset
```

#### *"I want multiple clients to wait until exactly N of them are ready"*
```bash
# Create a barrier for 3 clients
# Each client does:
curl http://localhost:5505/barrier/myapp/wait?size=3

# The barrier triggers automatically when the 3rd client arrives
```

#### *"I need to know if a periodic task stops running"*
```bash
# Monitoring clients
curl http://localhost:5505/watchdog/myapp/wait

# Task that should be monitored
while true; do
    curl http://localhost:5505/watchdog/myapp/kick?expires=60000
    sleep 30
done

# If the task fails to kick within expires time,
# all waiting clients will be notified
```

## Configuration

Environment variables for customizing server behavior:

| Variable | Default | Description |
|----------|---------|-------------|
| `BOUNCER_HOST` | `0.0.0.0` | Server listen address |
| `BOUNCER_PORT` | `5505` | Server listen port |
| `BOUNCER_LOGLEVEL` | `INFO` | Log level (TRACE, DEBUG,INFO,WARN,ERROR,FATAL,PANIC) |
| `BOUNCER_READ_TIMEOUT` | `30` | HTTP read timeout (seconds) |
| `BOUNCER_WRITE_TIMEOUT` | `30` | HTTP write timeout (seconds) |


> [!WARNING]
> If your controllers require a `maxwait` time that exceeds the
> default write timeout of `30` seconds, increase the `BOUNCER_WRITE_TIMEOUT`
> value accordingly, otherwise the server might close the connection before the
> response is available