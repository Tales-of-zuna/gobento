import { source } from '@/lib/source';
import { DocsLayout } from 'fumadocs-ui/layouts/docs';
import { baseOptions } from '@/lib/layout.shared';
import { appName } from '@/lib/shared';
import './docs-theme.css';

/** The wordmark: a four-compartment bento and the plain name. */
function Wordmark() {
  return (
    <span className="docs-mark">
      <span className="docs-logo" aria-hidden="true">
        <i />
        <i />
        <i />
        <i />
      </span>
      {appName}
    </span>
  );
}

export default function Layout({ children }: LayoutProps<'/docs'>) {
  const base = baseOptions();
  return (
    <>
      <DocsLayout
        tree={source.getPageTree()}
        {...base}
        nav={{ ...base.nav, title: <Wordmark /> }}
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
    </>
  );
}
