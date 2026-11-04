## MODIFIED Requirements

### Requirement: Tailwind CSS for Vite frontends
With `--css tailwind` and a frontend whose `frontend-build` is `vite`, the
Tailwind component SHALL contribute the Vite plugin import and call and the
Tailwind stylesheet import to the frontend kind's points
`vite-config-imports`, `vite-plugins`, and `css-imports`, and SHALL add no
Tailwind source, runner, or `make generate` step of its own. Its npm
packages, `tailwindcss` and `@tailwindcss/vite`, SHALL be development
dependencies in the rules file's npm section, at the version of the
standalone Tailwind CLI, which the scaffold step adds to the frontend.

#### Scenario: Vite frontend with Tailwind
- **WHEN** a project is composed with `--frontend react`, `--frontend vue`, or `--frontend svelte` and `--css tailwind`
- **THEN** `frontend/vite.config.ts` imports `@tailwindcss/vite` and calls it last in its plugins, the stylesheet the app imports starts with `@import "tailwindcss";` followed by the frontend's own rules, and the project has no `tailwind.css`, Tailwind runner, or Tailwind step in its Makefile, CI workflow, or Dockerfile

#### Scenario: Tailwind's npm packages from the rules
- **WHEN** a project with a Vite frontend and `--css tailwind` is prepared
- **THEN** its scaffold step adds `@tailwindcss/vite` and `tailwindcss` at the version of the rules file's `tailwindcss` binary as development dependencies, and with `--frontend htmx` there is no scaffold step

#### Scenario: Vite frontend without the points
- **WHEN** a Vite frontend does not include the Vite points in its templates
- **THEN** its selections with `--css tailwind` are not supported, and composing one fails naming the Tailwind contribution and the point it lost
