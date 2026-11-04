## ADDED Requirements

### Requirement: Software-rendered WebGL
When the browser's WebGL is rendered in software rather than by a GPU, the
site SHALL still show the 3D mascot scene but hold it in its static pose and
draw it only when its content changes, so the scene never keeps the page
busy.

#### Scenario: Software rasterizer
- **WHEN** the WebGL renderer is a software rasterizer (for example SwiftShader or llvmpipe) and the visitor interacts with the page
- **THEN** the 3D scene replaces the still in its static pose, and no frames are drawn while nothing in the scene changes
