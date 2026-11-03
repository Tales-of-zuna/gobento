'use client';
/**
 * The gobento mascot, concept 1 "In the box": a sora-blue gopher sunk into
 * the big compartment of an open bento box, paws over the rim, with the lid
 * leaning behind as a backrest and four compartments of food beside it.
 * Box base at y = 0, facing +Z.
 *
 * All animation runs in one `useFrame` callback and is driven by elapsed
 * frame time, so it looks the same at any frame rate. With `animate` off the
 * mascot holds the approved pose, which is also what the static stills show.
 */
import { useFrame } from '@react-three/fiber';
import { useRef, type RefObject } from 'react';
import { MathUtils, type Group } from 'three';
import { BentoBox, Lid, LID_POSE, type Packed } from './parts/bento';
import { Gopher, PAW } from './parts/gopher';
import { blinkClosure, headTurn, nextBlinkDelay, type LookTarget } from './math';
import { PALETTE } from './toon';

/** Every compartment filled, as in the approved design. */
export const ALL_PACKED: Packed = [true, true, true, true];

/** Props of {@link Mascot}. */
export interface MascotProps {
  /** Fur colour. Defaults to the approved sora blue. */
  readonly color?: string;
  /** Run the idle animation. False holds the approved static pose. */
  readonly animate?: boolean;
  /** Wave the right paw while true; it eases in and out. */
  readonly wave?: boolean;
  /** Lift the lid a little while true (for hover). */
  readonly lidLift?: boolean;
  /** Blink now and then. */
  readonly blink?: boolean;
  /** Where the pointer is, read every frame so moving it never re-renders React. */
  readonly lookAt?: RefObject<LookTarget | null>;
  /** Which food compartments are filled; one that fills drops its food in. */
  readonly packed?: Packed;
}

const DEG = Math.PI / 180;
/** The gopher sits low in the left compartment, chin just above the rim, leaning a little. */
const GOPHER_POSE = {
  position: [-0.68, -0.32, -0.12] as const,
  rotation: { x: -6 * DEG, y: 8 * DEG, z: 5 * DEG },
};

/** Idle breathing period (seconds) and how far the gopher rises in the box. */
const BOB_PERIOD = 2.8;
const BOB_HEIGHT = 0.06;
/** Waves per second, swing (radians), and how far the paw rises and opens out to wave. */
const WAVE_RATE = 1.6;
const WAVE_SWING = 0.35;
const WAVE_RAISE = 50 * DEG;
const WAVE_OPEN = 40 * DEG;
/** Lid lift on hover: extra tilt and rise. */
const LID_TILT = 6 * DEG;
const LID_RISE = 0.1;

/** Animation state that lives across frames but never needs a React render. */
interface Clock {
  time: number;
  nextBlink: number;
  blinkStart: number;
  waveAmount: number;
  lidAmount: number;
  yaw: number;
  pitch: number;
}

const REST: Omit<Clock, 'nextBlink'> = {
  time: 0,
  blinkStart: -1,
  waveAmount: 0,
  lidAmount: 0,
  yaw: 0,
  pitch: 0,
};

/** The mascot, ready to drop into an R3F scene. */
export function Mascot({
  color = PALETTE.sora,
  animate = true,
  wave = false,
  lidLift = false,
  blink = true,
  lookAt,
  packed = ALL_PACKED,
}: MascotProps) {
  const gopher = useRef<Group>(null);
  const face = useRef<Group>(null);
  const leftEye = useRef<Group>(null);
  const rightEye = useRef<Group>(null);
  const wavePaw = useRef<Group>(null);
  const lid = useRef<Group>(null);
  const clock = useRef<Clock>({ ...REST, nextBlink: 1.2 });

  useFrame((_, rawDelta) => {
    const c = clock.current;
    // A long gap (tab in the background, paused canvas) must not cause a jump.
    const delta = Math.min(rawDelta, 0.1);
    if (animate) c.time += delta;
    else Object.assign(c, REST);
    const t = c.time;

    // Idle bob: the gopher breathes up and down inside its compartment.
    const breath = animate ? 0.5 - 0.5 * Math.cos((2 * Math.PI * t) / BOB_PERIOD) : 0;

    // Blink at random intervals.
    let closure = 0;
    if (animate && blink) {
      if (c.blinkStart < 0 && t >= c.nextBlink) c.blinkStart = t;
      if (c.blinkStart >= 0) {
        closure = blinkClosure(t - c.blinkStart);
        if (closure === 0 && t > c.blinkStart) {
          c.blinkStart = -1;
          c.nextBlink = t + nextBlinkDelay(Math.random());
        }
      }
    }
    const eyeScale = 1 - closure * 0.9;
    leftEye.current?.scale.setY(eyeScale);
    rightEye.current?.scale.setY(eyeScale);

    // Look towards the pointer: the face slides around the round body, and
    // the body leans a little forward or back. Eased so it never snaps.
    const target = animate ? headTurn(lookAt?.current ?? null) : { yaw: 0, pitch: 0 };
    c.yaw = MathUtils.damp(c.yaw, target.yaw, 4, delta);
    c.pitch = MathUtils.damp(c.pitch, target.pitch, 4, delta);
    face.current?.rotation.set(0, c.yaw, 0);
    const sway = animate ? Math.sin((2 * Math.PI * t) / BOB_PERIOD - 0.6) * 0.02 : 0;
    gopher.current?.position.set(
      GOPHER_POSE.position[0],
      GOPHER_POSE.position[1] + breath * BOB_HEIGHT,
      GOPHER_POSE.position[2],
    );
    gopher.current?.rotation.set(
      GOPHER_POSE.rotation.x + c.pitch * 0.5,
      GOPHER_POSE.rotation.y,
      GOPHER_POSE.rotation.z + sway,
    );

    // Wave: the right paw rises off the rim and swings while `wave` is on.
    c.waveAmount = MathUtils.damp(c.waveAmount, animate && wave ? 1 : 0, 5, delta);
    const swing = Math.sin(2 * Math.PI * WAVE_RATE * t) * WAVE_SWING * c.waveAmount;
    wavePaw.current?.rotation.set(
      PAW.rest.x - WAVE_RAISE * c.waveAmount,
      0,
      PAW.rest.z + WAVE_OPEN * c.waveAmount + swing,
    );

    // Lid: lifts and tilts up a touch while hovered.
    c.lidAmount = MathUtils.damp(c.lidAmount, animate && lidLift ? 1 : 0, 6, delta);
    lid.current?.position.setY(LID_POSE.hinge[1] + LID_RISE * c.lidAmount);
    lid.current?.rotation.set(LID_POSE.tilt + LID_TILT * c.lidAmount, 0, 0);
  });

  return (
    <group>
      <BentoBox packed={packed} animate={animate} />
      <Lid ref={lid} />
      <Gopher
        ref={gopher}
        color={color}
        face={face}
        leftEye={leftEye}
        rightEye={rightEye}
        wavePaw={wavePaw}
        position={[...GOPHER_POSE.position]}
        rotation={[GOPHER_POSE.rotation.x, GOPHER_POSE.rotation.y, GOPHER_POSE.rotation.z]}
      />
    </group>
  );
}
