// Register the fork's usage persistence hook separately so upstream changes to
// service.go imports do not conflict with the fork-only registration.
package cliproxy

import _ "github.com/router-for-me/CLIProxyAPI/v8/internal/usage"
