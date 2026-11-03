import { SiteFooter } from '@/components/site/SiteFooter';
import { SiteNav } from '@/components/site/SiteNav';

/** The configurator uses the site's shared bar and footer, like every page. */
export default function Layout({ children }: LayoutProps<'/configurator'>) {
  return (
    <>
      <SiteNav active="configurator" />
      <main className="flex flex-1 flex-col">{children}</main>
      <SiteFooter />
    </>
  );
}
