# Maintainer guide

This guide is for people with admin access to
[Tales-of-zuna/gobento](https://github.com/Tales-of-zuna/gobento). Repository
settings live on GitHub, not in the repository, so they are listed here to be
applied by hand. Each setting comes with the reason it exists.

## Repository settings

### About: description and topics

The **About** box on the repository page (gear icon) is the first thing people
and search engines see. Keep it accurate as the project grows:

- **Description:** "Go project generator: choose your framework, database,
  data access, architecture and dependency versions — compatibility-checked and
  proven to build in CI." (Before `v0.1.0` it ended with "Early development.")
- **Topics:** `go`, `golang`, `cli`, `project-generator`, `scaffolding`,
  `code-generator`, `boilerplate`, `templates`.
- **Website:** `https://tales-of-zuna.github.io/gobento/` (see
  [GitHub Pages](#github-pages)).

The same values can be set from a terminal:

```sh
gh repo edit Tales-of-zuna/gobento --description "..." --add-topic go,golang
```

### Branch protection for `main`

**Settings → Rules → Rulesets** (or **Settings → Branches → Add branch
protection rule**), targeting `main`:

| Setting | Value | Why |
|---|---|---|
| Require a pull request before merging | On (required approvals: 0 while there is a single maintainer; raise to 1 when a second maintainer joins) | Every change goes through a pull request and passes CI before it reaches `main`. GitHub does not let you approve your own pull request, so requiring an approval with one maintainer would block every merge. |
| Require status checks to pass | On, with the checks `lint`, `test`, `vuln`, `actions`, `pr-title`, and `website-ok` | These are the CI job names. `test` runs as a matrix, so select each `test (...)` entry GitHub lists after the first CI run. `pr-title` enforces the Conventional Commits title that becomes the squash commit. `website-ok` sums up the website checks (see [Website checks](#website-checks)). |
| Require branches to be up to date before merging | On | Checks run against the code that will actually land. |
| Require linear history | On | Keeps `main` a simple list of squash commits, which release-please reads to build the changelog. |
| Block force pushes and deletions | On | `main` history is never rewritten. |

Status checks only appear in the picker after they have run once, so open a
first pull request before configuring them.

See [About protected branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches).

### Merge options

**Settings → General → Pull Requests**:

| Setting | Value | Why |
|---|---|---|
| Allow merge commits | Off | One merge method keeps history predictable. |
| Allow squash merging | On, default message **Pull request title** (or "Pull request title and description") | The PR title, checked by `pr-title`, becomes the Conventional Commit on `main`, so contributors' branch commits can stay free-form. |
| Allow rebase merging | Off | Rebasing would put unchecked commit messages on `main`. |
| Automatically delete head branches | On | Merged branches do not pile up. |

See [Configuring commit squashing](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/configuring-commit-squashing-for-pull-requests)
and [Managing the automatic deletion of branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/managing-the-automatic-deletion-of-branches).

### Security

**Settings → Advanced Security** (called **Code security** on some accounts):

| Setting | Value | Why |
|---|---|---|
| Private vulnerability reporting | Enabled | [SECURITY.md](../SECURITY.md) and the issue form contact link send reporters to the private "Report a vulnerability" form. |
| Dependabot alerts | Enabled | Notifies maintainers about known vulnerabilities in dependencies. |
| Dependabot security updates | Enabled | Opens fix pull requests for those alerts automatically. Regular version updates are configured in `.github/dependabot.yml`; major upgrades of the website toolchain (TypeScript, `@types/node`, Next.js, React, three.js/R3F) are ignored there and done as planned OpenSpec changes, with `@types/node` kept on the Node major in `website/.nvmrc`. |

See [Configuring private vulnerability reporting](https://docs.github.com/en/code-security/security-advisories/working-with-repository-security-advisories/configuring-private-vulnerability-reporting-for-a-repository),
[Configuring Dependabot alerts](https://docs.github.com/en/code-security/dependabot/dependabot-alerts/configuring-dependabot-alerts),
and [Configuring Dependabot security updates](https://docs.github.com/en/code-security/dependabot/dependabot-security-updates/configuring-dependabot-security-updates).

### GitHub Actions

**Settings → Actions → General → Workflow permissions**:

| Setting | Value | Why |
|---|---|---|
| Allow GitHub Actions to create and approve pull requests | On | release-please opens the release pull request with the workflow's `GITHUB_TOKEN`. |

Leave the default workflow permissions on read-only. Each workflow requests
the extra permissions it needs.

See [Managing GitHub Actions settings for a repository](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/enabling-features-for-your-repository/managing-github-actions-settings-for-a-repository).

### GitHub Pages

The website in `website/` is published by `.github/workflows/pages.yml` to
<https://tales-of-zuna.github.io/gobento/>.

**Settings → Pages → Build and deployment**:

| Setting | Value | Why |
|---|---|---|
| Source | **GitHub Actions** | The workflow builds the static export and deploys it with `actions/deploy-pages`; there is no `gh-pages` branch. |

The first deployment creates the `github-pages` environment. Leave its
deployment branch rule at the default (`main` only, from the protected
branch), so only `main` can publish. The workflow runs on every push to `main`
that touches `website/`, `components/`, `internal/catalogjson/`, or
`internal/tools/catalogexport/` (the code that writes the website's
`catalog.json`), and can be started by hand from **Actions → Pages → Run
workflow**.

See [Configuring a publishing source](https://docs.github.com/en/pages/getting-started-with-github-pages/configuring-a-publishing-source-for-your-github-pages-site#publishing-with-a-custom-github-actions-workflow).

### Website checks

`.github/workflows/website.yml` runs on every pull request and every push to
`main`. Its first job, `changes`, looks at the changed files; when none is in
`website/`, `components/`, `internal/catalogjson/`,
`internal/tools/catalogexport/`, or the workflow itself, the other jobs are
skipped. A workflow-level path filter was not used
on purpose: a filtered workflow does not report at all on unrelated pull
requests, so a required check from it would block them forever.

| Job | What it checks |
|---|---|
| `website` | `pnpm install --frozen-lockfile`, that `go list ./...` sees nothing under `website/`, lint, type-check, Prettier, Vitest, and the static build |
| `website-e2e` | Playwright end-to-end tests with axe accessibility audits against the static export under `/gobento/`; when it fails, the HTML report, traces, and screenshots are attached to the run as the `playwright-report` artifact (kept for 7 days) |
| `lighthouse` | The Lighthouse budget (below); the reports are attached to the run as the `lighthouse-reports` artifact |
| `website-ok` | Always runs, and fails if any job above failed or was cancelled (skipped counts as passing) |

After the first run, add **only `website-ok`** to the required status checks
of the `main` ruleset. It covers the other three jobs, and unlike them it also
reports on pull requests that do not touch the site.

**Lighthouse budget.** The `project-website` spec requires the home page, the
configurator, and a docs page to score at least 90 (0.9) for both performance
and accessibility in a mobile Lighthouse audit. `website/lighthouserc.cjs`
holds the URLs and the assertions; each page is audited five times and the
median run counts, which smooths out runner noise. Lighthouse simulates a
mid-range phone on a slow 4G connection, so a failure usually means the page
ships too much JavaScript, CSS, or fonts, or renders its main text late. Open
the HTML report from the artifact to see which metric dropped. Fix the page
rather than lowering the budget; changing the budget means changing the spec.
The Lighthouse CI version is pinned in the workflow (`LHCI_VERSION`) and in
CONTRIBUTING.md; bump both together.

### Labels

**Issues → Labels**: the issue forms and docs use these labels. GitHub creates
`bug`, `enhancement`, `question`, and `good first issue` by default; create the
others.

| Label | Used by |
|---|---|
| `bug` | Bug report form |
| `enhancement` | Feature request form |
| `component` | Component request form |
| `triage` | All issue forms, until a maintainer has looked at the issue |
| `question` | Questions, as described in [SUPPORT.md](../SUPPORT.md) |
| `good first issue` | Issues suited to newcomers, linked from [CONTRIBUTING.md](../CONTRIBUTING.md) |

## Cutting a release

Releases are automated with [release-please](https://github.com/googleapis/release-please)
and GoReleaser (see `.github/workflows/release.yml`):

1. Every push to `main` makes release-please open or update a release pull
   request. It proposes the next SemVer version and a `CHANGELOG.md` entry
   built from the Conventional Commit titles. While gobento is pre-1.0, breaking
   changes bump the minor version.
2. To release, review and **merge the release pull request**. Nothing is
   published before that.
3. The same workflow then tags the release, creates the GitHub release, and
   runs GoReleaser to build and attach the Linux, macOS, and Windows binaries
   with `checksums.txt`.

To change the proposed version, edit the release pull request as described in
the release-please documentation, or add a `Release-As: x.y.z` footer to a
commit on `main`.

**First release.** With no previous release, release-please proposes `1.0.0`
regardless of `bump-minor-pre-major`, so `release-please-config.json`
temporarily set `"release-as": "0.1.0"` for the first release. `v0.1.0` has
been released and that line has been removed; release-please now derives each
version from the commits since the previous release. Do not add `release-as`
back to the config, because it would force every later release to the same
version. Use a `Release-As: x.y.z` commit footer for a one-off override.

## Bumping golangci-lint

The golangci-lint version is pinned in two places, which must stay in sync:

- the golangci-lint action's `version` input in `.github/workflows/ci.yml`
- the install command in [CONTRIBUTING.md](../CONTRIBUTING.md)

To bump it:

1. Read the [golangci-lint changelog](https://golangci-lint.run/docs/product/changelog/)
   for new or changed linters.
2. Update the version in both files in one pull request (for example
   `ci: bump golangci-lint to vX.Y.Z`).
3. Run `golangci-lint run` locally with the new version and fix any new
   findings, or adjust [.golangci.yml](../.golangci.yml), in the same pull
   request.

### Tooling versions of generated projects

Generated projects get their own images, action SHAs, and golangci-lint
version from the `tooling` section of
[components/versions.yaml](../components/versions.yaml), separately from
gobento's own CI. Dependabot does not update that file, so check it when
gobento's own actions or golangci-lint are bumped, and at least once per
release:

1. Look up the latest releases and their commit SHAs with
   `gh api repos/<owner/repo>/releases/latest --jq .tag_name` and
   `gh api repos/<owner/repo>/commits/<tag> --jq .sha`, and the runtime image
   digest with `docker buildx imagetools inspect gcr.io/distroless/static-debian13:nonroot`.
2. Update the entries and the "Last checked" date, then run
   `go test ./internal/compose -run TestGolden -update` and review the diff.
3. Run `go test -count=1 -run TestIntegration ./internal/generate/`: it lints
   every generated project with the new golangci-lint version, runs actionlint
   on every workflow, and builds the images. Fix new findings in the templates
   in the same pull request.
