# 官方 CPA + CPA-Manager-Plus 部署

[English](README.md)

本仓库现在只部署两个未经修改的官方镜像：

| 服务 | 官方镜像 | 职责 |
|---|---|---|
| CLIProxyAPI | eceasy/cli-proxy-api | API 代理、认证、管理 API、文件日志 |
| CPA-Manager-Plus | seakee/cpa-manager-plus | 管理界面、请求历史、用量/费用分析、SQLite 持久化 |

原来为统计持久化维护的 CPA 源码分叉已退役。仓库只保留部署配置、说明和验证脚本，不再合并上游源码、编译 Go、构建镜像或发布 fork 版本。

## 全新部署

把 docker-compose.yml、config.example.yaml、.env.example 放到 Docker 部署目录。源码和 Git 操作留在 Guest VM，Docker 操作在宿主机进行。

~~~sh
# [Host via SSH] Docker 宿主机，部署目录，仅限全新安装。
cp .env.example .env
cp config.example.yaml config.yaml
~~~

编辑 .env，设置足够强的 MANAGEMENT_KEY；编辑 config.yaml，把 access.api-keys 下的示例值换成真实客户端 API Key。管理密码和客户端 API Key 用途不同。management.secret-key 保持空字符串，Compose 会注入 MANAGEMENT_PASSWORD。

~~~sh
# [Host via SSH] Docker 宿主机，部署目录。
docker compose config --quiet
docker compose config --images
docker compose pull
docker compose up -d --no-build
~~~

打开 http://localhost:18317/management.html 完成 Plus 初始化：

| 字段 | 填写内容 |
|---|---|
| Manager 管理员密钥 | .env 中的 MANAGEMENT_KEY |
| CPA 地址 | http://cli-proxy-api:8317 |
| CPA Management Key | 同一个 MANAGEMENT_KEY |
| 请求监控 | 开启 |
| 采集模式 | auto |

监控、历史记录统一从 18317 入口查看。8317 提供 CPA API 和嵌入式管理面板。按原有方式配置或导入模型服务凭据，认证文件继续保存在 auths 挂载目录。

## 数据保存位置

| 数据 | 持久化位置 |
|---|---|
| CPA 配置 | ./config.yaml |
| CPA 认证文件 | ./auths |
| CPA 程序日志、请求/响应文件日志 | ./logs |
| CPA 插件 | ./plugins |
| Plus 数据库、加密密钥及其他状态 | cpa-manager-plus-data 卷，挂载到 /data |

Plus 卷里包括 usage.sqlite 和 data.key，两者都要保留。升级时保持原部署目录、Compose project name、服务名和卷标识不变；换 project name 可能会选中一个新的空卷。

新模板默认开启用量采集和程序日志写盘。完整请求/响应日志可通过 observability.logs.request-log 按需开启。已经采集入库的历史可跨容器重启、重建保留；CPA 内存队列中尚未采集的事件仍可能因过期或 CPA 重启而丢失。

原 fork 的 usage-stats.json 快照及 /usage-stats 接口停止维护。旧 JSON 文件保留作备份，不会自动导入 Plus；Plus 使用自己的事件历史和 JSONL 导入/导出格式。

## 从现有 fork 部署迁移

保留真实 .env、config.yaml、认证文件、日志、插件及 Plus 数据。不要用全新安装模板覆盖它们。

1. 切换镜像前备份。下面适用于默认挂载路径：

~~~sh
# [Host via SSH] Docker 宿主机，原有部署目录。
umask 077
backup_dir=$(mktemp -d ../cpa-backup.XXXXXX)
docker compose stop
docker cp cpa-manager-plus:/data "$backup_dir/manager-data"
for path in config.yaml .env auths logs plugins; do
  if [ -e "$path" ]; then cp -a "$path" "$backup_dir/"; fi
done
printf 'Backup directory: %s\n' "$backup_dir"
~~~

如果设置过 CLI_PROXY_*_PATH，请改为备份实际挂载路径。停止 Plus 后复制整个 /data，可一并保留一致的 SQLite 数据库、可能存在的 WAL 文件和加密密钥。备份目录包含敏感配置，应妥善保存。

2. 只更新部署目录中的 docker-compose.yml。检查旧 .env：若设置了旧镜像 CLI_PROXY_IMAGE，改为 eceasy/cli-proxy-api:latest；CPA_MANAGER_IMAGE 使用 seakee/cpa-manager-plus:latest。保留 MANAGEMENT_KEY、路径覆盖项和 COMPOSE_PROJECT_NAME。
3. 确认解析出的镜像都是官方镜像，然后启动：

~~~sh
# [Host via SSH] Docker 宿主机，保持原部署目录和 project name。
docker compose config --quiet
docker compose config --images
docker compose pull
docker compose up -d --no-build
~~~

4. 从 18317 入口确认登录、原有历史、新请求采集正常。已有 Plus 初始化配置应继续有效；若突然重新进入 setup，先核对 project name、数据卷挂载和 data.key，不要直接新建一份配置。

迁移不需要删除数据卷。保留已有卷，不运行清卷命令。历史 fork 镜像 v8.0.3-fork.1 和 Git 历史仍可用于回退，本次改造不删除历史 release 或镜像。

## 后续更新与回退

默认跟随两个官方 latest 标签；更新只需拉取并重建容器：

~~~sh
# [Host via SSH] Docker 宿主机，部署目录。
docker compose pull
docker compose up -d --no-build
~~~

若要控制升级节奏，在 .env 中用 CLI_PROXY_IMAGE、CPA_MANAGER_IMAGE 固定官方版本标签。首次迁移以 CPA v8.0.3、Plus v1.14.1 为核对基线。跨大版本升级前阅读官方说明，并保留数据库和配置备份供回退。

这种方式移除了 fork 合并冲突和本仓库编译失败的来源。官方镜像可用性、软件自身问题和版本兼容性仍由上游版本决定，不能承诺以后任何更新都不会出错。

CI 只验证部署：检查 Compose 和官方镜像来源、让 Plus setup 连接 CPA，再确认合成历史记录和连接配置在两个容器重建后仍存在。不使用真实模型凭据或外部模型调用；历史测试验证的是导入与存储，不是完整模型请求采集链路。

## 官方项目

- [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI)
- [CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus)
- [CPA 完整配置参考](https://github.com/router-for-me/CLIProxyAPI/blob/main/config.example.yaml)
- [Plus 部署和备份说明](https://github.com/seakee/CPA-Manager-Plus/blob/main/README_CN.md)
