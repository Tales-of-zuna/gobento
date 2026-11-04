# Proposal

## Why

`gobento list versions` shows every version a generated project gets from
the rules file except the project tooling: the container images, the GitHub
Actions, and golangci-lint, which the `tooling` section of
`components/versions.yaml` holds since the project tooling change. A user who
wants to know which runtime image digest or action SHA a project will use has
to generate one or read the rules file. The vendored assets, which are fixed
per release like the tooling, are already listed, and the catalog document
the website is built from exports the assets but not the tooling.

## What Changes

- The catalog document (`internal/catalogjson`, `gobento list --json`,
  `website/src/generated/catalog.json`) gets a `tooling` field:
  `images` (`key`, `image`, `tag` with `{go}` kept, `digest` or `""`),
  `actions` (`name`, `version`, `sha`), and `tools` (`name`, `version`), each
  sorted by key or name.
- `gobento list versions` prints, after the vendored assets, a "Project
  tooling" table with the type, name, version (an image's `image:tag`), and
  what pins it (digest or commit SHA, `-` for none); `--json` adds `tooling`.
- The website's catalog schema accepts `tooling` (defaulting to empty lists);
  no page shows it yet.
- Tests: catalogjson export test, CLI snapshots (the frozen fixture gains the
  tooling), `TestListJSON` field list, a website schema test.
- Docs: `website/content/docs/cli/list.mdx` (sample output and JSON fields).

## Capabilities

### Modified Capabilities

- `catalog-discovery`: a new requirement lists the tooling versions, and the
  JSON document includes them.

## Impact

- `internal/catalogjson`, `internal/cli/list.go` and its testdata,
  `website/lib/catalog/schema.ts`, `website/src/generated/catalog.json`.
- Adding a JSON field is not a breaking change under the documented
  compatibility rule (later releases may add fields).
