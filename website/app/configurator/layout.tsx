import { HomeLayout } from 'fumadocs-ui/layouts/home';
import { baseOptions } from '@/lib/layout.shared';

/** The configurator uses the site's shared navigation, like the home page. */
export default function Layout({ children }: LayoutProps<'/configurator'>) {
  return <HomeLayout {...baseOptions()}>{children}</HomeLayout>;
}
