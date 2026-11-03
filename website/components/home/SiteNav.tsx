'use client';
/** The site bar: wordmark, Docs, Configurator, GitHub, theme switch, and a mobile menu. */
import Link from 'next/link';
import { useEffect, useId, useRef, useState } from 'react';
import { docsRoute } from '@/lib/shared';
import { BentoMark } from './BentoMark';
import { configuratorRoute, githubUrl } from './links';
import { PressAnchor } from './motion';
import { ThemeToggle } from './ThemeToggle';
import s from './home.module.css';

export function SiteNav() {
  const [open, setOpen] = useState(false);
  const menuId = useId();
  const button = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (!open) return;
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setOpen(false);
        button.current?.focus();
      }
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open]);

  const close = () => setOpen(false);

  return (
    <header className={s.header}>
      <div className={`${s.in} ${s.nav}`}>
        <Link className={s.mark} href="/">
          <BentoMark />
          gobento
        </Link>
        <div className={s.tools}>
          <nav className={s.links} aria-label="Main">
            <ul>
              <li>
                <Link className={s.navLink} href={`${docsRoute}/`}>
                  Docs
                </Link>
              </li>
              <li>
                <Link className={s.navLink} href={configuratorRoute} prefetch={false}>
                  Configurator
                </Link>
              </li>
              <li>
                <PressAnchor className={s.gh} href={githubUrl}>
                  GitHub
                </PressAnchor>
              </li>
            </ul>
          </nav>
          <ThemeToggle />
          <button
            ref={button}
            type="button"
            className={`${s.icon} ${s.burger}`}
            aria-label="Menu"
            aria-expanded={open}
            aria-controls={menuId}
            onClick={() => setOpen((value) => !value)}
          >
            <i aria-hidden="true" />
            <i aria-hidden="true" />
            <i aria-hidden="true" />
          </button>
        </div>
        <nav id={menuId} className={s.menu} aria-label="Mobile" hidden={!open}>
          <ul>
            <li>
              <Link className={s.menuLink} href={`${docsRoute}/`} onClick={close}>
                Docs
              </Link>
            </li>
            <li>
              <Link
                className={s.menuLink}
                href={configuratorRoute}
                prefetch={false}
                onClick={close}
              >
                Configurator
              </Link>
            </li>
            <li>
              <a className={s.menuLink} href={githubUrl} onClick={close}>
                GitHub
              </a>
            </li>
          </ul>
        </nav>
      </div>
    </header>
  );
}
