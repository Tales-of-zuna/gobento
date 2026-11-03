'use client';
/**
 * The live mascot scene, loaded lazily (never on the server). It renders only
 * while on screen and in a visible tab, waves on mount and on hover, and
 * follows the pointer with its head. With `animate` off it holds a static
 * pose and draws only when something changes.
 */
import { useRef } from 'react';
import { useInView, usePageVisible, usePointerTarget, useWave } from './hooks';
import { MascotCanvas } from './MascotCanvas';
import type { Packed } from './parts/bento';

/** Props of {@link MascotScene}. */
export interface MascotSceneProps {
  /** False for reduced motion: one static pose, drawn on demand. */
  readonly animate: boolean;
  readonly packed?: Packed;
  /** The first frame is on screen; the still can step aside. */
  readonly onReady: () => void;
  /** The WebGL context was lost; fall back to the still. */
  readonly onError: () => void;
}

const WAVE_MS = 2400;

export default function MascotScene({ animate, packed, onReady, onError }: MascotSceneProps) {
  const container = useRef<HTMLDivElement>(null);
  const inView = useInView(container);
  const pageVisible = usePageVisible();
  const lookAt = usePointerTarget(container);
  const { waving, restart } = useWave(WAVE_MS);

  const frameloop = !animate ? 'demand' : inView && pageVisible ? 'always' : 'never';

  return (
    <div ref={container} className="absolute inset-0" onPointerEnter={restart} aria-hidden>
      <MascotCanvas
        frameloop={frameloop}
        animate={animate}
        wave={waving}
        lookAt={lookAt}
        packed={packed}
        onReady={onReady}
        onContextLost={onError}
      />
    </div>
  );
}
