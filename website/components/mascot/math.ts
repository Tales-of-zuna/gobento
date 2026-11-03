/**
 * Pure helpers for placing and animating the mascot. Nothing here touches
 * React or WebGL, so every function is unit-tested on its own.
 */

/** A 3-component vector as a plain tuple. */
export type Vec3 = readonly [number, number, number];
/** A quaternion as a plain `[x, y, z, w]` tuple. */
export type Quat = readonly [number, number, number, number];

/** Position and orientation for a feature stuck onto a surface. */
export interface SurfacePose {
  readonly position: Vec3;
  readonly quaternion: Quat;
}

function normalize([x, y, z]: Vec3): Vec3 {
  const length = Math.hypot(x, y, z);
  return [x / length, y / length, z / length];
}

/**
 * Rotation that turns +Z onto the unit vector `to` (the shortest arc), so a
 * feature modelled facing +Z ends up facing along `to`.
 */
export function quaternionFromZ([x, y, z]: Vec3): Quat {
  const w = 1 + z;
  if (w < 1e-6) return [1, 0, 0, 0];
  const q: Quat = [-y, x, 0, w];
  const length = Math.hypot(...q);
  return [q[0] / length, q[1] / length, q[2] / length, q[3] / length];
}

/**
 * Pose for a feature on an ellipsoid centred at the origin with the given
 * radii: the surface point along `direction`, moved `push` along the ray, with
 * local +Z facing outward along the ray. Facing the ray rather than the true
 * surface normal is deliberate; it matches the approved prototype, where
 * features fan out radially from the centre of the head.
 */
export function ellipsoidPose(radii: Vec3, direction: Vec3, push = 0): SurfacePose {
  const d = normalize(direction);
  const distance = 1 / Math.hypot(d[0] / radii[0], d[1] / radii[1], d[2] / radii[2]);
  const reach = distance + push;
  return {
    position: [d[0] * reach, d[1] * reach, d[2] * reach],
    quaternion: quaternionFromZ(d),
  };
}

/** Deterministic PRNG (mulberry32), so scattered food looks the same on every render. */
export function mulberry32(seed: number): () => number {
  let a = seed | 0;
  return () => {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

/** Clamps `value` into `[min, max]`. */
export function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

/** State of a one-dimensional spring. */
export interface SpringState {
  readonly position: number;
  readonly velocity: number;
}

/** Spring constants; the defaults give one soft overshoot, like a dropped piece settling. */
export interface SpringConfig {
  readonly stiffness: number;
  readonly damping: number;
}

export const DROP_SPRING: SpringConfig = { stiffness: 170, damping: 14 };

/** Fixed integration step, so the spring behaves the same at any frame rate. */
const SPRING_STEP = 1 / 240;

/**
 * Advances a unit-mass spring towards `target` by `dt` seconds, using fixed
 * semi-implicit Euler sub-steps so the result does not depend on frame rate.
 */
export function stepSpring(
  state: SpringState,
  target: number,
  dt: number,
  { stiffness, damping }: SpringConfig = DROP_SPRING,
): SpringState {
  let { position, velocity } = state;
  let remaining = Math.min(Math.max(dt, 0), 0.25);
  while (remaining > 1e-9) {
    const h = Math.min(SPRING_STEP, remaining);
    velocity += (-stiffness * (position - target) - damping * velocity) * h;
    position += velocity * h;
    remaining -= h;
  }
  return { position, velocity };
}

/** Whether a spring is close enough to `target` to stop animating. */
export function springAtRest(state: SpringState, target: number): boolean {
  return Math.abs(state.position - target) < 1e-3 && Math.abs(state.velocity) < 1e-3;
}

/** Duration of one blink in seconds. */
export const BLINK_DURATION = 0.16;

/**
 * How closed the eyes are (0 open, 1 shut) `t` seconds into a blink: a quick
 * close and a slightly slower open.
 */
export function blinkClosure(t: number): number {
  if (t <= 0 || t >= BLINK_DURATION) return 0;
  const closeTime = BLINK_DURATION * 0.4;
  return t < closeTime ? t / closeTime : 1 - (t - closeTime) / (BLINK_DURATION - closeTime);
}

/** Seconds until the next blink, between 2.2 and 5.2, from a random number in [0, 1). */
export function nextBlinkDelay(random: number): number {
  return 2.2 + random * 3;
}

/** Pointer position normalised to the mascot: -1..1 on each axis, +y down. */
export interface LookTarget {
  readonly x: number;
  readonly y: number;
}

/** Largest head turn towards the pointer, in radians. */
export const MAX_YAW = 0.38;
export const MAX_PITCH = 0.16;

/** Head yaw and pitch (radians) for a pointer target, clamped to a gentle range. */
export function headTurn(target: LookTarget | null): { yaw: number; pitch: number } {
  if (!target) return { yaw: 0, pitch: 0 };
  return {
    yaw: clamp(target.x, -1, 1) * MAX_YAW,
    pitch: clamp(target.y, -1, 1) * MAX_PITCH,
  };
}
