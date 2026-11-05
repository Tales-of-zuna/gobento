# Design

## Where the sections go

One new page, not one page per option: the sections are short, read as
variations of the same project, and link to each other. A single page keeps
them together and leaves the sidebar short. It follows Quick start in
Getting started, because it continues the same `shop` example.

Section headings stay the same, so the anchors (`#use-postgresql`, …) keep
their names on the new page. Nothing outside quick-start links to those
anchors (checked across the repo), so no other page needs edits.

## Measuring

The Lighthouse budget keeps testing quick-start, the page a newcomer opens
first. The change is measured locally against the same static export before
and after: document size, DOM size, and performance over five runs.
