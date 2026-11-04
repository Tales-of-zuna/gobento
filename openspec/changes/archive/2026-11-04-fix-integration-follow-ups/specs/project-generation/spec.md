## MODIFIED Requirements

### Requirement: No questions for invalid command-line input
If a name, module path, `--pin` value, or component flag given on the command
line is invalid, or the component flags match no supported selection, gobento
SHALL ask nothing and report the problems as it does without a terminal. When
`<name>` is also omitted, the report SHALL include the missing name together
with every other problem.

#### Scenario: Invalid flag value
- **WHEN** the user runs `gobento new shop --framework express` in a terminal
- **THEN** gobento asks nothing, exits non-zero listing the supported frameworks, and writes nothing

#### Scenario: Invalid flag value without a name
- **WHEN** the user runs `gobento new --framework express` in a terminal
- **THEN** gobento asks nothing and exits non-zero reporting both the missing project name and the unknown framework
