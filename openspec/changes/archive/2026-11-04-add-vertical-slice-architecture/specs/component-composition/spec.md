## MODIFIED Requirements

### Requirement: One owner per slot
Each slot SHALL be filled by exactly one component, so no two components ever
write into the same Go package or the same file. A slot directory MAY lie
inside another slot's directory; the outer slot's owner SHALL NOT write files
into the inner slot's directory or below it.

#### Scenario: Two components target one slot
- **WHEN** two selected components both fill the slot `store`
- **THEN** gobento exits non-zero naming both components and the slot, and writes nothing

#### Scenario: File inside a nested slot
- **WHEN** an architecture places the slot `handlers` inside the directory of the slot `models`, and the owner of `models` has a template that would be written into the `handlers` directory
- **THEN** gobento exits non-zero naming the component, the file, and the slot `handlers`, and writes nothing
