'use client';
/**
 * Motion for a page, loaded lazily. Components use the lightweight `m`
 * elements; their animation features (./features) are fetched after the
 * first render, so Motion adds almost nothing to the JavaScript a page needs
 * before it can paint. `strict` turns any full `motion.*` element, which
 * would pull every feature back into the page bundle, into an error. Motion
 * honours the visitor's reduced-motion setting.
 */
import { LazyMotion, MotionConfig } from 'motion/react';
import type { ReactNode } from 'react';

const loadFeatures = () => import('./features').then((features) => features.default);

export function SiteMotion({ children }: { children: ReactNode }) {
  return (
    <LazyMotion features={loadFeatures} strict>
      <MotionConfig reducedMotion="user">{children}</MotionConfig>
    </LazyMotion>
  );
}
