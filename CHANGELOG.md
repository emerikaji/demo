# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `user` routes.

### Changed

-  Getting tickets from the datastore now works similarly to events with a filter struct.

## [0.1.0] - 28/09/2026

### Added

- First functional build of the API:
  - Runs on `0.0.0.0:8080` inside a container, made available at `eaji.cc` via Traefik routing.
  - Stores data in-memory.
- `healthcheck` route.
- `auth` routes.