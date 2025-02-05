import itertools as it
import multiprocessing
import socket
import statistics
import subprocess
from functools import partial
from pathlib import Path
from time import perf_counter, sleep

import pytest
import requests

PATH = Path(__file__).parent.parent / "bouncer"


def find_free_port():
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.bind(("0.0.0.0", 0))
    portnum = s.getsockname()[1]
    s.close()

    return portnum


@pytest.fixture(scope="session")
def bouncer():
    port = find_free_port()
    process = subprocess.Popen([PATH, "bouncer"], env={"BOUNCER_PORT": str(port)})
    base_url = f"http://localhost:{port}"

    try:
        # Add timeout for server readiness
        start_time = perf_counter()
        while perf_counter() - start_time < 10:  # 10 second timeout
            try:
                rep = requests.get(f"{base_url}/.well-known/ready", timeout=1)
                if rep.status_code == 200:
                    break
                sleep(0.1)
            except (requests.exceptions.ConnectionError, requests.exceptions.Timeout):
                sleep(0.1)
        else:
            raise TimeoutError("Bouncer failed to start within timeout")

        yield base_url
    finally:
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()


def test_bouncer_is_alive(bouncer):
    rep = requests.get(f"{bouncer}/.well-known/ready")
    assert rep.status_code == 200


def tokenbucket_worker(url):
    resp = requests.get(url)
    return resp.status_code, perf_counter()


def test_tokenbucket(bouncer):
    # try making 20 requests without waiting. 10 should succeed immediately
    url = f"{bouncer}/tokenbucket/tb1/acquire?size=10&maxWait=0"
    with multiprocessing.Pool(30) as pool:
        start = perf_counter()
        results = pool.map(tokenbucket_worker, [url] * 20)
        end = perf_counter()

    results = [status for status, _ in results]

    assert results.count(204) == 10
    assert results.count(408) == 10
    # the 10 concurrent requests should complete in 20ms
    assert end - start == pytest.approx(0, abs=0.02)

    # try making 20 requests with waiting. 20 should succeed, but it should take
    # close to 2 seconds since the bucket is empty now
    url = f"{bouncer}/tokenbucket/tb1/acquire?size=10"
    with multiprocessing.Pool(30) as pool:
        start = perf_counter()
        results = pool.map(tokenbucket_worker, [url] * 20)

    status = [status for status, _ in results]
    end = [end for _, end in results]

    assert status.count(204) == 20
    # give 100ms margin for the requests to complete
    assert max(end) - start == pytest.approx(2, abs=0.1)


def test_tokenbucket_under_load(bouncer):
    url = f"{bouncer}/tokenbucket/loadtest1/acquire?size=1000&maxWait=0&interval=1000"

    with multiprocessing.Pool(50) as pool:
        start = perf_counter()
        results = pool.map(tokenbucket_worker, [url] * 2000)

    success_count = sum(1 for status, _ in results if status == 204)
    timeout_count = sum(1 for status, _ in results if status == 408)
    response_times = [end - start for _, end in results]

    # we should get 1000 successful responses
    assert success_count == 1000
    assert timeout_count == 1000

    # with 50 workers and 2000 requests, we should get the 2000 responses within
    # 200ms
    assert statistics.mean(response_times) == pytest.approx(0, abs=0.2)


def test_tokenbucket_refill_under_load(bouncer):
    url = f"{bouncer}/tokenbucket/loadtest2/acquire?size=1000&interval=1000"

    with multiprocessing.Pool(50) as pool:
        start = perf_counter()
        results = pool.map(tokenbucket_worker, [url] * 2000)

    success_count = sum(1 for status, _ in results if status == 204)
    response_times = [end - start for _, end in results]

    # we should get 1000 responses within 1s, and the next 1000 responses within
    # 1s after that
    assert success_count == 2000
    assert len([t for t in response_times if t < 1]) == 1000
    assert len([t for t in response_times if t < 2]) == 2000


def semaphore_worker(url, size=1):
    # track the time the semaphore was held by this process. it should not
    # overlap more than allowed by the semaphore size
    rep = requests.get(f"{url}/acquire?size={size}")
    assert rep.status_code == 200
    start = perf_counter()
    key = rep.text
    sleep(0.1)
    rep = requests.get(f"{url}/release?key={key}")
    assert rep.status_code == 204
    end = perf_counter()
    return start, end


