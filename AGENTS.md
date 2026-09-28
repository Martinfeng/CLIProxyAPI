# Deployment repository

This repository deploys official CLIProxyAPI and CPA-Manager-Plus images.
It does not contain or build either application's source code.

- Keep the default image repositories eceasy/cli-proxy-api and seakee/cpa-manager-plus.
- Do not reintroduce Go sources, Dockerfiles, upstream merge automation, or image/release publishing.
- Preserve the existing service names, ports, bind mounts, and cpa-manager-plus-data volume.
- Never overwrite real config.yaml, .env, auths, logs, plugins, or data volumes.
- Keep the Compose project name stable during migration so existing volumes are reused.
- Runtime configuration belongs in templates and documentation, not application patches.
- Use English comments and update both README.md and README_CN.md for deployment changes.
- Docker is unavailable in the Guest VM. Use GitHub Actions or the approved isolated Host Docker runbook for container validation.

Validation:

1. Parse the templates and run python3 -m py_compile on changed Python scripts.
2. Run docker compose config --quiet with a temporary test configuration.
3. Pass the rendered Compose JSON to .github/scripts/validate-deployment.py.
4. Run .github/scripts/smoke-deployment.py against disposable containers with test credentials only.

The smoke test checks official image startup, CPA/Manager setup connectivity,
and survival of a synthetic Manager history record across a container restart.
It must never run against a real deployment.
