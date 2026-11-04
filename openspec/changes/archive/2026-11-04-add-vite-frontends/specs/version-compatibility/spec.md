## ADDED Requirements

### Requirement: Node.js and npm versions are rules data
The rules file SHALL record the supported Node.js range for generation and
the Node.js release line of the generated tooling, and SHALL list the npm
packages components use, each with its components, its role (a scaffolder
run at generation time, or a dependency or development dependency added to
a scaffolded frontend), and one exact version fixed per gobento release.
Templates SHALL read these versions from the rules data. A pin of an npm
package SHALL be rejected as not pinnable.

#### Scenario: Scaffolder version from the rules
- **WHEN** a project with `--frontend svelte` is generated
- **THEN** the scaffolder runs as `npm create vite@<version>` with the version of `create-vite` in the rules file

#### Scenario: Pin of an npm package
- **WHEN** the user runs `gobento new shop ... --frontend react --pin create-vite=9.0.0`
- **THEN** gobento exits non-zero saying npm package versions are fixed per gobento release, and writes nothing

#### Scenario: Default Node.js line inside the range
- **WHEN** the rules file's Node.js release line lies outside its Node.js range
- **THEN** loading the catalog fails naming the line and the range

### Requirement: Node.js and npm rules are checked
The rules check SHALL reject an npm package without components or with a
malformed version, a scaffolder whose name does not start with `create-`,
npm packages without a Node.js rule, and a Node.js release line outside the
Node.js range.

#### Scenario: Scaffolder name
- **WHEN** the rules file lists a scaffolder named `vite`
- **THEN** loading the catalog fails saying a scaffolder's name starts with `create-`
