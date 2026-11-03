'use client';
/**
 * The "packing" visual: a 2D bento tray with one compartment per component
 * kind, filling as each kind is chosen, and above it the 3D mascot whose
 * box-body compartments pack in step. The 3D scene follows the same
 * safeguards as the home page hero: its code is only downloaded on the
 * client, when WebGL works and reduced motion is off; a failure or lost
 * context leaves the tray alone, which is also the text alternative.
 */
import { motion, useReducedMotion as useMotionReduced } from 'motion/react';
import dynamic from 'next/dynamic';
import { useState } from 'react';
import { useReducedMotion, useWebGLSupport } from '@/components/mascot/hooks';
import { MascotErrorBoundary } from '@/components/mascot/MascotErrorBoundary';
import { shouldMountScene } from '@/components/mascot/support';
import styles from './configurator.module.css';
import type { CellView } from './view-model';

const MascotScene = dynamic(() => import('@/components/mascot/MascotScene'), { ssr: false });

/** Props of {@link PackingStage}. */
export interface PackingStageProps {
  readonly cells: readonly CellView[];
  /** Which compartments show their food, in tray order. */
  readonly packed: readonly [boolean, boolean, boolean, boolean];
  /** Replays the packing animation. */
  readonly onRepack: () => void;
}

const pop = { type: 'spring', stiffness: 420, damping: 14 } as const;

export function PackingStage({ cells, packed, onRepack }: PackingStageProps) {
  const reducedMotion = useReducedMotion();
  const webgl = useWebGLSupport();
  const [ready, setReady] = useState(false);
  const [failed, setFailed] = useState(false);
  const mountScene = shouldMountScene({ webgl, reducedMotion, failed, ready });
  const fail = () => setFailed(true);
  const still = useMotionReduced();

  return (
    <div className={styles.stage} data-stage={mountScene && ready ? 'scene' : 'tray'}>
      <span className={`${styles.sticker} ${styles.sOni}`} aria-hidden />
      <span className={`${styles.sticker} ${styles.sTam}`} aria-hidden />
      <span className={`${styles.sticker} ${styles.sUme}`} aria-hidden />
      <span className={`${styles.sticker} ${styles.sSak}`} aria-hidden />
      {mountScene && (
        <div className={styles.scene} data-ready={ready}>
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
      <ul className={styles.tray} aria-label="Bento: chosen parts">
        {cells.map((cell, i) => {
          const filled = packed[i] === true && cell.value !== undefined;
          return (
            <li key={cell.kind} className={styles.cell} data-kind={cell.kind} data-filled={filled}>
              <span className={styles.ck}>{cell.label}</span>
              {filled ? (
                <>
                  <motion.span
                    className={styles.piece}
                    data-piece={cell.piece}
                    aria-hidden
                    initial={still ? false : { opacity: 0, y: -22, scale: 0.6 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    transition={pop}
                  />
                  <motion.b
                    className={styles.cv}
                    initial={still ? false : { opacity: 0, y: 4 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.3 }}
                  >
                    {cell.value}
                  </motion.b>
                  <span className={styles.food} aria-hidden>
                    {cell.food}
                  </span>
                </>
              ) : (
                <span className={styles.ce}>empty</span>
              )}
            </li>
          );
        })}
      </ul>
      <div className={styles.stageRow}>
        <p className={styles.stageNote}>Each choice packs a compartment.</p>
        <button type="button" className={styles.ghost} onClick={onRepack}>
          Repack
        </button>
      </div>
    </div>
  );
}
