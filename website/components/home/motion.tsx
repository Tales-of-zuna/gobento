'use client';
/**
 * Motion helpers for the home page.
 *
 * Reveals are progressive enhancement: the server-rendered HTML is fully
 * visible, so content shows without JavaScript, before hydration, and to
 * crawlers. Animations only add to that (motion.module.css):
 *
 * - Page-load reveals (the hero) are CSS keyframes gated by the
 *   `motion-ready` class, which an inline script in the root layout puts on
 *   <html> before the first paint. Without JavaScript the class never
 *   appears and nothing animates.
 * - In-view reveals start when their group scrolls into the window, after
 *   hydration. A group that is already on screen when the page hydrates
 *   stays as it is instead of blinking out and back in.
 *
 * Everything sits behind prefers-reduced-motion: no-preference. Presses
 * spring with Motion.
 */
import Link from 'next/link';
import { motion, MotionConfig, type HTMLMotionProps } from 'motion/react';
import { useEffect, useRef, useState, type ComponentProps, type ReactNode } from 'react';
import { cn } from '@/lib/utils';
import m from './motion.module.css';

/** Wraps the page so Motion honours the visitor's reduced-motion setting. */
export function HomeMotion({ children }: { children: ReactNode }) {
  return <MotionConfig reducedMotion="user">{children}</MotionConfig>;
}

/**
 * Plays its children's reveals in sequence: on page load, or with `inView`
 * when the group first scrolls into the window.
 */
export function Stagger({
  children,
  className,
  inView = false,
}: {
  children: ReactNode;
  className?: string;
  inView?: boolean;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const [playing, setPlaying] = useState(false);

  useEffect(() => {
    const group = ref.current;
    if (!inView || !group || typeof IntersectionObserver === 'undefined') return;
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
    group.querySelectorAll<HTMLElement>('[data-reveal-item]').forEach((item, i) => {
      item.style.setProperty('--i', String(i));
    });
    let first = true;
    const observer = new IntersectionObserver(([entry]) => {
      const initial = first;
      first = false;
      if (!entry?.isIntersecting) return;
      // Already visible at hydration: leave it be. Otherwise it is entering now.
      if (!initial) setPlaying(true);
      observer.disconnect();
    });
    observer.observe(group);
    return () => observer.disconnect();
  }, [inView]);

  return (
    <div
      ref={ref}
      className={cn(inView ? m.view : m.load, className)}
      data-reveal={playing ? 'play' : undefined}
    >
      {children}
    </div>
  );
}

/** One revealed child of {@link Stagger}: rises (default) or pops in. */
export function Item({
  children,
  className,
  kind = 'rise',
  as: Tag = 'div',
}: {
  children: ReactNode;
  className?: string;
  kind?: 'rise' | 'pop';
  as?: 'div' | 'li';
}) {
  return (
    <Tag className={cn(m.item, kind === 'pop' && m.pop, className)} data-reveal-item>
      {children}
    </Tag>
  );
}

const MotionLink = motion.create(Link);

/** A link that lifts on hover and springs in on press. */
export function PressLink(props: ComponentProps<typeof MotionLink>) {
  return (
    <MotionLink
      whileHover={{ x: -2, y: -2 }}
      whileTap={{ x: 3, y: 3, scale: 0.98 }}
      transition={{ type: 'spring', stiffness: 500, damping: 22 }}
      {...props}
    />
  );
}

/** The external-link variant of {@link PressLink}. */
export function PressAnchor(props: HTMLMotionProps<'a'>) {
  return (
    <motion.a
      whileHover={{ x: -2, y: -2 }}
      whileTap={{ x: 3, y: 3, scale: 0.98 }}
      transition={{ type: 'spring', stiffness: 500, damping: 22 }}
      {...props}
    />
  );
}
