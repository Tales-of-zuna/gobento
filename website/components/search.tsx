'use client';
import Link from 'next/link';
import {
  SearchDialog,
  SearchDialogClose,
  SearchDialogContent,
  SearchDialogHeader,
  SearchDialogIcon,
  SearchDialogInput,
  SearchDialogList,
  SearchDialogListItem,
  SearchDialogOverlay,
  type SearchItemType,
  type SharedProps,
} from 'fumadocs-ui/components/dialog/search';
import { useStaticSearch } from 'fumadocs-core/search/client';
import { useI18n } from 'fumadocs-ui/contexts/i18n';
import { basePath } from '@/lib/site';
import { DOCS_SECTIONS, sectionClass, sectionOfUrl } from '@/lib/docs-sections';
import { headStill, stillUrl } from '@/components/mascot/stills';

const KIND_LABELS: Record<string, string> = { page: 'Page', heading: 'Section', text: 'Text' };

/** Search dialog backed by the static Orama index exported at build time. */
export default function DefaultSearchDialog(props: SharedProps) {
  const { locale } = useI18n();
  // Static search fetches its index by absolute URL and ignores Next's basePath.
  const search = useStaticSearch({ from: `${basePath}/api/search`, locale });
  const close = () => props.onOpenChange(false);

  return (
    <SearchDialog {...search} {...props}>
      <SearchDialogOverlay className="docs-search-scrim" />
      <SearchDialogContent className="docs-search">
        <SearchDialogHeader className="docs-search-head">
          <SearchDialogIcon />
          <SearchDialogInput placeholder="Search docs" />
          <SearchDialogClose>Esc</SearchDialogClose>
        </SearchDialogHeader>
        {search.search.length === 0 && <EmptySearch onNavigate={close} />}
        <SearchDialogList Item={ResultItem} className="docs-search-list" />
      </SearchDialogContent>
    </SearchDialog>
  );
}

/** A result, marked with its section color and its kind (page, section, text). */
function ResultItem({
  item,
  onClick,
}: {
  item: SearchItemType;
  onClick: () => void;
  table?: unknown;
}) {
  const url = 'url' in item && typeof item.url === 'string' ? item.url : '';
  const section = sectionOfUrl(url);
  const kind = 'type' in item ? item.type : '';
  return (
    <SearchDialogListItem
      item={item}
      onClick={onClick}
      className={`docs-sr ${sectionClass(section)}`}
      data-kind={KIND_LABELS[kind] ?? undefined}
      data-section={section?.id}
    />
  );
}

/** Before a query: the mascot and a chip per section to jump straight in. */
function EmptySearch({ onNavigate }: { onNavigate: () => void }) {
  return (
    <div className="docs-search-empty">
      <div className="docs-search-emp">
        <picture>
          <source srcSet={stillUrl(headStill(256), 'webp')} type="image/webp" />
          <img src={stillUrl(headStill(256), 'png')} alt="" width={88} height={88} />
        </picture>
        <div>
          <p className="docs-search-emp-t">What are you looking for?</p>
          <p className="docs-search-emp-s">
            Search page titles, headings and text across the docs.
          </p>
        </div>
      </div>
      <nav aria-label="Jump to a section" className="docs-search-jump">
        <p className="docs-search-jump-t" aria-hidden="true">
          Or jump to a section
        </p>
        <ul>
          {DOCS_SECTIONS.map((section) => (
            <li key={section.id}>
              <Link
                href={section.href}
                className={`docs-chip ${sectionClass(section)}`}
                onClick={onNavigate}
              >
                <i className="docs-mk" aria-hidden="true" />
                {section.short}
              </Link>
            </li>
          ))}
        </ul>
      </nav>
    </div>
  );
}
