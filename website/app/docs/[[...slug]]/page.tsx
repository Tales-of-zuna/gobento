import { source } from '@/lib/source';
import { DocsBody, DocsPage } from 'fumadocs-ui/layouts/docs/page';
import { findNeighbour } from 'fumadocs-core/page-tree';
import { notFound } from 'next/navigation';
import { getMDXComponents } from '@/components/mdx';
import type { Metadata } from 'next';
import { createRelativeLink } from 'fumadocs-ui/mdx';
import { editUrl } from '@/lib/shared';
import { sectionClass, sectionOfSlug } from '@/lib/docs-sections';
import { DocsPageHeader } from '@/components/docs/page-header';
import { DocsPageFooter } from '@/components/docs/page-footer';
import { releaseStatus } from '@/lib/release.server';

export default async function Page(props: PageProps<'/docs/[[...slug]]'>) {
  const params = await props.params;
  const page = source.getPage(params.slug);
  if (!page) notFound();

  const MDX = page.data.body;
  const section = sectionOfSlug(params.slug);
  const { previous, next } = findNeighbour(source.getPageTree(), page.url);

  return (
    <DocsPage
      toc={page.data.toc}
      full={page.data.full}
      className={`docs-page ${sectionClass(section)}`}
      // The header band carries the breadcrumb, and the page end its own previous/next cards.
      breadcrumb={{ enabled: false }}
      footer={{ enabled: false }}
      // Fumadocs renders the table of contents in plain divs; make it a landmark.
      tableOfContent={{ container: { role: 'navigation', 'aria-labelledby': 'toc-title' } }}
      tableOfContentPopover={{ container: { role: 'navigation', 'aria-label': 'On this page' } }}
    >
      <DocsPageHeader
        title={page.data.title}
        description={page.data.description}
        section={section}
        isSectionPage={params.slug?.length === 1}
        released={releaseStatus().released}
      />
      <DocsBody className="docs-body">
        <MDX
          components={getMDXComponents({
            // Lets pages link to each other with relative file paths.
            a: createRelativeLink(source, page),
          })}
        />
      </DocsBody>
      <DocsPageFooter editHref={editUrl(page.path)} previous={previous} next={next} />
    </DocsPage>
  );
}

export async function generateStaticParams() {
  return source.generateParams();
}

export async function generateMetadata(props: PageProps<'/docs/[[...slug]]'>): Promise<Metadata> {
  const params = await props.params;
  const page = source.getPage(params.slug);
  if (!page) notFound();

  return {
    title: page.data.title,
    description: page.data.description,
  };
}
