# project-website Specification

## Purpose
Gives gobento a public website that explains the project in seconds, carries its
visual identity and mascot, and stays fast, accessible, and honest about the
project's status for every visitor and device.

## Requirements

### Requirement: Landing page
The site's home page SHALL present the project name, a one-sentence value
proposition, the animated mascot, a copyable quick-start command, and links to
the configurator, the documentation, and the GitHub repository.

#### Scenario: First visit
- **WHEN** a visitor opens the home page
- **THEN** the name, value proposition, quick-start command, and links to the configurator, docs, and GitHub are visible without scrolling on a 1280×800 viewport

#### Scenario: Copy quick start
- **WHEN** the visitor activates the copy control next to the quick-start command
- **THEN** the command is placed on the clipboard and a visible confirmation appears

### Requirement: Pre-release notice
Until gobento has a published release, every page SHALL show a notice that the
project is in early development and that commands shown are the planned
interface.

#### Scenario: No release yet
- **WHEN** the site is built while no gobento release exists
- **THEN** each page shows the early-development notice

### Requirement: Light and dark themes
The site SHALL offer light and dark themes, SHALL follow the system preference
by default, and SHALL remember an explicit choice for the visitor.

#### Scenario: System preference
- **WHEN** a first-time visitor's system prefers dark mode
- **THEN** the site renders in the dark theme without a flash of the light theme

#### Scenario: Explicit choice persists
- **WHEN** the visitor switches to the light theme and reloads the page
- **THEN** the light theme is used

### Requirement: Reduced motion
When the visitor's system requests reduced motion, the site SHALL disable
non-essential animation, including continuous mascot animation, while keeping
all content and functions available.

#### Scenario: Reduced motion requested
- **WHEN** `prefers-reduced-motion: reduce` is set
- **THEN** the mascot is shown in a static pose and page transitions and decorative animations do not play

### Requirement: Works without WebGL
The site SHALL remain fully usable when WebGL is unavailable or the 3D scene
fails to load, showing a static illustration of the mascot instead.

#### Scenario: WebGL unavailable
- **WHEN** the browser cannot create a WebGL context
- **THEN** a static mascot illustration appears in place of the 3D scene, and no error is shown to the visitor

### Requirement: Accessibility
Every page SHALL meet WCAG 2.2 level AA: sufficient contrast in both themes,
full keyboard operation with a visible focus indicator, and text alternatives,
with the 3D scene treated as decorative.

#### Scenario: Keyboard only
- **WHEN** a visitor navigates the home page and the configurator using only the keyboard
- **THEN** every interactive element is reachable in a logical order, shows a visible focus indicator, and can be operated

#### Scenario: Automated audit
- **WHEN** an automated accessibility audit (axe) runs against the home, configurator, and a docs page in both themes
- **THEN** it reports no violations

### Requirement: Performance
The home page SHALL render its text content before loading the 3D scene, and
the home, configurator, and a docs page SHALL each score at least 90 for
performance in a mobile Lighthouse audit.

#### Scenario: Lighthouse budget
- **WHEN** Lighthouse runs in mobile mode against the home, configurator, and a docs page
- **THEN** each scores at least 90 for performance and accessibility

### Requirement: Mascot attribution and non-endorsement
The site SHALL credit the Go gopher's original designer, Renée French, under
CC BY 4.0, and SHALL state that gobento is not affiliated with or endorsed by
the Go project.

#### Scenario: Footer credits
- **WHEN** any page is displayed
- **THEN** its footer contains the gopher attribution with a link to the license and the non-affiliation statement

### Requirement: Publishing
The site SHALL be published to `https://tales-of-zuna.github.io/gobento/`
from the `main` branch, and every pull request that changes the site SHALL be
checked by building it.

#### Scenario: Merge to main
- **WHEN** a change to the site is merged to `main`
- **THEN** the published site reflects it after the deployment workflow completes

#### Scenario: Broken site change
- **WHEN** a pull request makes the site fail to build, lint, type-check, or pass its tests
- **THEN** the pull request's checks fail

### Requirement: Release-aware install guidance
The site SHALL derive its release-dependent copy from the release manifest at
build time. Once gobento has a published release, no page SHALL show the
pre-release notice, and the home page SHALL name the current version and link
to the installation guide and the prebuilt binaries instead of saying there is
nothing to install.

#### Scenario: Released
- **WHEN** the site is built while the release manifest has a version other than `0.0.0`
- **THEN** no page shows the pre-release notice, and the home page names that version and links to the installation guide and the GitHub releases page

#### Scenario: Unreleased build
- **WHEN** the site is built while the release manifest has version `0.0.0`
- **THEN** every page shows the pre-release notice and the home page says there is nothing to install yet
