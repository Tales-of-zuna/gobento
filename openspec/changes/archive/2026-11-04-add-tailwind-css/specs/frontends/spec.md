# Spec Delta

## MODIFIED Requirements

### Requirement: Frontend assets are served from the binary
The htmx frontend SHALL serve htmx and its stylesheet from files embedded in
the Go binary, with no request to a CDN or other network location, so the
application works offline. The htmx file SHALL be a pinned upstream release,
vendored unmodified. The stylesheet SHALL be the frontend's plain CSS, or
the one the selected CSS framework contributes.

#### Scenario: Offline page
- **WHEN** a project generated with `--frontend htmx` runs on a machine without network access and a browser opens `/`
- **THEN** every script and stylesheet of the page is loaded from the application itself, under `/static/`

#### Scenario: Vendored file matches upstream
- **WHEN** gobento's tests run with network access
- **THEN** the shipped htmx file is byte-for-byte the file of the recorded upstream release

#### Scenario: Stylesheet styles the markup
- **WHEN** a project generated with `--frontend htmx`, with or without a CSS framework, serves `/static/styles.css`
- **THEN** the response is CSS that styles the classes the pages use

## ADDED Requirements

### Requirement: Frontends name their build system
Every frontend SHALL provide the trait `frontend-build`, naming how its
assets are built: `go` for a frontend built by the Go toolchain alone, such
as htmx, and `vite` for a frontend built by Vite. CSS frameworks SHALL use
this trait to decide what to contribute.

#### Scenario: htmx build system
- **WHEN** a user runs `gobento list components --kind frontend`
- **THEN** `htmx` is listed with `provides: frontend-build=go`
