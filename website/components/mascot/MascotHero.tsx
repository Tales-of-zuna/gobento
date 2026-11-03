'use client';
/**
 * The mascot as it appears on the page. The pre-rendered still is in the
 * HTML from the start (sized, so nothing shifts); the 3D scene is downloaded
 * once the visitor first interacts with the page and cross-fades in once its
 * first frame is drawn. The still stays when WebGL is missing, the scene
 * fails, reduced motion is on, or nobody interacts.
 */
import dynamic from 'next/dynamic';
import { useEffect, useState } from 'react';
import { cn } from '@/lib/utils';
import { useHasInteracted, useReducedMotion, useWebGLSupport } from './hooks';
import { MascotErrorBoundary } from './MascotErrorBoundary';
import type { Packed } from './parts/bento';
import { HERO_SIZE, HERO_STILLS, heroSrcSet, MASCOT_DESCRIPTION, stillUrl } from './stills';
import { shouldMountScene } from './support';

const MascotScene = dynamic(() => import('./MascotScene'), { ssr: false });

/** Props of {@link MascotHero}. */
export interface MascotHeroProps {
  /** Sizing of the 3:2 box; by default it fills the width of its container. */
  readonly className?: string;
  /** The still's displayed width (the `sizes` attribute), so it loads the right render. */
  readonly sizes: string;
  /** Which food compartments the live scene shows filled (the still has all four). */
  readonly packed?: Packed;
  /**
   * Leave out the text alternative and hide the mascot from assistive
   * technology, for pages where nearby content already says the same.
   */
  readonly decorative?: boolean;
  /** Told whenever the live scene replaces the still, or the still comes back. */
  readonly onSceneChange?: (showing: boolean) => void;
}

export function MascotHero({
  className,
  sizes,
  packed,
  decorative = false,
  onSceneChange,
}: MascotHeroProps) {
  const reducedMotion = useReducedMotion();
  const webgl = useWebGLSupport();
  const interacted = useHasInteracted();
  const [ready, setReady] = useState(false);
  const [failed, setFailed] = useState(false);
  const mountScene = shouldMountScene({ webgl, reducedMotion, failed, ready, interacted });
  const showScene = mountScene && ready;
  const fail = () => setFailed(true);

  useEffect(() => onSceneChange?.(showScene), [onSceneChange, showScene]);

  return (
    <div
      className={cn('relative aspect-[3/2]', className ?? 'w-full')}
      data-mascot={showScene ? 'scene' : 'still'}
      aria-hidden={decorative || undefined}
    >
      {!decorative && <span className="sr-only">{MASCOT_DESCRIPTION}</span>}
      <picture>
        <source type="image/webp" srcSet={heroSrcSet('webp')} sizes={sizes} />
        <img
          src={stillUrl(HERO_STILLS[0], 'png')}
          srcSet={heroSrcSet('png')}
          sizes={sizes}
          width={HERO_SIZE.width}
          height={HERO_SIZE.height}
          alt=""
          decoding="async"
          data-mascot-still
          className={cn(
            'absolute inset-0 size-full select-none transition-opacity duration-500 motion-reduce:transition-none',
            showScene && 'opacity-0',
          )}
        />
      </picture>
      {mountScene && (
        <div
          className={cn(
            'absolute inset-0 opacity-0 transition-opacity duration-500 motion-reduce:transition-none',
            showScene && 'opacity-100',
          )}
        >
          <MascotErrorBoundary onError={fail}>
            <MascotScene
              animate={!reducedMotion}
              packed={packed}
              onReady={() => setReady(true)}
              onError={fail}
            />
          </MascotErrorBoundary>
        </div>
      )}
    </div>
  );
}
