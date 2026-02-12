# CLI Proxy API (Fork)

本项目是 [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) 的 Fork 版本，在原项目基础上增加了少量实用增强功能。

> **推荐**：如果你不需要以下增强功能，建议直接使用 [原项目](https://github.com/router-for-me/CLIProxyAPI)。原项目维护更活跃，功能更完整。本 Fork 仅做针对性补充，并通过自动化流程保持与上游同步。

## 与原项目的区别

| 功能 | 原项目 | 本 Fork |
|------|--------|---------|
| 统计数据持久化 | 仅内存存储，容器重启后丢失 | 自动保存到磁盘，重启后恢复 |
| 数据保留策略 | 无限累积 | 自动清理 32 天前的数据 |
| 上游同步 | - | 每 2 小时自动检查并合并上游更新 |

## 快速开始

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

## 支持架构

| 架构 | 适用设备 |
|------|---------|
| `linux/amd64` | Intel / AMD 处理器 |
| `linux/arm64` | Apple Silicon (M1/M2/M3/M4)、ARM 服务器 |

镜像为多架构构建，Docker 会自动拉取对应平台的版本。

## 增强功能说明

### 统计数据持久化

原项目的使用统计仅存于内存，容器重启即丢失。本 Fork 增加了自动持久化机制：

- 每 5 分钟自动保存统计数据到磁盘
- 容器启动时自动加载历史数据
- 支持原子写入和备份，防止数据损坏

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `USAGE_STATS_PATH` | 统计数据文件路径 | `$AUTH_DIR/usage-stats.json` |
| `USAGE_STATS_AUTOSAVE_INTERVAL` | 自动保存间隔 | `5m` |

### 32 天数据保留

统计看板最多显示近 30 天的指标，超过 32 天的数据会被自动清理，防止文件无限增长。

## 相关链接

- [本项目 GitHub](https://github.com/Martinfeng/CLIProxyAPI)
- [原项目 GitHub](https://github.com/router-for-me/CLIProxyAPI)
- [原项目 Docker Hub](https://hub.docker.com/r/eceasy/cli-proxy-api)
