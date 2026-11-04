## MODIFIED Requirements

### Requirement: Automatic selection of the only option
When a selection flag is omitted and its component kind has exactly one
supported value, gobento SHALL select that value. When the kind has several
values and gobento does not ask questions (see "Ask for missing input in a
terminal"), gobento SHALL exit non-zero naming the flag and listing the values.

#### Scenario: Single option is chosen automatically
- **WHEN** the user omits `--db` and `sqlite` is the only supported database
- **THEN** the project is generated with `sqlite`

#### Scenario: Several options require a choice
- **WHEN** the user omits a selection flag whose kind has several supported values, and stdin is not a terminal or `--no-input` is given
- **THEN** gobento exits non-zero naming the flag and its supported values, and writes nothing

## ADDED Requirements

### Requirement: Ask for missing input in a terminal
In a terminal (stdin and stdout) without `--no-input`, `gobento new` SHALL
ask, before validating, for the project name when no `<name>` is given, then
for the module path (empty keeps the name), and for each kind whose flag is
omitted and that has several compatible options, in kind order. Answers SHALL
follow the name and module path rules. Nothing given by a flag SHALL be asked,
and `--dry-run` SHALL ask the same questions.

#### Scenario: Several frameworks, framework flag omitted
- **WHEN** the user runs `gobento new shop` in a terminal and the catalog has several supported frameworks
- **THEN** gobento asks for the framework, and not for the name, the module path, or a kind with a single option
- **AND** it generates the project with the chosen framework

#### Scenario: No name given
- **WHEN** the user runs `gobento new` in a terminal
- **THEN** gobento asks for the project name, rejecting names that break the project name rules, and then for the module path, using the name when the answer is empty

#### Scenario: Flags skip their questions
- **WHEN** the user runs `gobento new shop --framework nethttp` in a terminal
- **THEN** gobento does not ask for the framework

### Requirement: Questions offer only supported options
A component question SHALL offer only the options that form a supported
selection together with the component flags and the earlier answers. A kind
left with a single compatible option SHALL be selected without a question.

#### Scenario: Later options depend on earlier answers
- **WHEN** an earlier answer or flag rules out some options of a later kind
- **THEN** the later question offers only the options that form a supported selection with it, and is skipped when one option remains

### Requirement: No questions for invalid command-line input
If a name, module path, `--pin` value, or component flag given on the command
line is invalid, or the component flags match no supported selection, gobento
SHALL ask nothing and report the problems as it does without a terminal.

#### Scenario: Invalid flag value
- **WHEN** the user runs `gobento new shop --framework express` in a terminal
- **THEN** gobento asks nothing, exits non-zero listing the supported frameworks, and writes nothing

### Requirement: Equivalent command after questions
After asking at least one question, gobento SHALL print on stderr the
equivalent `gobento new` command, with every answer as an argument or flag
and the other flags of the run, so the run can be repeated without
questions.

#### Scenario: Framework chosen by a question
- **WHEN** the user answers `gin` to the framework question of `gobento new shop --dry-run`
- **THEN** stderr shows a command that starts with `gobento new shop` and contains `--framework gin` and `--dry-run`

### Requirement: Disable questions
gobento SHALL NOT ask any question when `--no-input` is given or when stdin or
stdout is not a terminal; it then behaves as if questions did not exist: a
missing `<name>` is an argument error, and a kind with several options and no
flag is reported as in "Automatic selection of the only option".

#### Scenario: Piped input
- **WHEN** the user runs `echo | gobento new` in a shell
- **THEN** gobento exits non-zero asking for exactly one project name, without asking a question

#### Scenario: No input flag in a terminal
- **WHEN** the user runs `gobento new shop --no-input` in a terminal and a kind has several options
- **THEN** gobento exits non-zero naming the kind's flag and its supported values, without asking a question

### Requirement: Abort while asking
When the user cancels a question with Ctrl-C or Esc, or the process is
interrupted while a question is open, gobento SHALL stop, write nothing, print
`gobento: interrupted; nothing was left behind`, and exit with status 130.
When the input ends before every question is answered, gobento SHALL exit
non-zero and write nothing.

#### Scenario: Ctrl-C at a question
- **WHEN** the user presses Ctrl-C while gobento asks for the framework
- **THEN** gobento exits with status 130, reports that it was interrupted, and creates no directory

#### Scenario: Input ends
- **WHEN** the input ends (for example, Ctrl-D in accessible mode) before the last question is answered
- **THEN** gobento exits non-zero saying the input ended, and creates no directory

### Requirement: Accessible questions
When the `ACCESSIBLE` environment variable is set to a non-empty value, or
`TERM` is `dumb`, gobento SHALL ask its questions as plain, line-based text
prompts, listing the options of a choice as numbered lines, instead of a
full-screen interface. The questions, options, validation, and results SHALL
be the same in both modes.

#### Scenario: Screen reader user
- **WHEN** the user runs `ACCESSIBLE=1 gobento new` in a terminal
- **THEN** each question is printed as a line of text, each option of a choice as a numbered line, and the answer is read as a typed line
