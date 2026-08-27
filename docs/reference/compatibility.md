# Compatibility matrix

| Component | Supported baseline | Notes |
|---|---|---|
| Go CLI | Go 1.26+ | Windows, Linux and macOS source builds |
| Node API | Node.js 24+ | strict ESM TypeScript build |
| Python plugin | Python 3.12+ | runtime uses standard library only |
| Windows plugin | .NET 8 LTS | Event XML works cross-platform; Windows exports are the input |
| JVM plugin | Java 17+ | release target 17; build verified with Maven |
| PHP plugin | PHP 8.3/8.4 | Composer development tools pinned in lockfile |
| PostgreSQL | 17 | optional persistent server mode |
| Plugin protocol | v1 | incompatible versions must be skipped with a warning |

Archive support: ZIP, TAR, TAR.GZ/TGZ, TAR.BZ2, TAR.XZ and single-file GZIP. Nested archives are classified but not recursively extracted.
