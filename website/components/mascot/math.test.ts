import { Quaternion, Vector3 } from 'three';
import { describe, expect, it } from 'vitest';
import {
  BLINK_DURATION,
  blinkClosure,
  daifukuProfile,
  ellipsoidPose,
  headTurn,
  latheRadiusAt,
  lathePose,
  MAX_PITCH,
  MAX_YAW,
  mulberry32,
  nextBlinkDelay,
  quaternionFromZ,
  springAtRest,
  stepSpring,
  type SpringState,
  type Vec3,
} from './math';

const rotateZ = (q: readonly number[]) =>
  new Vector3(0, 0, 1).applyQuaternion(new Quaternion(q[0], q[1], q[2], q[3]));

describe('quaternionFromZ', () => {
  it.each<Vec3>([
    [0, 0, 1],
    [1, 0, 0],
    [0, 1, 0],
    [0.36, 0.08, 1],
    [0, 0, -1],
  ])('turns +Z onto %j', (...target) => {
    const unit = new Vector3(...target).normalize();
    const result = rotateZ(quaternionFromZ([unit.x, unit.y, unit.z]));
    expect(result.distanceTo(unit)).toBeLessThan(1e-9);
  });
});

describe('ellipsoidPose', () => {
  it('lands on the surface of a sphere', () => {
    const { position } = ellipsoidPose([2, 2, 2], [1, 1, 0]);
    expect(Math.hypot(...position)).toBeCloseTo(2, 9);
  });

  it('lands on the surface of an ellipsoid', () => {
    const radii: Vec3 = [1.0584, 0.9212, 0.98];
    const [x, y, z] = ellipsoidPose(radii, [0.36, 0.12, 1]).position;
    expect((x / radii[0]) ** 2 + (y / radii[1]) ** 2 + (z / radii[2]) ** 2).toBeCloseTo(1, 9);
    // Matches the raycast placement of the approved prototype.
    expect([x, y, z].map((n) => n.toFixed(2))).toEqual(['0.33', '0.11', '0.92']);
  });

  it('pushes along the ray and faces outward', () => {
    const { position, quaternion } = ellipsoidPose([1, 1, 1], [0, 0, 1], -0.1);
    expect(position[2]).toBeCloseTo(0.9, 9);
    expect(rotateZ(quaternion).z).toBeCloseTo(1, 9);
  });
});

describe('daifuku body', () => {
  const profile = daifukuProfile(2.6, 1.18);

  it('stands on y = 0 and is widest below the middle', () => {
    expect(Math.min(...profile.height)).toBeCloseTo(0, 9);
    const widest = profile.radius.indexOf(Math.max(...profile.radius));
    expect(profile.height[widest]).toBeLessThan(Math.max(...profile.height) / 2);
  });

  it('interpolates the radius and is empty outside the profile', () => {
    expect(latheRadiusAt(profile, -0.1)).toBe(0);
    expect(latheRadiusAt(profile, 10)).toBe(0);
    expect(latheRadiusAt(profile, 1.2)).toBeGreaterThan(1);
  });

  it('places a feature on the surface where the prototype raycast did', () => {
    const { position, quaternion } = lathePose(profile, [0, 1.74, 0], [0.36, 0.12, 1]);
    expect(position.map((n) => n.toFixed(2))).toEqual(['0.32', '1.85', '0.89']);
    expect(Math.hypot(position[0], position[2])).toBeCloseTo(
      latheRadiusAt(profile, position[1]),
      3,
    );
    // Faces outward and a little up, like the surface there.
    const facing = rotateZ(quaternion);
    expect(facing.z).toBeGreaterThan(0.7);
    expect(facing.y).toBeGreaterThan(0.3);
  });
});

describe('mulberry32', () => {
  it('is deterministic per seed and stays in [0, 1)', () => {
    const a = mulberry32(7);
    const b = mulberry32(7);
    const values = Array.from({ length: 100 }, () => a());
    expect(values).toEqual(Array.from({ length: 100 }, () => b()));
    expect(values.every((v) => v >= 0 && v < 1)).toBe(true);
    expect(mulberry32(8)()).not.toBe(values[0]);
  });
});

describe('stepSpring', () => {
  const settle = (frameRate: number, seconds: number): SpringState => {
    let state: SpringState = { position: 1, velocity: 0 };
    for (let i = 0; i < frameRate * seconds; i++) state = stepSpring(state, 0, 1 / frameRate);
    return state;
  };

  it('settles at the target', () => {
    expect(springAtRest(settle(60, 3), 0)).toBe(true);
  });

  it('overshoots once, like a dropped piece bouncing', () => {
    let state: SpringState = { position: 1, velocity: 0 };
    let lowest = 1;
    for (let i = 0; i < 120; i++) {
      state = stepSpring(state, 0, 1 / 60);
      lowest = Math.min(lowest, state.position);
    }
    expect(lowest).toBeLessThan(0);
  });

  it('does not depend on the frame rate', () => {
    const at30 = settle(30, 0.3);
    const at144 = settle(144, 0.3);
    expect(at30.position).toBeCloseTo(at144.position, 2);
  });

  it('ignores negative time and caps long gaps', () => {
    const start: SpringState = { position: 1, velocity: 0 };
    expect(stepSpring(start, 0, -1)).toEqual(start);
    expect(stepSpring(start, 0, 10)).toEqual(stepSpring(start, 0, 0.25));
  });
});

describe('blinkClosure', () => {
  it('is open outside a blink and closes fully once inside it', () => {
    expect(blinkClosure(0)).toBe(0);
    expect(blinkClosure(BLINK_DURATION)).toBe(0);
    expect(blinkClosure(BLINK_DURATION * 0.4)).toBeCloseTo(1, 9);
    expect(blinkClosure(BLINK_DURATION * 0.2)).toBeCloseTo(0.5, 9);
  });

  it('waits between 2.2 and 5.2 seconds between blinks', () => {
    expect(nextBlinkDelay(0)).toBe(2.2);
    expect(nextBlinkDelay(0.999)).toBeLessThan(5.2);
  });
});

describe('headTurn', () => {
  it('looks straight ahead without a pointer', () => {
    expect(headTurn(null)).toEqual({ yaw: 0, pitch: 0 });
  });

  it('turns towards the pointer, clamped to a gentle range', () => {
    expect(headTurn({ x: 0.5, y: -0.5 })).toEqual({ yaw: MAX_YAW / 2, pitch: -MAX_PITCH / 2 });
    expect(headTurn({ x: 9, y: 9 })).toEqual({ yaw: MAX_YAW, pitch: MAX_PITCH });
  });
});
