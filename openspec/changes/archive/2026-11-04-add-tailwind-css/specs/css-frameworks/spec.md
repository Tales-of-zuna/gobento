# Spec Delta

## Purpose
Defines the optional css kind: how a CSS framework styles the selected
frontend through the frontend-build trait and the frontend's contribution
points, and Tailwind CSS for server-rendered and Vite-built frontends.

## ADDED Requirements

### Requirement: CSS frameworks are optional and need a frontend
gobento SHALL treat the CSS framework as an optional component kind `css`,
selected with `--css`. A CSS framework SHALL require the trait
`frontend-build`, which every frontend provides, so it is supported only
together with a frontend that it can build for. Leaving out the CSS
framework SHALL keep the frontend's own stylesheet.

#### Scenario: No CSS framework by default
- **WHEN** the user runs `gobento new shop --arch mvc --framework nethttp --db sqlite --data sqlc --frontend htmx` without `--css`
- **THEN** the project serves the htmx frontend's plain stylesheet, and generating it again with `--css none` produces the same files

#### Scenario: CSS framework without a frontend
- **WHEN** the user runs `gobento new shop --arch mvc --framework nethttp --db sqlite --data sqlc --css tailwind`
- **THEN** gobento exits non-zero stating that the combination is not supported because `css/tailwind` requires the trait `frontend-build`, which no selected component provides, and writes nothing

#### Scenario: Not asked without a frontend
- **WHEN** a user in a terminal answers the frontend question with "None (API only)"
- **THEN** gobento does not ask for a CSS framework

### Requirement: Frontend markup is independent of the CSS framework
A frontend's pages and generated code SHALL be the same for every CSS
choice; a CSS framework SHALL reach a frontend only through the frontend
kind's contribution points that the frontend includes.

#### Scenario: Same pages with Tailwind
- **WHEN** projects are generated with `--frontend htmx` and with `--frontend htmx --css tailwind`
- **THEN** `pages.templ` and `pages_templ.go` are identical in both, and only the stylesheet and the Tailwind files differ

### Requirement: Tailwind CSS for server-rendered frontends
With `--css tailwind` and a frontend whose `frontend-build` is `go`, the
project SHALL ship `tailwind.css`, the Tailwind source that styles the
frontend's markup and compiles the utility classes of its templates, and the
stylesheet compiled from it with the Tailwind version from the rules file,
served in place of the frontend's plain stylesheet.

#### Scenario: Styles served from the binary
- **WHEN** a project generated with `--frontend htmx --css tailwind` runs and a browser opens `/`
- **THEN** the page loads `/static/styles.css` from the application, and it is the Tailwind output that styles the page

#### Scenario: Utility class added to a page
- **WHEN** a user adds a Tailwind utility class to a `.templ` file of the web package and runs `make generate`
- **THEN** the stylesheet contains that utility

### Requirement: Tailwind needs no Node.js and no CLI to build
A project generated with `--css tailwind` for a `go` frontend build SHALL
build and test with the go command alone. Regenerating the stylesheet SHALL
need only Go: the project's Tailwind runner SHALL download the standalone
Tailwind CLI of the pinned version for the current platform, refuse it
unless its SHA-256 matches the rules file, cache it, and run it.

#### Scenario: Build without the Tailwind CLI
- **WHEN** `go build ./...` runs in a freshly generated Tailwind project on a machine without network access or Node.js
- **THEN** the build succeeds and the binary serves the stylesheet

#### Scenario: Tampered download
- **WHEN** the downloaded Tailwind CLI does not match the recorded SHA-256
- **THEN** the runner exits non-zero naming the expected and actual checksums, and does not run or keep the file

#### Scenario: Unsupported platform
- **WHEN** the runner starts on a platform the rules file has no binary for
- **THEN** it exits non-zero listing the supported platforms and explaining how to use another Tailwind CLI binary

### Requirement: Tailwind output is shipped and current
The shipped stylesheet SHALL be exactly what the pinned Tailwind CLI
compiles from `tailwind.css` in the generated project, for every
architecture and project name. `make generate` SHALL regenerate it, and the
generated CI workflow SHALL fail when the committed stylesheet differs from
what the pinned CLI compiles.

#### Scenario: Regeneration changes nothing
- **WHEN** `make generate` runs in a freshly generated project with `--frontend htmx --css tailwind`
- **THEN** no file changes

#### Scenario: Stale stylesheet fails CI
- **WHEN** a class is added to `pages.templ` and committed without regenerating the stylesheet
- **THEN** the CI step for Tailwind shows the difference and fails

#### Scenario: Project name is a utility
- **WHEN** a project named `container` is generated with `--frontend htmx --css tailwind` and its stylesheet is regenerated
- **THEN** no file changes

### Requirement: Tailwind CSS for Vite frontends
With `--css tailwind` and a frontend whose `frontend-build` is `vite`, the
Tailwind component SHALL contribute the Vite plugin import and call, the
Tailwind stylesheet import, and the npm development dependencies of the
pinned release to the frontend kind's points that the frontend includes, and
SHALL add no Tailwind source, runner, or `make generate` step of its own.

#### Scenario: Vite frontend with Tailwind
- **WHEN** a project is composed with a Vite frontend that includes the points `vite-config-imports`, `vite-plugins`, `css-imports`, and `npm-dev-dependencies`, and `--css tailwind`
- **THEN** the frontend's Vite configuration imports and calls `@tailwindcss/vite`, its main stylesheet starts with `@import "tailwindcss";`, its install step names `tailwindcss` and `@tailwindcss/vite` at the pinned version, and the project has no `tailwind.css`

#### Scenario: Vite frontend without the points
- **WHEN** a Vite frontend does not include the Vite points in its templates
- **THEN** its selections with `--css tailwind` are not supported, and composing one fails naming the Tailwind contribution and the point it lost
