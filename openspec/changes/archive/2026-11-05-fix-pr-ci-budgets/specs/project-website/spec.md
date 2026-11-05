## MODIFIED Requirements

### Requirement: Performance
The home page SHALL render its text content before loading the 3D scene, and
the home, configurator, and a docs page SHALL each score at least 90 for
performance in a mobile Lighthouse audit. Loading web fonts SHALL NOT move
content that is already on screen.

#### Scenario: Lighthouse budget
- **WHEN** Lighthouse runs in mobile mode against the home, configurator, and a docs page
- **THEN** each scores at least 90 for performance and accessibility

#### Scenario: Stable layout while fonts load
- **WHEN** Lighthouse runs in mobile mode against the home, configurator, and a docs page whose opening text contains inline code
- **THEN** each page's cumulative layout shift is at most 0.1
