## ADDED Requirements

### Requirement: List the Node.js rule and npm packages
`gobento list versions` (and `gobento list`) SHALL show the supported
Node.js range and release line, and every npm package with its role,
version, and components; the JSON output SHALL include them.

#### Scenario: npm packages in the listing
- **WHEN** the user runs `gobento list versions`
- **THEN** the output shows the Node.js range and `create-vite` with its version, the role `scaffolder`, and the frontends that use it
