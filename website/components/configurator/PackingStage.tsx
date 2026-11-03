'use client';
/**
 * The "packing" visual: a 2D bento tray with one compartment per component
 * kind, filling as each kind is chosen, and above it the mascot whose
 * box-body compartments pack in step. The mascot follows the home page
 * hero's rules (MascotHero): its pre-rendered still holds the space from the
 * first paint, so nothing shifts, and the 3D scene is downloaded only on the
 * client, after the first interaction, when WebGL works and reduced motion is
 * off. The tray is the text alternative, so the mascot is decorative here.
 */
import { useReducedMotion as useMotionReduced } from 'motion/react';
import * as m from 'motion/react-m';
import { useState } from 'react';
import { MascotHero } from '@/components/mascot/MascotHero';
import styles from './configurator.module.css';
import type { CellView } from './view-model';

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
  const [scene, setScene] = useState(false);
  const still = useMotionReduced();

  return (
    <div className={styles.stage} data-stage={scene ? 'scene' : 'tray'}>
      <span className={`${styles.sticker} ${styles.sOni}`} aria-hidden />
      <span className={`${styles.sticker} ${styles.sTam}`} aria-hidden />
      <span className={`${styles.sticker} ${styles.sUme}`} aria-hidden />
      <span className={`${styles.sticker} ${styles.sSak}`} aria-hidden />
      <div className={styles.scene}>
        <MascotHero
          className={styles.sceneMascot}
          sizes="(max-width: 899px) 300px, 390px"
          packed={packed}
          decorative
          onSceneChange={setScene}
        />
      </div>
      <ul className={styles.tray} aria-label="Bento: chosen parts">
        {cells.map((cell, i) => {
          const filled = packed[i] === true && cell.value !== undefined;
          return (
            <li key={cell.kind} className={styles.cell} data-kind={cell.kind} data-filled={filled}>
              <span className={styles.ck}>{cell.label}</span>
              {filled ? (
                <>
                  <m.span
                    className={styles.piece}
                    data-piece={cell.piece}
                    aria-hidden
                    initial={still ? false : { opacity: 0, y: -22, scale: 0.6 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    transition={pop}
                  />
                  <m.b
                    className={styles.cv}
                    initial={still ? false : { opacity: 0, y: 4 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.3 }}
                  >
                    {cell.value}
                  </m.b>
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
