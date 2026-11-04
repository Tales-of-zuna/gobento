# Design

## Context

After merging Tailwind CSS and the Vite frontends, two hooks for a CSS
framework in a Vite build existed: kind-level points in
`components/frontend/kind.yaml` (Tailwind) and base-layer points plus an
`npm:` section with `dev-dependency` packages (Vite). Neither reached the
other, so no Vite frontend was supported with `--css tailwind`.

## Decisions

### D1: Points live with the files that include them

The Vite points are declared once, in the frontend kind: only frontends'
files include them, and the kind already declares `stylesheet` for go
builds. `base.yaml` keeps the points of project-wide files. Names follow the
kind's existing ones (`vite-config-imports`, `vite-plugins`, `css-imports`),
which the Tailwind contributions already used. Alternative: keep them in
`base.yaml`; rejected because a base point suggests every project's files
may include it, and the kind-level guard by `frontend-build` documents
which frontends do.

### D2: A template may include base points and its own kind's

The Vite branch let any template include base points (the architectures'
`.gitignore` includes `gitignore`); the Tailwind branch restricted
component templates to their kind's points. Both are kept: base points are
shared by every file, a kind's points belong to its components, and naming
another kind's point is an error. The check that a non-empty contribution
reaches some template stays, so a frontend that drops a point makes its
selections unsupported instead of silently losing CSS.

### D3: npm versions are rules data, conditioned by traits

The `npm-dev-dependencies` point wrote versions in a template. It is
removed; the Tailwind packages join the `npm:` section, which the scaffold
step already installs. Binaries and npm packages gain the `traits:`
condition modules have, so the CLI is resolved only for
`frontend-build: go` and the npm packages only for `vite`; the listing, the
catalog document, and the configurator follow from the data. A test keeps
the three Tailwind versions equal, since upstream releases them together.

### D4: npm 12 and lock-only installs

npm 12 (`allow-remote` default `none`) refuses, in `--package-lock-only`
mode, the registry tarball of `@tailwindcss/oxide-wasm32-wasi`, whose
bundled dependencies must be read from the tarball (npm/cli#9800); a full
install and `npm ci` accept it. Generation retries such a refusal with
`--allow-remote=all` and then requires every `resolved` URL of the lock file
to be on the configured registry's host, which is what npm 12's default
means to guarantee. Alternative: install `node_modules` fully; rejected as
slower and contrary to the scaffold step's no-install design.

## Risks

- `--allow-remote=all` is unknown to npm 10; npm ignores unknown flags with a
  warning (checked with npm 10.9.2), and the retry only happens on
  `EALLOWREMOTE`, which npm 10 never reports.