def test_semaphore_size_1_and_5_clients(bouncer):
    # start and end should be close and never overlap
    url = f"{bouncer}/semaphore/s1"

    with multiprocessing.Pool(5) as pool:
        results = pool.map(semaphore_worker, [url] * 5)

    # no results should overlap, since only one client can hold the semaphore
    results.sort()
    for a, b in it.pairwise(results):
        assert b[0] > a[1]

    stats = requests.get(f"{url}/stats").json()
    assert stats["acquired"] == 5
    assert stats["released"] == 5
    assert stats["total_wait_time"] == 0


def test_semaphore_size_10_and_10_clients(bouncer):
    # all 10 should overlap
    url = f"{bouncer}/semaphore/s2"

    with multiprocessing.Pool(10) as pool:
        results = pool.map(partial(semaphore_worker, size=10), [url] * 10)

    # all results should be close to each other, since all 10 clients can hold
    # the semaphore at the same time
    for a, b in it.combinations(results, 2):
        assert b[0] - a[0] == pytest.approx(0, abs=0.01)
        assert b[1] - a[1] == pytest.approx(0, abs=0.01)

    stats = requests.get(f"{url}/stats").json()
    assert stats["acquired"] == 10
    assert stats["released"] == 10
    assert stats["total_wait_time"] == 0


def test_semaphore_size_5_and_6_clients(bouncer):
    url = f"{bouncer}/semaphore/s3"

    with multiprocessing.Pool(6) as pool:
        results = pool.map(partial(semaphore_worker, size=5), [url] * 6)

    results.sort()
    # the first five should overlap, but the last one should not overlap with
    # any of the first five
    for a, b in it.pairwise(results[:5]):
        assert b[0] - a[0] == pytest.approx(0, abs=0.01)
        assert b[1] - a[1] == pytest.approx(0, abs=0.01)

    for other in results[:5]:
        assert results[5][0] > other[1]

    stats = requests.get(f"{url}/stats").json()
    assert stats["acquired"] == 6
    assert stats["released"] == 6

    # FIXME: total wait time should be non-zero
    # assert stats["total_wait_time"] > 0


def test_semaphore_recovery_after_expiration(bouncer):
    url = f"{bouncer}/semaphore/recovery"
    params = {"size": "1", "expires": "100"}

    # Get lock but don't release it
    response = requests.get(f"{url}/acquire", params=params)
    assert response.status_code == 200
    key = response.text

    # Wait for lock to expire
    sleep(0.2)

    # Should be able to acquire again
    response = requests.get(f"{url}/acquire", params=params)
    assert response.status_code == 200

    # releasing with the old key should fail
    response = requests.get(f"{url}/release?key={key}")
    assert response.status_code == 409

    stats = requests.get(f"{url}/stats").json()
    assert stats["acquired"] == 2
    assert stats["released"] == 1
    assert stats["expired"] == 1


def event_worker(url):
    if url.endswith("send"):
        sleep(0.1)
    response = requests.get(url)
    return response.status_code, perf_counter()


def test_event_wait_and_trigger(bouncer):
    url = f"{bouncer}/event/et2"

    # 10 clients should wait for the event, 1 client should trigger it after
    # waiting for 0.1s
    with multiprocessing.Pool(11) as pool:
        results = pool.map(event_worker, [f"{url}/wait"] * 10 + [f"{url}/send"])

    # all 11 clients should get 204
    assert all(status == 204 for status, _ in results)

    # the 10 clients should get the response within 0.01s of the trigger
    trigger = results[-1][1]
    for _, end in results[:-1]:
        assert end - trigger == pytest.approx(0, abs=0.01)


def test_event_wait_timeout(bouncer):
    url = f"{bouncer}/event/et3"

    # 10 clients should wait for the event, but the event should not be triggered
    # so they should all timeout
    with multiprocessing.Pool(10) as pool:
        results = pool.map(event_worker, [f"{url}/wait?maxWait=100"] * 10)

    # all 10 clients should get 408
    assert all(status == 408 for status, _ in results)


