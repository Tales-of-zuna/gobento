import Link from 'next/link';
import { MascotHero } from '@/components/mascot/MascotHero';
import { PrereleaseNotice } from '@/components/site/PrereleaseNotice';
import type { ReleaseStatus } from '@/lib/release';
import { docsRoute } from '@/lib/shared';
import { CopyCommand } from './CopyCommand';
import { configuratorRoute, QUICK_START } from './links';
import { Item, PressLink, Stagger } from './motion';
import { HeroInstallNote } from './ReleaseCopy';
import s from './home.module.css';

/** The sky band under the nav: value proposition, quick start, calls to action and the mascot. */
export function Hero({ release }: { release: ReleaseStatus }) {
  const { released } = release;
  return (
    <div className={`${s.band} ${s.bSora} ${s.bHero}`}>
      <div className={`${s.pat} ${s.sei}`} aria-hidden="true" />
      <section className={`${s.in} ${s.hero}`} aria-labelledby="hero-title">
        <Stagger>
          {!released && (
            <Item>
              <PrereleaseNotice released={released} />
            </Item>
          )}
          <Item>
            <h1 id="hero-title" className={s.h1}>
              Pack your Go project <em>like a bento.</em>
            </h1>
          </Item>
          <Item>
            <p className={s.sub}>
              Choose an architecture, HTTP framework, database, data access and, if you like, a
              frontend and a CSS framework. gobento picks versions that are known to work together
              and generates a project that builds.
            </p>
          </Item>
          <Item>
            <CopyCommand command={QUICK_START} />
            <HeroInstallNote release={release} />
          </Item>
          <Item className={s.ctas}>
            <PressLink className={`${s.btn} ${s.pri}`} href={configuratorRoute} prefetch={false}>
              Open the configurator
            </PressLink>
            <Link className={`${s.btn} ${s.ghost}`} href={`${docsRoute}/`}>
              Read the docs
            </Link>
          </Item>
        </Stagger>
        <Stagger className={s.stage}>
          <Item kind="pop" className={s.stageIn}>
            <div className={s.halo} aria-hidden="true" />
            <div className={s.mascot}>
              <MascotHero sizes="(max-width: 640px) min(366px, calc(100vw - 24px)), 600px" />
            </div>
            <span className={`${s.stk} ${s.sOni}`} aria-hidden="true" />
            <span className={`${s.stk} ${s.sTam}`} aria-hidden="true" />
            <span className={`${s.stk} ${s.sUme}`} aria-hidden="true" />
            <span className={`${s.stk} ${s.sEda}`} aria-hidden="true" />
            <span className={`${s.stk} ${s.sSak}`} aria-hidden="true" />
          </Item>
        </Stagger>
      </section>
    </div>
  );
}
