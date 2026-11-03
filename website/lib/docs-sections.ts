/**
 * The documentation sections and their bento color (approved "Docs v2"
 * design). A section is the first segment of a docs page's slug; its color
 * drives the sidebar group, the page header band, the TOC indicator, the
 * previous/next cards, and the search result markers.
 */

/** A food color of the design tokens (`--<color>-soft` pastel, `--<color>` marker). */
export type SectionColor = 'tamago' | 'sora' | 'shu' | 'matcha' | 'sakura';

export interface DocsSection {
  /** First slug segment of the section's pages. */
  id: string;
  title: string;
  /** Short label for chips. */
  short: string;
  color: SectionColor;
  /** Route (without the base path) of the section's first page. */
  href: string;
}

export const DOCS_SECTIONS: readonly DocsSection[] = [
  {
    id: 'getting-started',
    title: 'Getting started',
    short: 'Getting started',
    color: 'tamago',
    href: '/docs/getting-started/installation',
  },
  {
    id: 'concepts',
    title: 'Concepts',
    short: 'Concepts',
    color: 'sora',
    href: '/docs/concepts/components',
  },
  {
    id: 'cli',
    title: 'CLI reference',
    short: 'CLI reference',
    color: 'shu',
    href: '/docs/cli/new',
  },
  {
    id: 'components',
    title: 'Component reference',
    short: 'Components',
    color: 'matcha',
    href: '/docs/components',
  },
  {
    id: 'contributing',
    title: 'Contributing',
    short: 'Contributing',
    color: 'sakura',
    href: '/docs/contributing/adding-a-component',
  },
];

/** The section a page belongs to, from its slug; the introduction has none. */
export function sectionOfSlug(slug: readonly string[] | undefined): DocsSection | undefined {
  const first = slug?.[0];
  return DOCS_SECTIONS.find((section) => section.id === first);
}

/** The section of a docs URL (with or without the base path and trailing slash). */
export function sectionOfUrl(url: string): DocsSection | undefined {
  const match = /\/docs\/([^/#?]+)/.exec(url);
  return match ? DOCS_SECTIONS.find((section) => section.id === match[1]) : undefined;
}

/** Class that sets the `--gs` (pastel) and `--gm` (marker) variables of a section. */
export function sectionClass(section: DocsSection | undefined): string {
  return section ? `g-${section.color}` : 'g-none';
}
