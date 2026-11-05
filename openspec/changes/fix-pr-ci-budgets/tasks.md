# Tasks

## 1. Website layout stability

- [x] 1.1 Set the code font's `display` to `optional` in `website/app/layout.tsx` and update its comment
- [x] 1.2 Add a `cumulative-layout-shift` assertion (max 0.1, median run) to `website/lighthouserc.cjs`
- [x] 1.3 Build the website and run Lighthouse on all three pages locally; confirm every page has CLS at most 0.1 and performance at least 0.9

## 2. Integration timeout

- [x] 2.1 Pass `-timeout 60m` to the integration step in `.github/workflows/ci.yml` and to `make test-all`
- [x] 2.2 Add the flag to every documented `TestIntegration` command (CONTRIBUTING.md, docs/maintainers.md, `website/content/docs/contributing/adding-a-component.mdx`)

## 3. Verification

- [x] 3.1 Run website unit tests, lint, typecheck, and prettier; run actionlint and `openspec validate --strict`
- [ ] 3.2 Push to `next` and confirm every check on pull request #13 passes
