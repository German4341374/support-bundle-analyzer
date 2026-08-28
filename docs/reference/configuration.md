# Configuration reference

CLI flags override environment variables, which override file configuration, which overrides built-in defaults. Version 0.1 exposes core limits through Go defaults and API settings through environment variables; a unified configuration file is roadmap work.

| Variable | Default | Meaning |
|---|---|---|
| `SBA_HOST` | `127.0.0.1` | API bind address |
| `SBA_PORT` | `8080` | API TCP port |
| `SBA_ALLOW_REMOTE` | `false` | Explicitly permit a non-loopback bind |
| `SBA_ACCESS_TOKEN` | unset | Required bearer token in remote mode; minimum 24 characters |
| `SBA_INPUT_ROOT` | current directory | Root under which API-requested bundle paths must resolve |
| `SBA_WORKSPACE_ROOT` | `.sba/workspaces` | API-managed analysis workspaces |
| `SBA_CORE_BINARY` | `support-bundle-analyzer` | Trusted Go binary path |
| `SBA_ANALYSIS_TIMEOUT_SECONDS` | `900` | Hard process deadline, from 1 to 86400 seconds |
| `SBA_RATE_LIMIT_MAX` | `120` | Maximum requests per client IP per minute |
| `SBA_EXPENSIVE_RATE_LIMIT_MAX` | `10` | Maximum analysis, redaction, or comparison requests per client IP per minute; cannot exceed the general limit |
| `LOG_LEVEL` | `info` | structured API log level |

Do not put access tokens or database passwords in version control. Use `.env.example` only as a list of required names.
