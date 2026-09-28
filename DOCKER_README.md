# 旧 CPA fork 镜像已停止更新

本仓库已改为直接部署两个官方镜像：

- CPA：eceasy/cli-proxy-api
- Plus Manager：seakee/cpa-manager-plus

不再构建或发布 kaelsen/cli-proxy-api。历史镜像保留供回退。

请使用当前 [docker-compose.yml](https://github.com/Martinfeng/CLIProxyAPI/blob/main/docker-compose.yml)，并按 [迁移说明](https://github.com/Martinfeng/CLIProxyAPI/blob/main/README_CN.md) 备份后切换。

保留原配置、认证文件、日志、插件和 cpa-manager-plus-data 卷；检查旧 .env 中的镜像覆盖项，避免继续使用旧 fork 镜像。Plus 的 usage.sqlite 和 data.key 需要一并保留。
