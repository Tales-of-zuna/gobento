# Spec Delta

## ADDED Requirements

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
