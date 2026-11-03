import Link from 'next/link';
import { headStill, stillUrl } from '@/components/mascot/stills';
import { cn } from '@/lib/utils';
import { configuratorRoute, docsHomeRoute, githubUrl, repoFile } from './links';
import s from './site.module.css';

/** Props of {@link SiteFooter}. */
export interface SiteFooterProps {
  /** Sits right under another divider (the docs' five-color strip): no gap, no top rule. */
  readonly flush?: boolean;
}

/** The site footer: links, license, and the mascot attribution (DESIGN.md). */
export function SiteFooter({ flush = false }: SiteFooterProps) {
  return (
    <footer className={cn(s.footer, flush && s.flush)}>
      <div className={s.foot}>
        <div className={s.frow}>
          <div className={s.fmark}>
            {/* eslint-disable-next-line @next/next/no-img-element -- static export, pre-sized still */}
            <img
              src={stillUrl(headStill(64), 'webp')}
              width={64}
              height={64}
              alt=""
              decoding="async"
            />
            <Link className={s.mark} href="/">
              gobento
            </Link>
          </div>
          <nav className={s.flinks} aria-label="Footer">
            <ul>
              <li>
                <Link href={docsHomeRoute}>Docs</Link>
              </li>
              <li>
                <Link href={configuratorRoute} prefetch={false}>
                  Configurator
                </Link>
              </li>
              <li>
                <a href={githubUrl}>GitHub</a>
              </li>
              <li>
                <a href={repoFile('CODE_OF_CONDUCT.md')}>Code of Conduct</a>
              </li>
              <li>
                <a href={repoFile('SECURITY.md')}>Security</a>
              </li>
            </ul>
          </nav>
        </div>
        <p className={s.lic}>
          <a href={repoFile('LICENSE')}>MIT licensed</a>
        </p>
        <p className={s.credit}>
          The gobento mascot is an original character inspired by the Go gopher, designed by Renée
          French and licensed under{' '}
          <a href="https://creativecommons.org/licenses/by/4.0/" rel="license">
            CC BY 4.0
          </a>
          . gobento is not affiliated with or endorsed by the Go project.
        </p>
      </div>
    </footer>
  );
}
