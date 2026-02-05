# CLI Proxy API (Fork)

Fork of [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) with enhancements.

## Source Code

**GitHub:** [https://github.com/Martinfeng/CLIProxyAPI](https://github.com/Martinfeng/CLIProxyAPI)

## Fork Enhancements

- **Usage Statistics Persistence** - Automatically saves usage statistics to disk, surviving container restarts
- **32-Day Data Retention** - Automatic cleanup of statistics older than 32 days

## Quick Start

```bash
docker pull kaelsen/cli-proxy-api:latest
```

## Docker Compose

```yaml
services:
  cli-proxy-api:
    image: kaelsen/cli-proxy-api:latest
    ports:
      - "8317:8317"
    volumes:
      - ./config.yaml:/CLIProxyAPI/config.yaml
      - ./auths:/root/.cli-proxy-api
      - ./logs:/CLIProxyAPI/logs
    restart: unless-stopped
```

## Supported Architectures

| Architecture | Tag |
|--------------|-----|
| linux/amd64 | `latest` |
| linux/arm64 | `latest` |

Both Intel/AMD and Apple Silicon (M1/M2/M3/M4) are supported.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `USAGE_STATS_PATH` | Path to usage statistics file | `$AUTH_DIR/usage-stats.json` |
| `USAGE_STATS_AUTOSAVE_INTERVAL` | Auto-save interval | `5m` |

## Links

- [GitHub Repository](https://github.com/Martinfeng/CLIProxyAPI)
- [Upstream Project](https://github.com/router-for-me/CLIProxyAPI)
