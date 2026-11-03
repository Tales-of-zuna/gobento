import Link from 'next/link';
import type * as PageTree from 'fumadocs-core/page-tree';
import { sectionClass, sectionOfUrl } from '@/lib/docs-sections';
import { headStill, stillUrl } from '@/components/mascot/stills';

/**
 * The end of a docs page in the "Docs v2" design: the mascot pointing at the
 * edit link, then previous/next cards marked with their section's color.
 */
export function DocsPageFooter({
  editHref,
  previous,
  next,
}: {
  editHref: string;
  previous?: PageTree.Item;
  next?: PageTree.Item;
}) {
  return (
    <div className="docs-end not-prose">
      <div className="docs-guide">
        <picture className="docs-guide-mascot">
          <source srcSet={stillUrl(headStill(256), 'webp')} type="image/webp" />
          {/* Decorative: the bubble carries the message. */}
          <img src={stillUrl(headStill(256), 'png')} alt="" width={76} height={76} loading="lazy" />
        </picture>
        <p className="docs-bubble">
          <span>Something unclear or out of date?</span>
          <a href={editHref} target="_blank" rel="noreferrer noopener">
            Edit this page on GitHub
          </a>
        </p>
      </div>
      {(previous || next) && (
        <nav aria-label="Previous and next pages" className="docs-pn">
          {previous && <PageCard item={previous} direction="previous" />}
          {next && <PageCard item={next} direction="next" />}
        </nav>
      )}
    </div>
  );
}

function PageCard({ item, direction }: { item: PageTree.Item; direction: 'previous' | 'next' }) {
  const section = sectionOfUrl(item.url);
  const label = direction === 'previous' ? 'Previous' : 'Next';
  return (
    <Link
      href={item.url}
      className={`docs-pc docs-pc-${direction} ${sectionClass(section)}`}
      rel={direction === 'previous' ? 'prev' : 'next'}
    >
      <b>{item.name}</b>
      <small>
        {direction === 'previous' && <span aria-hidden="true">‹</span>}
        {label}
        {section && (
          <>
            <span aria-hidden="true">·</span>
            <i className="docs-mk" aria-hidden="true" />
            {section.title}
          </>
        )}
        {direction === 'next' && <span aria-hidden="true">›</span>}
      </small>
    </Link>
  );
}
