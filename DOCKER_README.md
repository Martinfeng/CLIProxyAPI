# CLI Proxy API (Fork)

本项目是 [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) 的 Fork 版本，整合 [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) 作为 Manager Server，提供 SQLite 持久化的请求统计、模型价格、Codex 巡检等完整管理能力。

> 不需要以下增强，建议直接使用 [原项目](https://github.com/router-for-me/CLIProxyAPI)。本 Fork 仅做针对性补充，并通过自动化流程保持与上游同步。

## 与原项目的区别

| 功能 | 原项目 | 本 Fork |
|------|--------|---------|
| 请求统计持久化 | 仅内存（重启丢失） | 内置磁盘 JSON + 32 天保留；可选 SQLite（CPA-Manager-Plus） |
| Management 面板 | router-for-me/Cli-Proxy-API-Management-Center | [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) |
| 监控/模型价格/Codex 巡检 | 无 | Plus Manager Server 提供（`:18317` 入口） |
| 上游同步 | - | 每 2 小时自动检查 CPA + Plus，任一更新即出 fork release |

## 一键部署

```bash
curl -O https://raw.githubusercontent.com/Martinfeng/CLIProxyAPI/main/docker-compose.yml
curl -o config.yaml https://raw.githubusercontent.com/Martinfeng/CLIProxyAPI/main/config.example.yaml
curl -o .env https://raw.githubusercontent.com/Martinfeng/CLIProxyAPI/main/.env.example
# 编辑 .env，把 MANAGEMENT_KEY 换成你自己的密码（openssl rand -hex 24 之类）
# 编辑 config.yaml，把 api-keys 段那三个 your-api-key-X 占位删了或换成真 key
docker compose up -d
```

升级：`docker compose pull && docker compose up -d`

> **一个密码到底（v7.1.50-fork.1 起）**：`.env` 里的 `MANAGEMENT_KEY` 同时注入到 CPA 的 `MANAGEMENT_PASSWORD` 和 Plus 的 `CPA_MANAGER_ADMIN_KEY`。Plus setup 页问 admin key 和 CPA management key 时**两次填同一个值**就行；`config.yaml` 里 `remote-management.secret-key` 保持空字符串。没设 `MANAGEMENT_KEY` `docker compose up` 会直接报错不启动（避免裸跑无密码）。

> **从旧版升级（v7.1.50-fork.1 之前装过的看这里）**：服务名从 `cpa-manager` 改成 `cpa-manager-plus`，`docker compose down` 不会清掉旧容器，新容器会撞 `:18317 port is already allocated`。先跑：
>
> ```bash
> docker stop cpa-manager && docker rm cpa-manager
> docker volume rm cliproxyapi_cpa-manager-data 2>/dev/null || true
> docker compose pull && docker compose up -d
> ```

> **首次部署后还要在浏览器走一次 setup**：详细步骤见每次 [GitHub Release 说明](https://github.com/Martinfeng/CLIProxyAPI/releases/latest)，简要流程：
>
> 1. 浏览器进 `http://localhost:18317/management.html`
> 2. setup 页填：Manager 管理员密钥 = `MANAGEMENT_KEY`；CPA URL = `http://cli-proxy-api:8317`；CPA Management Key = **同一个** `MANAGEMENT_KEY`
> 3. 之后每次都用 `MANAGEMENT_KEY` 登录

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

### Manager Server（CPA-Manager-Plus，docker-compose 内置）

`docker-compose.yml` 已包含 `seakee/cpa-manager-plus` 容器，监听 `:18317`。它通过 CPA 的 `/v0/management/usage-queue` 消费 usage 事件并写入 SQLite，提供监控、模型价格估算、Codex 服务端巡检等能力。CPA 自带的磁盘 JSON 持久化与之并行运行，互不冲突。

`:8317/management.html` 入口仍可用作纯 CPA 管理面板，但 Plus 设计下监控/价格/巡检在该入口被故意隐藏，完整能力需要走 `:18317`。

## 相关链接

- 本项目 GitHub: https://github.com/Martinfeng/CLIProxyAPI
- 原项目 GitHub: https://github.com/router-for-me/CLIProxyAPI
- Manager Server: https://github.com/seakee/CPA-Manager-Plus
