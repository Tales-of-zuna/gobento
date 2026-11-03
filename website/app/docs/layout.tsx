import { source } from '@/lib/source';
import { DocsLayout } from 'fumadocs-ui/layouts/docs';
import { SiteFooter } from '@/components/site/SiteFooter';
import { SiteNav } from '@/components/site/SiteNav';
import './docs-theme.css';

/**
 * Title of the phone header Fumadocs keeps under the site bar (search and the
 * sidebar drawer). The site bar already carries the wordmark, so the desktop
 * sidebar hides it (docs-theme.css).
 */
const navTitle = <span className="docs-nav-label">Documentation</span>;

/**
 * The docs: the site's bar on top ("Docs v2"), Fumadocs' sidebar, page and
 * table of contents below it, then the five-color strip and the site footer.
 * GitHub and the theme switch live in the site bar, so Fumadocs' own copies
 * are off; search stays in the sidebar.
 */
export default function Layout({ children }: LayoutProps<'/docs'>) {
  return (
    <div className="docs-shell">
      <SiteNav active="docs" sticky />
      <DocsLayout
        tree={source.getPageTree()}
        nav={{ title: navTitle, url: '/docs/' }}
        themeSwitch={{ enabled: false }}
        // Every section stays open, as in the design: the sidebar is the map of the docs.
        sidebar={{ defaultOpenLevel: 1 }}
      >
        {children}
      </DocsLayout>
      <div className="docs-strip" aria-hidden="true">
        <i />
        <i />
        <i />
        <i />
        <i />
      </div>
      <SiteFooter flush />
    </div>
  );
}
