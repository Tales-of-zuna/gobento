# Tasks

## 1. Dependencies

- [x] 1.1 Add `charm.land/huh/v2` at its latest release (v2.0.3 on proxy.golang.org) and `golang.org/x/term`, run `go mod tidy`, keep `go 1.26.0`; verify with `go build ./...` and `govulncheck ./...` reporting no vulnerabilities

## 2. Shared validation rules

- [x] 2.1 Extract `generate.CheckName` and `generate.CheckModule` from `Prepare` with unchanged messages, and have `Prepare` call them; verify `go test ./internal/generate` and the several-problems snapshot in `internal/cli` pass unchanged, plus new table tests for both functions

## 3. Decision logic

- [x] 3.1 Add `internal/cli/prompt.go` with `question`, `asker`, `canAsk`, and `interview` (name, then module only when the name was asked, then each kind with several options among the matching supported selections, in kind order); verify with unit tests over hand-built selections for: skipping flagged kinds, silent single option, options filtered by earlier answers, name and module validation, and an empty module answer keeping the name
- [x] 3.2 Build the equivalent command after questions (module when it differs, asked/flagged/multi-option kinds, `--go`, `--pin`, `--dir`, `--dry-run`, shell quoting); verify with unit tests

## 4. huh rendering

- [x] 4.1 Add `internal/cli/prompt_huh.go`: one huh form per question, Ctrl-C and Esc as quit keys, accessible mode through a one-byte, EOF-recording reader and a colorprofile writer, context cancellation in both modes; verify with tests driving accessible mode (choice by number, re-prompt after an invalid name, default module on Enter, end of input, cancelled context)
- [x] 4.2 Choose accessible mode from `ACCESSIBLE` (non-empty) or `TERM=dumb`; verify with a unit test

## 5. Command wiring

- [x] 5.1 Wire prompts into `gobento new`: optional `<name>` only when asking, `--no-input`, terminal detection on stdin and stdout, no questions after an invalid command-line value, equivalent command on stderr, `errAborted` mapped to status 130 with the interrupted message; verify with CLI tests using a scripted asker and the two-framework catalog (asks the framework, skips it with `--framework`, `--no-input` error unchanged, invalid flag asks nothing, dry run asks and lists, abort exits 130 and generates nothing), and the existing non-interactive tests unchanged
- [x] 5.2 Update the help text (`Use`, `Long`, an example with `--no-input`) and `TestNewHelp`; verify `go test ./internal/cli`
- [x] 5.3 Update the docs: `README.md` (usage and roadmap tick), `website/content/docs/cli/new.mdx` (synopsis, argument, `--no-input`, a "Questions" section with accessibility and abort, exit status), `website/content/docs/getting-started/quick-start.mdx`, and `website/content/docs/concepts/components.mdx` (omitted flags); verify every documented command, message, and exit status against the implementation and `TestDocsSeveralProblems`, `TestWebsiteCommands` pass

## 6. Verification

- [x] 6.1 Run `go test ./...`, `go vet ./...`, `make lint`, `go build ./cmd/gobento`, and a non-interactive smoke run (`echo | go run ./cmd/gobento new shop --dry-run` prints the same output as before, and `echo | go run ./cmd/gobento new` keeps the argument error); verify all pass
- [x] 6.2 Run `openspec validate add-interactive-prompts --strict`; verify it reports the change as valid

## Workflow follow-up

- Archive the change with `openspec archive add-interactive-prompts -y` and verify with `openspec validate --specs --strict`.
- Commit on `feat/prompts` with a Conventional Commits message; do not push.
