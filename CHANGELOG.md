# Changelog

All notable changes to this project are documented in this file.

## Unreleased

## 0.1.4 - 2026-09-24

### Fixed

- Correct the Redis startup documentation: use `BootUpRedisContext(ctx, configs, log)` and pass `nil` as the logger to disable command logging.

## 0.1.3 - 2026-09-24

### Added

- Add an optional Redis command logging hook through `BootUpRedisContext`.
- Add the `STANDALONE` Redis mode constant.

### Security

- Redis command logs include only command names and safely identified keys; values and complete argument lists are excluded.

### Changed

- Successful Redis commands and pipelines log at Info level; failures log at Error level with the shared logger caller and stacktrace behavior.
- Redis protocol-negotiation and capability-probe commands are excluded from command logs because older servers may reject them during a successful driver fallback.

## 0.1.2 - 2026-09-21

### Fixed

- Avoid unsupported console stderr synchronization errors during shutdown on Windows while retaining file-log synchronization errors.

## 0.1.1 - 2026-09-07

### Changed

- Suppress the MongoDB driver's successful `endSessions` command log during shutdown.

## 0.1.0 - 2026-09-07

### Added

- Context-aware startup and operation APIs for MySQL, Redis, and MongoDB.
- Atomic named-resource registration with concurrent lifecycle protection.
- Configurable SQL, Redis, and MongoDB connection-pool settings.
- Redis custom commands, pipelines, scripts, and common data-structure commands.
- MongoDB command execution, default database support, and arbitrary insert ID support.
- Structured error stacks and configurable logger caller depth.
- API compatibility, race, reconnect, lifecycle, validation, and integration tests.

### Changed

- Configuration validation now fails before publishing partially initialized resources.
- Shutdown operations unregister resources and report aggregated close errors.
- XORM `MaxLife` follows its documented unit of seconds.
- Legacy context-free APIs are deprecated but remain available.
- Go source comments and public error messages use English.

### Security

- Request query strings, MongoDB payloads, and connection URIs are excluded from default logs.
