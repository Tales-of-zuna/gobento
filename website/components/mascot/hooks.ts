'use client';
/** Browser state the mascot reacts to, as small React hooks. */
import {
  useCallback,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
  type RefObject,
} from 'react';
import { clamp, type LookTarget } from './math';
import { prefersReducedMotion, REDUCED_MOTION_QUERY, supportsWebGL } from './support';

const noSubscription = () => () => {};

/** Reduced-motion preference, live; null during the server render. */
export function useReducedMotion(): boolean | null {
  return useSyncExternalStore(
    (onChange) => {
      const query = window.matchMedia(REDUCED_MOTION_QUERY);
      query.addEventListener('change', onChange);
      return () => query.removeEventListener('change', onChange);
    },
    () => prefersReducedMotion(window.matchMedia.bind(window)),
    () => null,
  );
}

let webglProbe: boolean | undefined;

/** Whether WebGL works here, probed once per page load; null during the server render. */
export function useWebGLSupport(): boolean | null {
  return useSyncExternalStore(
    noSubscription,
    () => (webglProbe ??= supportsWebGL(() => document.createElement('canvas'))),
    () => null,
  );
}

/** Whether the page is visible (not a background tab). */
export function usePageVisible(): boolean {
  return useSyncExternalStore(
    (onChange) => {
      document.addEventListener('visibilitychange', onChange);
      return () => document.removeEventListener('visibilitychange', onChange);
    },
    () => document.visibilityState === 'visible',
    () => true,
  );
}

/** Whether `target` intersects the viewport. Assumed visible until the first report. */
export function useInView(target: RefObject<Element | null>): boolean {
  const [inView, setInView] = useState(true);
  useEffect(() => {
    const element = target.current;
    if (!element || typeof IntersectionObserver === 'undefined') return;
    const observer = new IntersectionObserver(([entry]) => {
      if (entry) setInView(entry.isIntersecting);
    });
    observer.observe(element);
    return () => observer.disconnect();
  }, [target]);
  return inView;
}

/**
 * Tracks the pointer relative to the centre of `target`, normalised to the
 * viewport (-1..1, +y down). Stored in a ref so pointer moves never re-render.
 */
export function usePointerTarget(target: RefObject<Element | null>): RefObject<LookTarget | null> {
  const look = useRef<LookTarget | null>(null);
  useEffect(() => {
    const onMove = (event: PointerEvent) => {
      const element = target.current;
      if (!element) return;
      const box = element.getBoundingClientRect();
      look.current = {
        x: clamp((event.clientX - (box.left + box.width / 2)) / (window.innerWidth / 2), -1, 1),
        y: clamp((event.clientY - (box.top + box.height * 0.3)) / (window.innerHeight / 2), -1, 1),
      };
    };
    const onLeave = () => {
      look.current = null;
    };
    window.addEventListener('pointermove', onMove, { passive: true });
    document.documentElement.addEventListener('pointerleave', onLeave);
    return () => {
      window.removeEventListener('pointermove', onMove);
      document.documentElement.removeEventListener('pointerleave', onLeave);
    };
  }, [target]);
  return look;
}

/**
 * A wave that plays for `durationMs` on mount and again whenever `restart`
 * is called (for example on hover).
 */
export function useWave(durationMs: number): { waving: boolean; restart: () => void } {
  const [waving, setWaving] = useState(true);
  const timer = useRef<number | undefined>(undefined);
  const schedule = useCallback(() => {
    window.clearTimeout(timer.current);
    timer.current = window.setTimeout(() => setWaving(false), durationMs);
  }, [durationMs]);
  useEffect(() => {
    schedule();
    return () => window.clearTimeout(timer.current);
  }, [schedule]);
  const restart = useCallback(() => {
    setWaving(true);
    schedule();
  }, [schedule]);
  return { waving, restart };
}
