# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 12/09/2026

### Added

- `swagger` route to display the OpenAPI spec in Swagger UI. Uses the UNPKG CDN.
- Compatibility for middlewares in the Server package.

### Changed

- Updated the server mechanism to return an http.Server directly.

## [0.3.0] - 11/09/2026

### Added

- `openapi.yaml` route to get the OpenAPI specification.

### Fixed

- Event sorting was inconsistent when the limit was hit. The in-memory store now sorts first then applies pagination.

## [0.2.0] - 11/09/2026

### Added

- `user` routes.
- `event` routes.
- `ticket` routes.

### Changed

- Getting tickets from the datastore now works similarly to events with a filter struct.
- Events that are in draft status won't be canceled but deleted outright.
- Ticket capacity is now fully managed by the datastore and not the outfacing API.

### Fixed

- Reservations never expired. Expiration is now checked when getting, listing or confirming tickets.

## [0.1.0] - 28/09/2026

### Added

- First functional build of the API:
  - Runs on `0.0.0.0:8080` inside a container, made available at `eaji.cc` via Traefik routing.
  - Stores data in-memory.
- `healthcheck` route.
- `auth` routes.