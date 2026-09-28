#!/usr/bin/env python3
"""Validate the rendered deployment without exposing its environment values."""

import json
import sys
from pathlib import Path


def require(condition, message):
    if not condition:
        raise SystemExit(message)


def main():
    config = json.load(sys.stdin)
    services = config["services"]
    images = {
        "cli-proxy-api": "eceasy/cli-proxy-api",
        "cpa-manager-plus": "seakee/cpa-manager-plus",
    }
    require(set(services) == set(images), "Exactly the two official services are required")
    for name, repository in images.items():
        service = services[name]
        image_repository = service["image"].split("@", 1)[0].split(":", 1)[0]
        require(image_repository == repository, f"{name} must use its official image")
        require("build" not in service, f"{name} must not build a local image")
        require(service["container_name"] == name, f"{name} container name changed")
        require(service["restart"] == "unless-stopped", f"{name} restart policy changed")

    cpa = services["cli-proxy-api"]
    manager = services["cpa-manager-plus"]
    cpa_mounts = {mount["target"]: mount for mount in cpa["volumes"]}
    for target in (
        "/CLIProxyAPI/config.yaml",
        "/root/.cli-proxy-api",
        "/CLIProxyAPI/logs",
        "/CLIProxyAPI/plugins",
    ):
        require(cpa_mounts.get(target, {}).get("type") == "bind", f"Missing bind: {target}")

    manager_mounts = {mount["target"]: mount for mount in manager["volumes"]}
    data_mount = manager_mounts.get("/data", {})
    require(data_mount.get("type") == "volume", "Manager data must use a persistent volume")
    require(data_mount.get("source") == "cpa-manager-plus-data", "Existing data volume key changed")
    environment = manager["environment"]
    require(environment.get("USAGE_DB_PATH") == "/data/usage.sqlite", "Database path changed")
    require(environment.get("CPA_MANAGER_DATA_KEY_PATH") == "/data/data.key", "Data key path changed")
    require(
        bool(environment.get("CPA_MANAGER_ADMIN_KEY"))
        and environment["CPA_MANAGER_ADMIN_KEY"] == cpa["environment"].get("MANAGEMENT_PASSWORD"),
        "A shared nonempty management key is required",
    )
    for path in ("go.mod", "go.sum", "Dockerfile", "cmd", "internal", "sdk"):
        require(not Path(path).exists(), f"Application source/build path remains: {path}")
    print("Deployment validation passed: official images, stable mounts, and no application build.")


if __name__ == "__main__":
    main()
