import Link from 'next/link';
import type { ReactNode } from 'react';
import { DocsDescription, DocsTitle } from 'fumadocs-ui/layouts/docs/page';
import { sectionClass, type DocsSection } from '@/lib/docs-sections';

/**
 * The page header band of the "Docs v2" design: the section's pastel with a
 * seigaiha pattern on the right, a tilted four-color tray, and a scalloped
 * bottom edge. Patterns are decorative and never sit behind body text.
 */
export function DocsPageHeader({
  title,
  description,
  section,
  isSectionPage,
}: {
  title: ReactNode;
  description?: ReactNode;
  section: DocsSection | undefined;
  /** True when the page is the section itself (no separate section crumb). */
  isSectionPage: boolean;
}) {
  return (
    <header className={`docs-band ${sectionClass(section)}`} data-section={section?.id ?? 'intro'}>
      <div className="docs-band-pat" aria-hidden="true" />
      <div className="docs-band-deco" aria-hidden="true">
        <i />
        <i />
        <i />
        <i />
      </div>
      <div className="docs-band-in">
        <nav aria-label="Breadcrumb" className="docs-crumbs">
          <ol>
            <li>
              <Link href="/docs">Docs</Link>
            </li>
            {section && !isSectionPage && (
              <li>
                <Link href={section.href} className="docs-crumb-sec">
                  <i className="docs-mk" aria-hidden="true" />
                  {section.title}
                </Link>
              </li>
            )}
            <li aria-current="page">{title}</li>
          </ol>
        </nav>
        <DocsTitle className="docs-h1">{title}</DocsTitle>
        <DocsDescription className="docs-lead">{description}</DocsDescription>
      </div>
      <div className="docs-band-scal" aria-hidden="true" />
    </header>
  );
}
