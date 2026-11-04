## ADDED Requirements

### Requirement: Responsive input
The configurator SHALL process each keystroke and activation promptly while
the 3D bento runs, including when WebGL is software-rendered, so that typing,
choosing, expanding Advanced, and copying never wait on the scene.

#### Scenario: Typing with the 3D bento running
- **WHEN** the 3D bento is showing and the visitor types 16 characters into the module path
- **THEN** all 16 characters are processed in under 2 seconds, and the command shows the typed module path
