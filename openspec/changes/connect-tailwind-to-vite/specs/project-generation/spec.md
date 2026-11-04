## MODIFIED Requirements

### Requirement: Dry run with a scaffolder
With `--dry-run`, gobento SHALL list the files it writes itself and SHALL
say which scaffolder adds more files to which directory, which npm packages
npm adds to the scaffolded project, and that generation needs Node.js and
network access to the npm registry. It SHALL NOT run the scaffolder or
require Node.js.

#### Scenario: Dry run of a React project
- **WHEN** the user runs `gobento new shop ... --frontend react --dry-run`
- **THEN** the output lists `frontend/vite.config.ts` and the other gobento files, says that `npm create vite@<version>` with the template `react-ts` adds more files to `frontend/` and writes `frontend/package-lock.json`, and nothing is written

#### Scenario: Dry run with Tailwind
- **WHEN** the user runs `gobento new shop ... --frontend react --css tailwind --dry-run`
- **THEN** the output also says that npm adds the development dependencies `@tailwindcss/vite` and `tailwindcss` at their versions to `frontend/package.json`

### Requirement: Scaffold step
When the selection includes a component with a scaffolder, generation SHALL
run it after writing gobento's files to the temporary directory and before
moving the project into place, merge its files under the component's
directory with gobento's files taking precedence, and record the npm
dependency tree in a lock file without installing the dependencies. When
npm refuses a remote tarball while writing the lock file (npm 12's
`allow-remote`), generation SHALL repeat that command allowing remote
tarballs and SHALL fail unless every package in the lock file resolves to
the configured npm registry's host. A failure or interruption SHALL leave
nothing behind, including files written by processes the scaffolder
started.

#### Scenario: Scaffolder fails
- **WHEN** the scaffolder or npm fails during generation (for example, without network access)
- **THEN** gobento exits non-zero with npm's output, and no generated files remain on disk

#### Scenario: Interrupted while scaffolding
- **WHEN** the user presses Ctrl-C while the scaffolder runs
- **THEN** gobento stops the scaffolder and every process it started, removes everything it created, and exits with status 130

#### Scenario: npm 12 refuses a registry tarball
- **WHEN** npm 12 refuses, with `EALLOWREMOTE`, the registry tarball of an optional dependency while writing the lock file of a React project with `--css tailwind`
- **THEN** generation repeats the install with `--allow-remote=all` and succeeds, and it fails, leaving nothing behind, if the lock file then resolves a package to another host
