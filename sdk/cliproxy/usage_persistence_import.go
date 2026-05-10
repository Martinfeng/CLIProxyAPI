// Fork-only：把 internal/usage 的 init() 注册（统计持久化 hook）从 service.go 拔出来，
// 单独放在这个文件里，避免每次 upstream 改 service.go 的 import 段都触发 content 冲突。
package cliproxy

import _ "github.com/router-for-me/CLIProxyAPI/v7/internal/usage"
