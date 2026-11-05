# Spec Delta

## ADDED Requirements

### Requirement: Frontend code generation in the tooling
When the selected frontend ships generated code, such as the Go code templ
generates from `pages.templ`, `make generate` SHALL regenerate it with the
version from the rules file, the CI workflow SHALL fail when the committed
generated code is not what that version generates, and the README SHALL name
the command without make. The `.dockerignore` and `.gitignore` SHALL exclude
the files templ writes in watch mode. A project without a frontend SHALL get
none of this.

#### Scenario: Regenerate the pages with make
- **WHEN** `make generate` runs in a project generated with `--frontend htmx`
- **THEN** it runs `go generate` in the web package, whose `//go:generate` line runs the pinned templ, and the committed files stay unchanged

#### Scenario: Fresh templ code passes CI
- **WHEN** the CI step for templ runs in a freshly generated project with `--frontend htmx`
- **THEN** regenerating changes no file and the step succeeds

#### Scenario: Stale templ code fails CI
- **WHEN** `pages.templ` is changed and committed without regenerating `pages_templ.go`
- **THEN** the CI step for templ shows the difference and fails

#### Scenario: Watch-mode files stay out of the image
- **WHEN** a project is generated with `--frontend htmx`
- **THEN** its `.dockerignore` and `.gitignore` list `*_templ.txt`

### Requirement: Images serve the frontend
The image of a project with a frontend SHALL serve the pages and their static
files from the binary, with the API under `/api/`, and the README's Docker
section SHALL say so.

#### Scenario: Image serves the pages
- **WHEN** the image of a project generated with `--db sqlite --frontend htmx` is started with port 8080 published
- **THEN** `GET /api/health` responds 200, `GET /` responds 200 with the page, and `GET /static/htmx.min.js` responds 200
