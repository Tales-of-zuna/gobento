import { describe, expect, it, vi } from 'vitest';
import {
  prefersReducedMotion,
  REDUCED_MOTION_QUERY,
  shouldMountScene,
  supportsWebGL,
  type CanvasFactory,
  type SceneConditions,
} from './support';

describe('prefersReducedMotion', () => {
  it('reads the reduced-motion media query', () => {
    const matchMedia = vi.fn((query: string) => ({ matches: query === REDUCED_MOTION_QUERY }));
    expect(prefersReducedMotion(matchMedia)).toBe(true);
    expect(matchMedia).toHaveBeenCalledWith('(prefers-reduced-motion: reduce)');
  });

  it('is false when the query does not match', () => {
    expect(prefersReducedMotion(() => ({ matches: false }))).toBe(false);
  });

  it('is false when matchMedia is missing or throws', () => {
    expect(prefersReducedMotion(undefined)).toBe(false);
    expect(
      prefersReducedMotion(() => {
        throw new Error('unsupported');
      }),
    ).toBe(false);
  });
});

/** A canvas whose getContext answers from `contexts` by context id. */
function fakeCanvas(contexts: Record<string, unknown>): {
  factory: CanvasFactory;
  loseContext: ReturnType<typeof vi.fn>;
} {
  const loseContext = vi.fn();
  const context = { getExtension: () => ({ loseContext }) };
  const factory = (() => ({
    getContext: (id: string) => (contexts[id] ? context : null),
  })) as unknown as CanvasFactory;
  return { factory, loseContext };
}

describe('supportsWebGL', () => {
  it('accepts WebGL 2 and releases the probe context', () => {
    const { factory, loseContext } = fakeCanvas({ webgl2: true });
    expect(supportsWebGL(factory)).toBe(true);
    expect(loseContext).toHaveBeenCalledOnce();
  });

  it('falls back to WebGL 1', () => {
    expect(supportsWebGL(fakeCanvas({ webgl: true }).factory)).toBe(true);
  });

  it('rejects a browser without any WebGL context', () => {
    expect(supportsWebGL(fakeCanvas({}).factory)).toBe(false);
  });

  it('rejects a browser where creating a context throws', () => {
    const throwing = (() => ({
      getContext: () => {
        throw new Error('blocked');
      },
    })) as unknown as CanvasFactory;
    expect(supportsWebGL(throwing)).toBe(false);
  });
});

describe('shouldMountScene', () => {
  const base: SceneConditions = {
    webgl: true,
    reducedMotion: false,
    failed: false,
    ready: false,
    interacted: true,
  };
  const cases: { name: string; when: Partial<SceneConditions>; mount: boolean }[] = [
    { name: 'WebGL works and motion is fine', when: {}, mount: true },
    { name: 'WebGL is unavailable', when: { webgl: false }, mount: false },
    { name: 'WebGL is not probed yet (server render)', when: { webgl: null }, mount: false },
    { name: 'the preference is not known yet', when: { reducedMotion: null }, mount: false },
    { name: 'reduced motion is requested up front', when: { reducedMotion: true }, mount: false },
    {
      name: 'reduced motion is turned on after the scene started',
      when: { reducedMotion: true, ready: true },
      mount: true,
    },
    { name: 'the visitor has not interacted yet', when: { interacted: false }, mount: false },
    {
      name: 'interaction is not known yet (server render)',
      when: { interacted: null },
      mount: false,
    },
    {
      name: 'reduced motion is requested and the visitor interacts',
      when: { reducedMotion: true, interacted: true },
      mount: false,
    },
    { name: 'the scene failed', when: { failed: true }, mount: false },
    { name: 'a running scene failed', when: { failed: true, ready: true }, mount: false },
  ];
  it.each(cases)('$name → mount: $mount', ({ when, mount }) => {
    expect(shouldMountScene({ ...base, ...when })).toBe(mount);
  });
});
