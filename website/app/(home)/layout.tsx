import { HomeMotion } from '@/components/home/motion';

export default function Layout({ children }: LayoutProps<'/'>) {
  return <HomeMotion>{children}</HomeMotion>;
}