def test_event_wait_already_triggered(bouncer):
    url = f"{bouncer}/event/et4"

    # trigger before any clients are waiting
    response = requests.get(f"{url}/send")
    assert response.status_code == 204

    # 10 clients should wait for the event and get 204 immediately
    with multiprocessing.Pool(10) as pool:
        results = pool.map(event_worker, [f"{url}/wait"] * 10)

    # all 10 clients should get 204
    assert all(status == 204 for status, _ in results)

    # the 10 clients should get the response immediately
    trigger = results[-1][1]
    for _, end in results[:-1]:
        assert end - trigger == pytest.approx(0, abs=0.01)


def counter_worker(url):
    response = requests.get(url)
    return response.status_code, response.text


def test_counter_multiple_clients(bouncer):
    url = f"{bouncer}/counter/c1"

    with multiprocessing.Pool(10) as pool:
        results = pool.map(counter_worker, [f"{url}/count"] * 1000)

    # all clients should get 200
    assert all(status == 200 for status, _ in results)

    # all clients should get a different value
    sorted([int(c[1]) for c in results]) == list(range(1, 1001))

    # the counter should be 1000
    response = requests.get(f"{url}/value")
    assert response.status_code == 200
    assert response.text == "1000"

    # reset the counter
    response = requests.get(f"{url}/reset")
    assert response.status_code == 204

    # the counter should be 0
    response = requests.get(f"{url}/value")
    assert response.status_code == 200
    assert response.text == "0"


def watchdog_worker(url):
    response = requests.get(url)
    return response.status_code, response.text


def test_watchdog_no_kicks(bouncer):
    url = f"{bouncer}/watchdog/wd1"

    with multiprocessing.Pool(10) as pool:
        results = pool.map(watchdog_worker, [f"{url}/wait?maxWait=100"] * 10)

    # all clients should get 408
    assert all(status == 408 for status, _ in results)


def test_watchdog_with_single_kick(bouncer):
    url = f"{bouncer}/watchdog/wd2"

    with multiprocessing.Pool(11) as pool:
        results = pool.map(
            watchdog_worker,
            [f"{url}/wait?maxWait=1000"] * 10 + [f"{url}/kick?expires=100"],
        )

    # all clients should get 204
    assert all(status == 204 for status, _ in results)


def barrier_worker(url):
    response = requests.get(url)
    return response.status_code, perf_counter()


def test_barrier_timeout(bouncer):
    url = f"{bouncer}/barrier/b1"

    # with size=10 but only 9 clients, all clients should timeout
    with multiprocessing.Pool(9) as pool:
        start = perf_counter()
        results = pool.map(barrier_worker, [f"{url}/wait?size=10&maxWait=100"] * 5)

    # all clients should get 408
    assert all(status == 408 for status, _ in results)

    for _, end in results:
        assert end - start == pytest.approx(0.1, abs=0.02)


def test_barrier_success(bouncer):
    url = f"{bouncer}/barrier/b2"

    # start 9 clients
    clients = []
    for _ in range(9):
        clients.append(
            multiprocessing.Process(
                target=barrier_worker, args=(f"{url}/wait?size=10",)
            )
        )
        clients[-1].start()

    # wait for 0.1s
    sleep(0.1)

    # start the 10th client
    start = perf_counter()
    response = requests.get(f"{url}/wait?size=10")
    end = perf_counter()

    # the 10th client should get 204
    assert response.status_code == 204

    # all clients should complete within 0.1s of each other
    for client in clients:
        client.join()
        assert client.exitcode == 0

    for client in clients:
        assert end - start == pytest.approx(0, abs=0.1)

    # barrier is fully reusable afterwards
    with multiprocessing.Pool(10) as pool:
        results = pool.map(barrier_worker, [f"{url}/wait?size=10"] * 10)

    # all clients should get 204
    assert all(status == 204 for status, _ in results)


def test_error_cases(bouncer):
    cases = [
        (f"{bouncer}/tokenbucket/error/acquire?size=-1", 400),
        # (f"{bouncer}/tokenbucket/error/acquire?interval=-1", 400),
        (f"{bouncer}/semaphore/error/release?key=invalid", 409),
    ]

    for url, expected_status in cases:
        response = requests.get(url)
        assert response.status_code == expected_status, url
