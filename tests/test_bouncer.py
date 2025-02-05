import itertools as it
import multiprocessing
import socket
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
    # get free port
    port = find_free_port()

    # start the bouncer server
    bouncer = subprocess.Popen([PATH, "bouncer"], env={"BOUNCER_PORT": str(port)})

    base_url = f"http://localhost:{port}"

    # wait for the server to start
    for _ in range(100):
        try:
            rep = requests.get(f"{base_url}/.well-known/ready")
            if rep.status_code == 200:
                break
            sleep(0.1)

        except requests.exceptions.ConnectionError:
            sleep(0.1)

    yield base_url

    # stop the bouncer server
    bouncer.kill()


def test_bouncer_is_alive(bouncer):
    rep = requests.get(f"{bouncer}/.well-known/ready")
    assert rep.status_code == 200


def get_status(url):
    return requests.get(url).status_code


def test_tokenbucket(bouncer):
    # try making 20 requests without waiting. 10 should succeed immediately
    url = f"{bouncer}/tokenbucket/tb1/acquire?size=10&maxWait=0"
    with multiprocessing.Pool(30) as pool:
        start = perf_counter()
        results = pool.map(get_status, [url] * 20)
        end = perf_counter()

    assert results.count(204) == 10
    assert results.count(408) == 10
    # the 10 concurrent requests should complete in 20ms
    assert end - start == pytest.approx(0, abs=0.02)

    # try making 20 requests with waiting. 20 should succeed, but it should take
    # close to 2 seconds since the bucket is empty now
    url = f"{bouncer}/tokenbucket/tb1/acquire?size=10"
    with multiprocessing.Pool(30) as pool:
        start = perf_counter()
        results = pool.map(get_status, [url] * 20)
        end = perf_counter()

    assert results.count(204) == 20
    # give 100ms margin for the requests to complete
    assert end - start == pytest.approx(2, abs=0.1)


def get_semaphore(url, size=1):
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
        results = pool.map(get_semaphore, [url] * 5)

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
        results = pool.map(partial(get_semaphore, size=10), [url] * 10)

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
        results = pool.map(partial(get_semaphore, size=5), [url] * 6)

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
