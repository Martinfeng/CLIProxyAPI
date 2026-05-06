# CLI Proxy API (Fork)

本项目是 [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) 的 Fork 版本，整合了 [seakee/CPA-Manager](https://github.com/seakee/CPA-Manager) 作为 Management 面板，提供更完整的使用统计和一键部署体验。

> **推荐**：如果不需要以下增强，建议直接使用 [原项目](https://github.com/router-for-me/CLIProxyAPI)。本 Fork 仅做针对性补充，并通过自动化流程保持与上游同步。

## 与原项目的区别

| 功能 | 原项目 | 本 Fork |
|------|--------|---------|
| 统计数据持久化 | 仅内存（重启丢失） | 内置磁盘 JSON + 32 天保留；可选 SQLite 持久化（Usage Service） |
| Management 面板 | router-for-me/Cli-Proxy-API-Management-Center | [seakee/CPA-Manager](https://github.com/seakee/CPA-Manager)（含 monitoring/usage tab） |
| 上游同步 | - | 每 2 小时自动检查 CPA + CPA-Manager，任一更新即出 fork release |

## 一键部署（推荐）

```bash
curl -O https://raw.githubusercontent.com/Martinfeng/CLIProxyAPI/main/docker-compose.yml
docker compose up -d
```

启动后：

- CPA 主面板：http://localhost:8317/management.html
- Usage Service（SQLite 持久化）：http://localhost:18317/management.html

首次访问进入 cloud deploy 模式：在浏览器里设置 Management 密码、添加 provider、保存即可。所有数据落在 named volume，无需预创建任何主机目录。

## 仅拉镜像

```bash
docker pull kaelsen/cli-proxy-api:latest
```

## 支持架构

| 架构 | 适用设备 |
|------|---------|
| `linux/amd64` | Intel / AMD 处理器 |
| `linux/arm64` | Apple Silicon (M1/M2/M3/M4)、ARM 服务器 |

镜像为多架构构建，Docker 会自动拉取对应平台版本。

## 增强细节

### 内置统计磁盘持久化（默认开启）

- 每 5 分钟自动保存统计数据到磁盘
- 容器启动时自动加载历史数据
- 32 天后自动清理，防止文件无限增长
- 原子写入 + 备份，防止数据损坏

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `USAGE_STATS_PATH` | 统计数据文件路径 | `$AUTH_DIR/usage-stats.json` |
| `USAGE_STATS_AUTOSAVE_INTERVAL` | 自动保存间隔 | `5m` |

### 可选 SQLite 持久化（Usage Service）

`docker-compose.yml` 已包含 `seakee/cpa-manager` 容器作为独立 Usage Service，监听 `:18317`。它通过 CPA 的 `/v0/management/usage-queue` 消费 usage 事件并写入 SQLite，提供 monitoring/usage 视图。两套统计可同时运行，互不冲突。

## 相关链接

- 本项目 GitHub: https://github.com/Martinfeng/CLIProxyAPI
- 原项目 GitHub: https://github.com/router-for-me/CLIProxyAPI
- Management 面板: https://github.com/seakee/CPA-Manager
