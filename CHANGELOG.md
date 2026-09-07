# Changelog

All notable changes to this project are documented in this file.

## Unreleased

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
