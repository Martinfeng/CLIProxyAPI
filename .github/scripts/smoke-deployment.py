#!/usr/bin/env python3
"""Exercise disposable CI containers using only synthetic data and credentials."""

import hashlib
import json
import os
import subprocess
import time
from datetime import datetime, timezone
from urllib.error import HTTPError, URLError
from urllib.request import ProxyHandler, Request, build_opener


CPA = "http://127.0.0.1:8317"
MANAGER = "http://127.0.0.1:18317"
OPENER = build_opener(ProxyHandler({}))


def request(url, key, method="GET", payload=None, content_type="application/json"):
    headers = {"Authorization": f"Bearer {key}"}
    data = None
    if payload is not None:
        data = payload if isinstance(payload, bytes) else json.dumps(payload).encode()
        headers["Content-Type"] = content_type
    with OPENER.open(Request(url, data=data, headers=headers, method=method), timeout=10) as response:
        return response.read()


def wait_ready(key):
    deadline = time.monotonic() + 90
    while time.monotonic() < deadline:
        try:
            request(CPA + "/v0/management/config", key)
            request(MANAGER + "/health", key)
            return
        except (HTTPError, URLError, TimeoutError, ConnectionError):
            time.sleep(1)
    raise AssertionError("Official services did not become ready")


def service_info(key):
    return json.loads(request(MANAGER + "/usage-service/info", key))


def exported_events(key):
    data = request(MANAGER + "/v0/management/usage/export", key)
    return [json.loads(line) for line in data.splitlines() if line.strip()]


def main():
    key = os.environ.get("MANAGEMENT_KEY")
    if (
        os.environ.get("GITHUB_ACTIONS") != "true"
        or not os.environ.get("COMPOSE_PROJECT_NAME", "").startswith("cpa-official-ci-")
        or key != "ci-official-images-smoke-only"
    ):
        raise SystemExit("This test is restricted to the disposable GitHub Actions stack")

    wait_ready(key)
    info = service_info(key)
    assert info.get("setupRequired") is True, "Refusing to modify an already-configured Manager"
    request(
        MANAGER + "/setup",
        key,
        method="POST",
        payload={
            "cpaBaseUrl": "http://cli-proxy-api:8317",
            "managementKey": key,
            "collectorMode": "auto",
            "requestMonitoringEnabled": True,
            "ensureUsageStatisticsEnabled": True,
        },
    )
    info = service_info(key)
    assert info.get("configured") is True and info.get("setupRequired") is False, info
    print("PASS: official Manager setup connects to the official CPA management API")

    event_hash = hashlib.sha256(b"official-images-deployment-smoke").hexdigest()
    now = datetime.now(timezone.utc)
    event = {
        "event_hash": event_hash,
        "timestamp_ms": int(now.timestamp() * 1000),
        "timestamp": now.isoformat().replace("+00:00", "Z"),
        "model": "synthetic-model",
        "tokens": {"input_tokens": 3, "output_tokens": 2, "total_tokens": 5},
        "failed": False,
    }
    request(
        MANAGER + "/v0/management/usage/import",
        key,
        method="POST",
        payload=(json.dumps(event) + "\n").encode(),
        content_type="application/x-ndjson",
    )
    assert any(event.get("event_hash") == event_hash for event in exported_events(key))
    print("PASS: a synthetic usage history record was stored and exported")

    subprocess.run(
        ["docker", "compose", "up", "-d", "--force-recreate", "--no-build", "--pull", "never"],
        check=True,
    )
    wait_ready(key)
    info = service_info(key)
    assert info.get("configured") is True and info.get("setupRequired") is False, info
    assert any(event.get("event_hash") == event_hash for event in exported_events(key))
    print("PASS: Manager setup and history survived recreation of both official containers")


if __name__ == "__main__":
    main()
