import Link from 'next/link';
import { stillUrl } from '@/components/mascot/stills';
import { docsRoute } from '@/lib/shared';
import { configuratorRoute, githubUrl, repoFile } from './links';
import s from './home.module.css';

/** The site footer: links, license, and the mascot attribution (DESIGN.md). */
export function SiteFooter() {
  return (
    <footer className={`${s.in} ${s.footer}`}>
      <div className={s.foot}>
        <div className={s.frow}>
          <div className={s.fmark}>
            {/* eslint-disable-next-line @next/next/no-img-element -- static export, pre-sized still */}
            <img src={stillUrl(64, 'webp')} width={64} height={64} alt="" decoding="async" />
            <Link className={s.mark} href="/">
              gobento
            </Link>
          </div>
          <nav className={s.flinks} aria-label="Footer">
            <ul>
              <li>
                <Link href={`${docsRoute}/`}>Docs</Link>
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
