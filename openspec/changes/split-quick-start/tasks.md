# Tasks

## 1. Split the page

- [x] 1.1 Move the sections from "With the vertical-slice architecture" through "Tailwind CSS with React, Vue, or Svelte" into `website/content/docs/getting-started/choose-your-stack.mdx`, with front matter and an intro that links back to Quick start
- [x] 1.2 Point quick-start's intro and flag explanation at the new page's anchors, and add it to the Next steps list
- [x] 1.3 Add the page to `getting-started/meta.json` after quick-start

## 2. Tests

- [x] 2.1 Add the page to the page list in `website/tests/docs.spec.ts`
- [x] 2.2 Run website unit tests, lint, typecheck, prettier, and the docs e2e tests

## 3. Verification

- [x] 3.1 Build and compare quick-start before and after locally: document size, DOM size, Lighthouse performance (five runs)
- [ ] 3.2 Push to `next` and confirm every check on pull request #13 passes
