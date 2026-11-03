import { ClosingCta } from '@/components/home/ClosingCta';
import { Conveyor } from '@/components/home/Conveyor';
import { Hero } from '@/components/home/Hero';
import { HowItPacks } from '@/components/home/HowItPacks';
import { ProvenInCI } from '@/components/home/ProvenInCI';
import { Stack } from '@/components/home/Stack';
import { SiteFooter } from '@/components/site/SiteFooter';
import { SiteNav } from '@/components/site/SiteNav';
import { Versions } from '@/components/home/Versions';
import s from '@/components/home/home.module.css';
import { releaseStatus } from '@/lib/release.server';

export default function HomePage() {
  const release = releaseStatus();
  return (
    <div className={s.page}>
      <SiteNav variant="overlay" />
      <main>
        <Hero release={release} />
        <Conveyor />
        <HowItPacks />
        <Stack />
        <Versions />
        <ProvenInCI />
        <ClosingCta release={release} />
      </main>
      <SiteFooter />
    </div>
  );
}
