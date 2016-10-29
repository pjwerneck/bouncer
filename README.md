
# bouncer

[![Go Report Card](https://goreportcard.com/badge/github.com/pjwerneck/bouncer)](https://goreportcard.com/report/github.com/pjwerneck/bouncer)
[![Build Status](https://travis-ci.org/pjwerneck/bouncer.svg?branch=master)](https://travis-ci.org/pjwerneck/bouncer)


**WARNING: This is alpha code and may not be suitable for production usage.**

This is a simple RPC service to provide throttling, rate-limiting, and synchronization for distributed applications. It's intended as a replacement for makeshift solutions using memcached or Redis.

## Examples

#### *"I want to limit something to only one operation per second"*

You need a token bucket. Just do this before each operation:

    $ curl http://myhost:5505/v1/tokenbucket/myapp/acquire

#### *"How to increase the limit to twenty per second?"*

Use `size=20`:

    $ curl http://myhost:5505/v1/tokenbucket/myapp/acquire?size=20

#### *"But I can't have all twenty starting at the same time!"*

If you don't want bursts of activity, set interval to `1/rate`:

    $ curl http://myhost:5505/v1/tokenbucket/myapp/acquire?interval=50

#### *"What if I have a resource that can be used by only one client at a time?"*

Use a semaphore:

    $ $KEY = curl http://myhost:5505/v1/semaphore/myapp/acquire
    $ # do your stuff
    $ curl http://myhost:5505/v1/semaphore/myapp/release?key=$KEY

#### *"Now I need to limit it to ten concurrent clients."*

Use a semaphore with `size=10`:

    $ $KEY = curl http://myhost:5505/v1/semaphore/myapp/acquire?size=10
    $ # do your stuff
    $ curl http://myhost:5505/v1/semaphore/myapp/release?key=$KEY

#### *"I have some clients that must wait for something else to finish."*

You can use an event:

    $ # do this on any number of waiting clients
    $ curl http://myhost:5505/v1/event/myapp/wait
    $ # and this on the client doing the operation, when it's finished
    $ curl http://myhost:5505/v1/event/myapp/send?message=wakeup


## API

### Requests

All requests use the `GET` method. All arguments are sent using query string parameters. A controller can be easily integrated into anything that can perform an HTTP request.

A request automatically creates a controller if it doesn't exist. If it already exists and the parameters are different, the request updates the controller automatically.

### Parameters

The parameters are normalized for all controllers. It's not an error to pass an unneeded parameter, but it's an error to pass an unknown parameter.

All time values are in milliseconds. All numeric values are integers.

**`expires=[integer]`**

The expiration time for `semaphore` and `watchdog` controllers. Default is `60000`, one minute.

**`interval=[integer]`**

The time interval between `tokenbucket` refills. Default is `1000`, one second.

**`key=[alphanumeric]`**

The unique value used by `semaphore` to release a hold.

**`maxwait=[integer]`**

The max time to wait for a response. The default is `-1`, wait forever. A value of zero never blocks, returning immediately.

**`size=[integer]`**

The size of `tokenbucket` and `semaphore`. The default is `1`. It must be greater than zero.

> Keep in mind that changing the size of a controller affects its current state. For instance, if you reduce or increase the size of a tokenbucket, it will take into account the tokens already acquired in the current interval. If you reduce the size a semaphore, a client won't be able to acquire a hold until the extra clients were released.


### Success Responses

Status codes are very specific so clients can use them to understand the response for valid requests without parsing the response body.

**`200 OK`**

For succesful requests that need to return some value, e.g. `semaphore/acquire`, `event/wait`.

**`204 No Content`**

For successful requests that don't return any value.

**`400 Bad Request`**

An invalid request, e.g., missing a required parameter, invalid value, etc.

**`408 Request Timeout`**

The `maxwait` value was exceeded while waiting for a response.

**`409 Conflict`**

The current state of the controller is incompatible with this request.


## Token Bucket

The `tokenbucket` is an implementation of the Token Bucket algorithm. The bucket has a limited size, and every `interval` the bucket is refilled to capacity with tokens. Each `tokenbucket.get` request takes a token out of the bucket, or waits for a token to be added if the bucket is empty.

### Acquire
***`/v1/tokenbucket/<name>/acquire <size=1> <interval=1000> <maxwait=-1>`***

In most cases you can just set `size` to the desired number of requests per second.

The `size` value must be a positive integer. If you need a fractional ratio of requests per second, you should reduce the fraction and set `size` and `interval` accordingly. Keep in mind that all tokens are added at once, and a naive conversion might result in long wait times. For instance, if you want 10 requests per minute use `size=1&interval=6000`, not `size=10&interval=60000`.

Bursts of activity can happen if there are many clients waiting refill. If that's undesirable, you can reduce `size` and `interval` by a common factor. You can go as far as setting `size=1` and use the `interval` to control the average rate, but in mind that high rates can result in significantly higher CPU usage.

**Responses:**

- `204 No Content`
- `408 Request Timeout`

## Semaphore

A `semaphore` can be used to control concurrent access to shared resources.

### Acquire
***`/v1/semaphore/<name>/acquire <size=1> <key=?> <expires=60000> <maxwait=-1>`***

A semaphore has a number of slots equal to `size`. The `semaphore.acquire` method stores the `key` value in the next available slot. If there are no available slots, the request waits until `maxwait`.

If `key` is not provided, a random UUID is generated.

If `expires` is provided, the hold is automatically released after the given time. You should provide a reasonable value if there's the possibility of a client crashing and never releasing it.

**Responses:**

- `200 OK`, if successful. The `key` value is returned on the response body.
- `408 Request Timeout`, if `maxwait` was exceeded.

### Release
***`/v1/semaphore/<name>/release <key>`***

Releases the previously acquired hold. A release always returns immediately.

**Responses:**

- `204 No Content`, if successful.
- `409 Conflict`, there's no current hold for the given `key`.

## Event

An `event` can be used to synchronize clients, when you want all of them to start doing something immediately after a signal.

### Wait
***`/v1/event/<name>/wait <maxwait=-1>`***

Keeps the client waiting for a signal.

**Responses:**

- `204 No Content`, if the signal was received.
- `408 Request Timeout`, if `maxwait` was exceeded.

### Send
***`/v1/event/<name>/send`***

Sends the signal to all waiting requests.

**Responses:**

- `204 No Content`, if successful
- `409 Conflict`, if already sent.

## Watchdog

A `watchdog` can be used to synchronize clients to do something when a recurring request stops or takes too long.

### Wait
***`/v1/watchdog/<name>/wait <maxwait=-1>`***

Keeps the client waiting for a signal

**Responses:**

- `204 No Content`, if the signal was received.
- `408 Request Timeout`, if `maxwait` was exceeded.

### Kick
***`/v1/watchdog/<name>/kick <expires=60000>`***

Resets the watchdog timer. A signal will be sent to the clients if another `kick` isn't received within the `expires` time.

**Responses:**

- `204 No Content` always.



## High Availability

TODO

## FAQ

TODO
