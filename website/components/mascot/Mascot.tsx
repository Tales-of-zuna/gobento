'use client';
/**
 * The gobento mascot, concept 3 "Box body": a sora-blue gopher whose body is
 * a bento box turned on its side. Base at y = 0, facing +Z.
 *
 * All animation runs in one `useFrame` callback and is driven by elapsed
 * frame time, so it looks the same at any frame rate. With `animate` off the
 * mascot holds the approved pose, which is also what the static stills show.
 */
import { useFrame } from '@react-three/fiber';
import { useRef, type RefObject } from 'react';
import { MathUtils, type Group } from 'three';
import { BentoBody, BOX, type Packed } from './parts/bento';
import { GopherHead } from './parts/face';
import { Arm, Foot } from './parts/limbs';
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
  /** Wave the raised arm while true; it eases in and out. */
  readonly wave?: boolean;
  /** Blink now and then. */
  readonly blink?: boolean;
  /** Where the pointer is, read every frame so moving it never re-renders React. */
  readonly lookAt?: RefObject<LookTarget | null>;
  /** Which compartments show food; a compartment that fills drops its food in. */
  readonly packed?: Packed;
}

const DEG = Math.PI / 180;
const FEET_HEIGHT = 0.28;
const TORSO_Y = FEET_HEIGHT + BOX.height / 2;
const HEAD_Y = FEET_HEIGHT + BOX.height + 0.62;
const HEAD_TILT = -5 * DEG;
const SHOULDER_Y = FEET_HEIGHT + BOX.height * 0.74;
const SHOULDER_X = BOX.width / 2 + 0.02;
/** The right arm is raised in a hello; the left hangs relaxed. */
const WAVE_ARM_REST = { x: -10 * DEG, z: 140 * DEG };
const RELAXED_ARM_REST = { x: -15 * DEG, z: -22 * DEG };

/** Idle breathing period, seconds. */
const BOB_PERIOD = 2.6;
const BOB_HEIGHT = 0.05;
/** Waves per second and swing (radians) of the hello. */
const WAVE_RATE = 1.7;
const WAVE_SWING = 0.34;

/** Animation state that lives across frames but never needs a React render. */
interface Clock {
  time: number;
  nextBlink: number;
  blinkStart: number;
  waveAmount: number;
  yaw: number;
  pitch: number;
}

/** The mascot, ready to drop into an R3F scene. */
export function Mascot({
  color = PALETTE.sora,
  animate = true,
  wave = false,
  blink = true,
  lookAt,
  packed = ALL_PACKED,
}: MascotProps) {
  const upperBody = useRef<Group>(null);
  const head = useRef<Group>(null);
  const leftEye = useRef<Group>(null);
  const rightEye = useRef<Group>(null);
  const waveArm = useRef<Group>(null);
  const relaxedArm = useRef<Group>(null);
  const clock = useRef<Clock>({
    time: 0,
    nextBlink: 1.2,
    blinkStart: -1,
    waveAmount: 0,
    yaw: 0,
    pitch: 0,
  });

  useFrame((_, rawDelta) => {
    const c = clock.current;
    // A long gap (tab in the background, paused canvas) must not cause a jump.
    const delta = Math.min(rawDelta, 0.1);
    if (animate) c.time += delta;
    else Object.assign(c, { time: 0, blinkStart: -1, waveAmount: 0, yaw: 0, pitch: 0 });
    const t = c.time;

    // Idle bob: the whole upper body breathes up and down; the feet stay planted.
    const breath = animate ? 0.5 - 0.5 * Math.cos((2 * Math.PI * t) / BOB_PERIOD) : 0;
    upperBody.current?.position.setY(breath * BOB_HEIGHT);

    // Blink at random intervals.
    let closure = 0;
    if (animate && blink) {
      if (c.blinkStart < 0 && t >= c.nextBlink) c.blinkStart = t;
      if (c.blinkStart >= 0) {
        closure = blinkClosure(t - c.blinkStart);
        if (closure === 0 && t - c.blinkStart > 0) {
          c.blinkStart = -1;
          c.nextBlink = t + nextBlinkDelay(Math.random());
        }
      }
    }
    const eyeScale = 1 - closure * 0.9;
    leftEye.current?.scale.setY(eyeScale);
    rightEye.current?.scale.setY(eyeScale);

    // Gentle head turn towards the pointer, eased so it never snaps.
    const target = animate ? headTurn(lookAt?.current ?? null) : { yaw: 0, pitch: 0 };
    c.yaw = MathUtils.damp(c.yaw, target.yaw, 4, delta);
    c.pitch = MathUtils.damp(c.pitch, target.pitch, 4, delta);
    const sway = animate ? Math.sin((2 * Math.PI * t) / BOB_PERIOD - 0.6) * 0.03 : 0;
    head.current?.rotation.set(c.pitch, c.yaw, HEAD_TILT + sway);

    // Wave: the raised arm swings while `wave` is on, easing in and out.
    c.waveAmount = MathUtils.damp(c.waveAmount, animate && wave ? 1 : 0, 5, delta);
    const swing = Math.sin(2 * Math.PI * WAVE_RATE * t) * WAVE_SWING * c.waveAmount;
    waveArm.current?.rotation.set(WAVE_ARM_REST.x, 0, WAVE_ARM_REST.z + swing);
    relaxedArm.current?.rotation.set(RELAXED_ARM_REST.x, 0, RELAXED_ARM_REST.z - breath * 0.06);
  });

  return (
    <group>
      <group ref={upperBody}>
        <BentoBody packed={packed} animate={animate} position={[0, TORSO_Y, 0]} />
        <group ref={head} position={[0, HEAD_Y, 0]} rotation={[0, 0, HEAD_TILT]}>
          <GopherHead color={color} leftEye={leftEye} rightEye={rightEye} />
        </group>
        <Arm
          ref={waveArm}
          color={color}
          position={[SHOULDER_X, SHOULDER_Y, 0]}
          rotation={[WAVE_ARM_REST.x, 0, WAVE_ARM_REST.z]}
        />
        <Arm
          ref={relaxedArm}
          color={color}
          position={[-SHOULDER_X, SHOULDER_Y, 0]}
          rotation={[RELAXED_ARM_REST.x, 0, RELAXED_ARM_REST.z]}
        />
      </group>
      {([-1, 1] as const).map((side) => (
        <Foot
          key={side}
          color={color}
          position={[side * 0.56, 0.17, 0.34]}
          rotation={[0, side * 10 * DEG, 0]}
          scale={1.3}
        />
      ))}
    </group>
  );
}
