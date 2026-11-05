# Spec Delta

## ADDED Requirements

### Requirement: List the tooling versions
After the vendored assets, `gobento list versions` SHALL print every project
tooling version of the rules file: each container image with its key, name,
tag, and digest if it has one, each GitHub Action with its version and commit
SHA, and each tool with its version.

#### Scenario: Tooling versions
- **WHEN** a user runs `gobento list versions`
- **THEN** the output lists the runtime image with its digest, the Go build image with the tag `{go}`, `actions/checkout` with its version and 40-character commit SHA, and `golangci-lint` with its version

## MODIFIED Requirements

### Requirement: JSON output
With `--json`, every form of `gobento list` SHALL print one JSON document
instead of text. `gobento list --json` SHALL print the same document the
website is built from, with the fields `kinds`, `components`, `go`, `modules`,
`assets`, `tooling`, `supported`, and `rules`. Lists SHALL be empty lists rather than null.

#### Scenario: Same data as the website
- **WHEN** `gobento list --json` runs for a gobento build
- **THEN** its output is byte for byte the catalog document the exporter writes for the website from the same catalog

### Requirement: JSON output of a subcommand
With `--json`, each subcommand SHALL print an object with only the fields it
lists, in the format of the full document: `components` prints `kinds` and
`components`, `selections` prints `supported`, and `versions` prints `go`,
`modules`, `assets`, and `tooling`. Filters SHALL apply to the JSON as to the text.

#### Scenario: Scripts read the components
- **WHEN** a script runs `gobento list components --kind database --json`
- **THEN** it receives an object whose `kinds` holds only the database kind and whose `components` holds the database components with their `id`, `kind`, `name`, `title`, `description`, `provides`, and `requires`

#### Scenario: Scripts read the tooling
- **WHEN** a script runs `gobento list versions --json`
- **THEN** its `tooling` object holds `images`, `actions`, and `tools`, and the `runtime` image's `digest` is the one in the rules file
