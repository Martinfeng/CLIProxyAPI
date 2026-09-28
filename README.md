# Official CPA + CPA-Manager-Plus deployment

[简体中文](README_CN.md)

This repository deploys two unmodified official images:

| Service | Official image | Responsibility |
|---|---|---|
| CLIProxyAPI | eceasy/cli-proxy-api | API proxying, credentials, management API, and file logs |
| CPA-Manager-Plus | seakee/cpa-manager-plus | Management UI, request history, usage/cost analysis, and SQLite persistence |

This repository previously maintained a CPA fork with usage persistence. It now contains only deployment configuration, documentation, and deployment checks. It no longer merges upstream source, compiles Go, builds images, or publishes fork releases.

## New installation

Place docker-compose.yml, config.example.yaml, and .env.example in your Docker deployment directory. Keep Git/source work in the Guest VM; run Docker on its host.

~~~sh
# [Host via SSH] Docker host, deployment directory, NEW installations only.
cp .env.example .env
cp config.example.yaml config.yaml
~~~

Edit .env and set MANAGEMENT_KEY to a strong random password. In config.yaml, replace the example access.api-keys entry with a real client API key. The management password and client API key serve different purposes. Keep management.secret-key empty; Compose supplies MANAGEMENT_PASSWORD.

~~~sh
# [Host via SSH] Docker host, deployment directory.
docker compose config --quiet
docker compose config --images
docker compose pull
docker compose up -d --no-build
~~~

Open http://localhost:18317/management.html and complete Plus setup:

| Field | Value |
|---|---|
| Manager admin key | MANAGEMENT_KEY from .env |
| CPA URL | http://cli-proxy-api:8317 |
| CPA Management Key | The same MANAGEMENT_KEY |
| Request monitoring | Enabled |
| Collector mode | auto |

Use port 18317 for monitoring and history. Port 8317 serves the CPA API and its embedded management panel. Import or configure provider credentials as usual; they remain under the existing auths mount.

## Storage

| Data | Persistent location |
|---|---|
| CPA configuration | ./config.yaml |
| CPA credentials | ./auths |
| CPA application/request log files | ./logs |
| CPA plugins | ./plugins |
| Plus SQLite database, encryption key, and other state | cpa-manager-plus-data volume mounted at /data |

The named volume contains usage.sqlite and data.key. Preserve both. Keep the same deployment directory, Compose project name, service names, and volume key during upgrades; changing the project name can select a different empty volume.

The template enables usage collection and application file logging. Full request/response logging is optional through observability.logs.request-log. Already-ingested history survives container restarts and recreation. Events still in CPA's in-memory telemetry queue can be lost on expiry or a CPA restart before Plus collects them.

The former fork's usage-stats.json snapshots and /usage-stats APIs are retired. Old JSON files remain available for backup, but are not automatically imported into Plus. Plus uses its own event history and JSONL import/export format.

## Migrate an existing fork deployment

Keep your real .env, config.yaml, credentials, logs, plugins, and Plus data. Do not replace them with the new-install templates.

1. Before changing images, back up the existing deployment. For default bind paths:

~~~sh
# [Host via SSH] Docker host, EXISTING deployment directory.
umask 077
backup_dir=$(mktemp -d ../cpa-backup.XXXXXX)
docker compose stop
docker cp cpa-manager-plus:/data "$backup_dir/manager-data"
for path in config.yaml .env auths logs plugins; do
  if [ -e "$path" ]; then cp -a "$path" "$backup_dir/"; fi
done
printf 'Backup directory: %s\n' "$backup_dir"
~~~

If CLI_PROXY_*_PATH overrides point elsewhere, back up those actual paths instead. Copying /data while Plus is stopped preserves a consistent SQLite database, any WAL files, and its encryption key. Keep the backup private.

2. Replace only docker-compose.yml with this repository's current file. In your existing .env, change any old CLI_PROXY_IMAGE override to eceasy/cli-proxy-api:latest and CPA_MANAGER_IMAGE to seakee/cpa-manager-plus:latest. Preserve MANAGEMENT_KEY, path overrides, and COMPOSE_PROJECT_NAME.
3. Check the resolved images, then start the two official containers:

~~~sh
# [Host via SSH] Docker host, SAME deployment directory/project.
docker compose config --quiet
docker compose config --images
docker compose pull
docker compose up -d --no-build
~~~

4. Verify Plus login, existing history, and newly collected requests at port 18317. Existing Plus setup should remain configured. If it unexpectedly asks for setup again, check the project name, data volume mount, and data.key before creating a new configuration.

Do not delete the existing volume or run volume-removal commands during this migration. The historical fork release v8.0.3-fork.1 and Git history remain available for rollback; no old release or image is deleted by this change.

## Updates and rollback

By default, both services use their official latest tags. Updating requires pulling and recreating the containers; this repository performs no builds:

~~~sh
# [Host via SSH] Docker host, deployment directory.
docker compose pull
docker compose up -d --no-build
~~~

For controlled upgrades, pin official version tags with CLI_PROXY_IMAGE and CPA_MANAGER_IMAGE in .env. The initial migration is checked with CPA v8.0.3 and Plus v1.14.1. Review their release notes before changing major versions and keep a database/configuration backup for rollback. Official image availability, upstream behavior, and compatibility can still change; the fork merge/compilation failure path has been removed.

CI validates Compose and official image references, connects Plus setup to CPA, and verifies that a synthetic usage record and the saved connection survive recreation of both containers. It uses no real provider credentials or model calls. The history test exercises import/storage, not end-to-end model request collection.

## Upstream projects

- [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI)
- [CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus)
- [Full CPA configuration reference](https://github.com/router-for-me/CLIProxyAPI/blob/main/config.example.yaml)
- [Plus deployment and backup documentation](https://github.com/seakee/CPA-Manager-Plus/blob/main/README_CN.md)
