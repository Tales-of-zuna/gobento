'use client';
/**
 * Switches between the light and dark themes. The theme follows the system
 * until the visitor picks one; next-themes (inside Fumadocs' RootProvider)
 * stores the choice and applies it before paint, so there is no flash.
 */
import { useTheme } from 'fumadocs-ui/provider/base';
import { useSyncExternalStore } from 'react';
import s from './home.module.css';

const noSubscription = () => () => {};

export function ThemeToggle() {
  const { resolvedTheme, setTheme } = useTheme();
  // The resolved theme is only known in the browser; render a neutral label on the server.
  const mounted = useSyncExternalStore(
    noSubscription,
    () => true,
    () => false,
  );
  const next = resolvedTheme === 'dark' ? 'light' : 'dark';
  return (
    <button
      type="button"
      className={s.icon}
      aria-label={mounted ? `Switch to ${next} theme` : 'Switch theme'}
      data-theme-toggle
      onClick={() => setTheme(next)}
    >
      <span className={s.moon} aria-hidden="true" />
    </button>
  );
}
