import { ClosingCta } from '@/components/home/ClosingCta';
import { Conveyor } from '@/components/home/Conveyor';
import { Hero } from '@/components/home/Hero';
import { HowItPacks } from '@/components/home/HowItPacks';
import { ProvenInCI } from '@/components/home/ProvenInCI';
import { SiteFooter } from '@/components/home/SiteFooter';
import { SiteNav } from '@/components/home/SiteNav';
import { Stack } from '@/components/home/Stack';
import { Versions } from '@/components/home/Versions';
import s from '@/components/home/home.module.css';
import { releaseStatus } from '@/lib/release.server';

export default function HomePage() {
  const { released } = releaseStatus();
  return (
    <div className={s.page}>
      <SiteNav />
      <main>
        <Hero released={released} />
        <Conveyor />
        <HowItPacks />
        <Stack />
        <Versions />
        <ProvenInCI />
        <ClosingCta />
      </main>
      <SiteFooter />
    </div>
  );
}
