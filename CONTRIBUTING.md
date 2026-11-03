# Contributing to gobento

Thank you for your interest in gobento. Every contribution helps, whether it is
a bug report, a documentation fix, a test, or a new component. This guide
explains how to set up your environment, how changes are planned, and how pull
requests are reviewed.

By taking part, you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
If you have a question rather than a change, see [SUPPORT.md](SUPPORT.md). To
report a security problem, follow [SECURITY.md](SECURITY.md) instead of opening
an issue.

## Prerequisites

- [Go](https://go.dev/dl/) 1.26 or newer
- [Git](https://git-scm.com/)
- `make` (optional; every target has a plain command equivalent, listed below)
- [golangci-lint](https://golangci-lint.run/) v2, installed with the
  [official instructions](https://golangci-lint.run/docs/welcome/install/). For
  example, on Linux or macOS:

  ```sh
  # Keep this version in sync with .github/workflows/ci.yml.
  curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$(go env GOPATH)/bin" v2.15.0
  ```

  The official guide also covers Homebrew, Windows, and other package managers.
  Please use the same version as CI, so local results match.

## First build

```sh
git clone https://github.com/Tales-of-zuna/gobento.git
cd gobento
make lint test
```

If those pass, your environment is ready. Run `make` (or `make help`) to list
all targets.

## Make targets

The [Makefile](Makefile) is a thin wrapper. Each target runs one or two plain
commands, so you can always run them directly.

| Target | What it does | Plain command |
|---|---|---|
| `help` | Lists all targets (the default) | Read the [Makefile](Makefile) |
| `build` | Builds the binary into `bin/` | `go build -o bin/gobento ./cmd/gobento` |
| `test` | Runs fast unit tests with the race detector | `go test -short -race ./...` |
| `test-all` | Runs every test, including slow ones | `go test -race ./...` |
| `lint` | Runs the linters configured in [.golangci.yml](.golangci.yml) | `golangci-lint run` |
| `fmt` | Formats the code | `golangci-lint fmt` |
| `tidy` | Cleans up `go.mod` and `go.sum` | `go mod tidy` |
| `cover` | Prints test coverage per function | `go test -short -coverprofile=coverage.out ./...` then `go tool cover -func=coverage.out` |
| `clean` | Removes build and coverage output | Delete `bin/` and `coverage.out` |

### Notes for Windows

- If you do not have `make`, run the plain commands from the table above in
  PowerShell or Command Prompt.
- The race detector (`-race`) needs cgo and a C compiler. If you do not have
  one, run `go test -short ./...` instead. CI runs the race detector on Linux
  and macOS, so races are still caught.
- The repository uses LF line endings everywhere (see `.gitattributes`). Git
  handles this for you; please do not commit CRLF line endings.

## How changes are planned: OpenSpec

gobento is developed spec-first with [OpenSpec](https://github.com/Fission-AI/OpenSpec).
Every change to behavior starts as an OpenSpec change in
[`openspec/changes/`](openspec/changes), before any code is written. A change
holds a proposal (why and what), a design (how), delta specs (the required
behavior), and a task list.

The workflow is:

1. **Propose.** Create a change in `openspec/changes/<change-name>/` describing
   the problem and the intended behavior. Open a pull request or an issue to
   discuss it early; small proposals are fine.
2. **Review.** Maintainers and contributors agree on the proposal and design
   before implementation starts. This avoids wasted work.
3. **Apply.** Implement the tasks, keeping code, tests, and specs in step.
4. **Archive.** After the change is merged, it is archived and its specs are
   merged into `openspec/specs/`, the source of truth for how
   gobento behaves.

Install the [OpenSpec CLI](https://github.com/Fission-AI/OpenSpec) to create,
validate, and archive changes, for example `openspec list` and
`openspec validate <change-name>`. If you use an AI coding assistant, you can
generate its OpenSpec integration locally with
`openspec init --tools <your-tool>` (for example `claude`, `cursor`, or
`codex`). Those generated files are gitignored and stay on your machine; only
the `openspec/` directory is committed.

Changes that do not alter behavior, such as typo fixes, refactors, test
additions, or documentation updates, do not need an OpenSpec change. Say
"no behavior change" in the pull request instead.

## Pull requests

### Titles follow Conventional Commits

Pull requests are squash-merged, and the PR title becomes the commit message on
`main`. That is why the **PR title** must follow
[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/); a CI
check enforces it. Commits inside your branch can be written however you like.

The format is `type(optional scope): description`, for example
`feat(cli): add --dry-run flag` or `docs: fix typo in README`.

Allowed types:

| Type | Use for |
|---|---|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation only |
| `refactor` | Code changes that neither fix a bug nor add a feature |
| `perf` | Performance improvements |
| `test` | Adding or fixing tests |
| `build` | Build system or dependencies |
| `ci` | CI configuration |
| `chore` | Maintenance that does not fit elsewhere |
| `revert` | Reverting an earlier change |

Mark a breaking change with `!` after the type or scope (for example
`feat!: rename the --db flag`), or with a `BREAKING CHANGE:` footer in the PR
description. Releases and the changelog are generated from these titles.

### Before you open a pull request

- Run `make lint test` (or the plain commands) and make sure both pass.
- Add or update tests for the behavior you changed.
- Update documentation that your change affects.
- Link the OpenSpec change, or state "no behavior change".

The pull request template contains this checklist.

### Review process

- A maintainer reviews every pull request. Expect a first response within a
  few days; feel free to ping the PR if it has been quiet for a week.
- All CI checks must pass before merging.
- Reviews are a conversation. Questions and suggestions are about the code, not
  about you, and it is fine to push back with reasons.
- Once approved, a maintainer squash-merges the pull request.

## Good first issues

New to the project or to open source? Look for issues labeled
[`good first issue`](https://github.com/Tales-of-zuna/gobento/labels/good%20first%20issue).
They are scoped to be approachable, and maintainers are happy to guide you. If
you want to work on one, comment on the issue so others know it is taken. If
you get stuck, ask in the issue; there are no silly questions.

## Adding a component

Components (frameworks, databases, data access libraries, architectures,
frontends) are the heart of gobento. To suggest one, open a
[component request](https://github.com/Tales-of-zuna/gobento/issues/new?template=component_request.yml).
For how to build one, see [docs/components.md](docs/components.md).

## License

By contributing, you agree that your contributions are licensed under the
[MIT License](LICENSE) that covers the project.
