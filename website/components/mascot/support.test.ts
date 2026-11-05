import { describe, expect, it, vi } from 'vitest';
import {
  prefersReducedMotion,
  REDUCED_MOTION_QUERY,
  isSoftwareRenderer,
  probeWebGL,
  shouldMountScene,
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

const UNMASKED_RENDERER = 0x9246;

/**
 * A canvas whose getContext answers from `contexts` by context id, with a
 * context reporting `renderer` as its unmasked renderer name (or hiding it
 * when undefined).
 */
function fakeCanvas(
  contexts: Record<string, unknown>,
  renderer?: string,
): {
  factory: CanvasFactory;
  loseContext: ReturnType<typeof vi.fn>;
} {
  const loseContext = vi.fn();
  const context = {
    getExtension: (name: string) => {
      if (name === 'WEBGL_lose_context') return { loseContext };
      if (name === 'WEBGL_debug_renderer_info' && renderer !== undefined) {
        return { UNMASKED_RENDERER_WEBGL: UNMASKED_RENDERER };
      }
      return null;
    },
    getParameter: (parameter: number) => (parameter === UNMASKED_RENDERER ? renderer : null),
  };
  const factory = (() => ({
    getContext: (id: string) => (contexts[id] ? context : null),
  })) as unknown as CanvasFactory;
  return { factory, loseContext };
}

const GPU = 'ANGLE (NVIDIA, NVIDIA GeForce RTX 4070 Direct3D11 vs_5_0 ps_5_0, D3D11)';
const SWIFTSHADER =
  'ANGLE (Google, Vulkan 1.3.0 (SwiftShader Device (Subzero) (0x0000C0DE)), SwiftShader driver)';

describe('probeWebGL', () => {
  it('finds hardware WebGL 2 and releases the probe context', () => {
    const { factory, loseContext } = fakeCanvas({ webgl2: true }, GPU);
    expect(probeWebGL(factory)).toBe('hardware');
    expect(loseContext).toHaveBeenCalledOnce();
  });

  it('falls back to WebGL 1', () => {
    expect(probeWebGL(fakeCanvas({ webgl: true }, GPU).factory)).toBe('hardware');
  });

  it('finds a software rasterizer and releases the probe context', () => {
    const { factory, loseContext } = fakeCanvas({ webgl2: true }, SWIFTSHADER);
    expect(probeWebGL(factory)).toBe('software');
    expect(loseContext).toHaveBeenCalledOnce();
  });

  it('counts a hidden renderer name as hardware', () => {
    expect(probeWebGL(fakeCanvas({ webgl2: true }).factory)).toBe('hardware');
  });

  it('finds none without any WebGL context', () => {
    expect(probeWebGL(fakeCanvas({}).factory)).toBe('none');
  });

  it('finds none when creating a context throws', () => {
    const throwing = (() => ({
      getContext: () => {
        throw new Error('blocked');
      },
    })) as unknown as CanvasFactory;
    expect(probeWebGL(throwing)).toBe('none');
  });
});

describe('isSoftwareRenderer', () => {
  it.each([
    SWIFTSHADER,
    'Google SwiftShader',
    'llvmpipe (LLVM 17.0.6, 256 bits)',
    'Mesa softpipe',
    'ANGLE (Microsoft, Microsoft Basic Render Driver Direct3D11 vs_5_0 ps_5_0, D3D11)',
    'Apple Software Renderer',
  ])('%s is software', (renderer) => {
    expect(isSoftwareRenderer(renderer)).toBe(true);
  });

  it.each([
    GPU,
    'ANGLE (Intel, Mesa Intel(R) UHD Graphics 620 (KBL GT2), OpenGL 4.6)',
    'Apple M2',
    'Mali-G78',
    'Adreno (TM) 740',
    '',
  ])('%s is hardware', (renderer) => {
    expect(isSoftwareRenderer(renderer)).toBe(false);
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
