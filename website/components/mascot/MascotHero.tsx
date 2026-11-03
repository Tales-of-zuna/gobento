'use client';
/**
 * The mascot as it appears on the page. The pre-rendered still is in the
 * HTML from the start (sized, so nothing shifts); the 3D scene is downloaded
 * afterwards and cross-fades in once its first frame is drawn. The still
 * stays when WebGL is missing, the scene fails, or reduced motion is on.
 */
import dynamic from 'next/dynamic';
import { useState } from 'react';
import { cn } from '@/lib/utils';
import { useReducedMotion, useWebGLSupport } from './hooks';
import { MascotErrorBoundary } from './MascotErrorBoundary';
import { HERO_SIZE, HERO_STILLS, heroSrcSet, MASCOT_DESCRIPTION, stillUrl } from './stills';
import { shouldMountScene } from './support';

const MascotScene = dynamic(() => import('./MascotScene'), { ssr: false });

/** Props of {@link MascotHero}. */
export interface MascotHeroProps {
  readonly className?: string;
}

export function MascotHero({ className }: MascotHeroProps) {
  const reducedMotion = useReducedMotion();
  const webgl = useWebGLSupport();
  const [ready, setReady] = useState(false);
  const [failed, setFailed] = useState(false);
  const mountScene = shouldMountScene({ webgl, reducedMotion, failed, ready });
  const showScene = mountScene && ready;
  const fail = () => setFailed(true);

  return (
    <div
      className={cn('relative aspect-[3/2] w-full', className)}
      data-mascot={showScene ? 'scene' : 'still'}
    >
      <span className="sr-only">{MASCOT_DESCRIPTION}</span>
      <picture>
        <source type="image/webp" srcSet={heroSrcSet('webp')} />
        <img
          src={stillUrl(HERO_STILLS[0], 'png')}
          srcSet={heroSrcSet('png')}
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
            <MascotScene animate={!reducedMotion} onReady={() => setReady(true)} onError={fail} />
          </MascotErrorBoundary>
        </div>
      )}
    </div>
  );
}
