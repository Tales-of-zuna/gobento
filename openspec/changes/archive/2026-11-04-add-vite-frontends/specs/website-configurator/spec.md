## ADDED Requirements

### Requirement: File preview of scaffolded frontends
The configurator's file preview SHALL show, for a Vite frontend, the
gobento-owned files under `frontend/` and mark the directory as completed by
the official scaffolder at generation time, without listing scaffolded
files it cannot know.

#### Scenario: React preview
- **WHEN** the user selects the React frontend in the configurator
- **THEN** the file preview shows `frontend/` with `vite.config.ts` and a note that create-vite adds the rest, and the web package directory
