'use client';
/**
 * The site bar shared by every page: wordmark, Docs, Configurator, GitHub,
 * theme switch, and a mobile menu. On the home page it lies over the hero
 * band; on the configurator and docs it is a bar on the page wash.
 */
import Link from 'next/link';
import { useEffect, useId, useRef, useState } from 'react';
import { cn } from '@/lib/utils';
import { BentoMark } from './BentoMark';
import { configuratorRoute, docsHomeRoute, githubUrl } from './links';
import { ThemeToggle } from './ThemeToggle';
import s from './site.module.css';

/** The site section a page belongs to; its nav item is marked current. */
export type SiteSection = 'docs' | 'configurator';

/** Props of {@link SiteNav}. */
export interface SiteNavProps {
  /** Marks the matching nav item as the current section. */
  readonly active?: SiteSection;
  /** `overlay` lies over the hero band (home); `bar` is a solid bar (other pages). */
  readonly variant?: 'overlay' | 'bar';
  /** Keeps the bar at the top of the window on wider screens (docs). */
  readonly sticky?: boolean;
}

const ITEMS: readonly { section: SiteSection; label: string; href: string }[] = [
  { section: 'docs', label: 'Docs', href: docsHomeRoute },
  { section: 'configurator', label: 'Configurator', href: configuratorRoute },
];

export function SiteNav({ active, variant = 'bar', sticky = false }: SiteNavProps) {
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
  const current = (section: SiteSection) => (section === active ? 'page' : undefined);
  // The configurator is a large client page; only prefetch it on intent.
  const prefetch = (section: SiteSection) => (section === 'configurator' ? false : undefined);

  return (
    <header
      className={cn(s.header, variant === 'overlay' ? s.overlay : s.bar, sticky && s.sticky)}
      data-site-nav={variant}
    >
      <div className={s.nav}>
        <Link className={s.mark} href="/">
          <BentoMark />
          gobento
        </Link>
        <div className={s.tools}>
          <nav className={s.links} aria-label="Main">
            <ul>
              {ITEMS.map((item) => (
                <li key={item.section}>
                  <Link
                    className={s.navLink}
                    href={item.href}
                    prefetch={prefetch(item.section)}
                    aria-current={current(item.section)}
                  >
                    {item.label}
                  </Link>
                </li>
              ))}
              <li>
                <a className={s.gh} href={githubUrl}>
                  GitHub
                </a>
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
            {ITEMS.map((item) => (
              <li key={item.section}>
                <Link
                  className={s.menuLink}
                  href={item.href}
                  prefetch={prefetch(item.section)}
                  aria-current={current(item.section)}
                  onClick={close}
                >
                  {item.label}
                </Link>
              </li>
            ))}
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
