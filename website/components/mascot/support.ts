/**
 * Decides, before any 3D code is downloaded, whether the mascot scene may
 * run. Kept free of React so the logic is unit-tested directly.
 */

/** Media query for the visitor's reduced-motion preference. */
export const REDUCED_MOTION_QUERY = '(prefers-reduced-motion: reduce)';

/** The subset of `window` the reduced-motion check needs. */
export type MatchMedia = (query: string) => Pick<MediaQueryList, 'matches'>;

/** Whether the visitor asked for reduced motion. Unknown (no matchMedia) counts as no. */
export function prefersReducedMotion(matchMedia: MatchMedia | undefined): boolean {
  try {
    return matchMedia?.(REDUCED_MOTION_QUERY).matches ?? false;
  } catch {
    return false;
  }
}

/** Something that can hand out a canvas to probe for WebGL. */
export type CanvasFactory = () => Pick<HTMLCanvasElement, 'getContext'>;

/**
 * What the WebGL probe found: no context at all, a context drawn by a
 * software rasterizer, or one backed by a GPU (or whose renderer is hidden).
 */
export type WebGLSupport = 'none' | 'software' | 'hardware';

/**
 * Renderer names of software rasterizers: Chromium's SwiftShader, Mesa's
 * llvmpipe and softpipe, Windows' Microsoft Basic Render Driver, and any
 * renderer calling itself software (such as Apple Software Renderer).
 */
const SOFTWARE_RENDERER = /swiftshader|llvmpipe|softpipe|basic render driver|\bsoftware\b/i;

/** Whether a WebGL renderer name is a software rasterizer. */
export function isSoftwareRenderer(renderer: string): boolean {
  return SOFTWARE_RENDERER.test(renderer);
}

/**
 * Probes WebGL (WebGL 2, else WebGL 1) and classifies its renderer from the
 * unmasked renderer name; a browser that hides the name counts as hardware.
 * The probe context is released straight away so it does not count against
 * the browser's context limit. Any exception means none.
 */
export function probeWebGL(createCanvas: CanvasFactory): WebGLSupport {
  try {
    const canvas = createCanvas();
    const context = canvas.getContext('webgl2') ?? canvas.getContext('webgl');
    if (!context) return 'none';
    const debug = context.getExtension('WEBGL_debug_renderer_info');
    const renderer: unknown = debug ? context.getParameter(debug.UNMASKED_RENDERER_WEBGL) : '';
    context.getExtension('WEBGL_lose_context')?.loseContext();
    return typeof renderer === 'string' && isSoftwareRenderer(renderer) ? 'software' : 'hardware';
  } catch {
    return 'none';
  }
}

/** What the mascot wrapper knows when deciding whether to mount the scene. */
export interface SceneConditions {
  /** WebGL probe result, or null while not yet known (server render, first paint). */
  readonly webgl: boolean | null;
  /** Reduced-motion preference, or null while not yet known. */
  readonly reducedMotion: boolean | null;
  /** The scene threw, failed to load, or lost its WebGL context. */
  readonly failed: boolean;
  /** The scene has already drawn its first frame. */
  readonly ready: boolean;
  /**
   * The visitor has interacted with the page (moved the pointer, typed,
   * touched, scrolled), or null while not yet known. The scene waits for it,
   * so the 3D code never loads while the page itself is loading.
   */
  readonly interacted: boolean | null;
}

/**
 * Whether to mount the 3D scene. The static still is the answer until the
 * visitor first interacts with the page, and whenever WebGL is missing, the
 * scene failed, or reduced motion was requested before the scene started (so
 * reduced motion never downloads the 3D code). A scene that is already
 * running stays, and switches to its static pose instead.
 */
export function shouldMountScene({
  webgl,
  reducedMotion,
  failed,
  ready,
  interacted,
}: SceneConditions): boolean {
  if (webgl !== true || failed || reducedMotion === null) return false;
  if (ready) return true;
  return interacted === true && !reducedMotion;
}
