/**
 * The mascot's toon look: palette, a soft three-step light ramp, toon
 * materials with a fresnel rim and a crisp vinyl highlight, and inverted-hull
 * outline materials.
 *
 * Materials are shared through a {@link ToonKit}: one kit per scene, created
 * once and disposed with the scene, so every part that asks for the same
 * colour and options gets the same material (and the same shader program).
 */
import {
  BackSide,
  BufferGeometry,
  Color,
  DataTexture,
  LinearFilter,
  MeshBasicMaterial,
  MeshToonMaterial,
  RGBAFormat,
  ShaderMaterial,
  Vector3,
} from 'three';
import { mergeVertices } from 'three/addons/utils/BufferGeometryUtils.js';

/** Direction A palette plus the body colour of the approved concept ("sora" blue). */
export const PALETTE = {
  sora: '#A9C6E6',
  washi: '#FBF7F0',
  sumi: '#3A2A22',
  shu: '#C04A2E',
  matcha: '#4E8F3A',
  rice: '#FFFDF7',
  tamago: '#F4C74F',
  tamagoDeep: '#E2A93A',
  ume: '#D2453F',
  edamame: '#7DB24A',
  nori: '#2D3A2C',
  blush: '#F29A82',
  tooth: '#FFFFFF',
  cream: '#FFF1DC',
} as const;

/** Outline thickness (world units) for meshes that do not set their own. */
export const DEFAULT_OUTLINE = 0.038;

/** Look options for {@link ToonKit.toon}. */
export interface ToonOptions {
  /** Strength of the fresnel rim band, 0 to 1. */
  readonly rim?: number;
  /** Strength of the crisp vinyl highlight, 0 to 1. */
  readonly spec?: number;
  /** Colour the shadow band is tinted towards instead of plain grey. */
  readonly shadowTint?: string;
}

/**
 * Light directions in view space, shared by every toon material of a kit.
 * The scene updates them each frame from the key light and the rim position.
 */
export interface ToonLightUniforms {
  readonly uKeyDir: { value: Vector3 };
  readonly uRimDir: { value: Vector3 };
}

/** A per-scene cache of shared mascot materials. */
export interface ToonKit {
  readonly lights: ToonLightUniforms;
  /** Toon material with the three-step ramp, rim and highlight. */
  toon(color: string, options?: ToonOptions): MeshToonMaterial;
  /** Unlit material (highlights in the eyes, blush). */
  flat(color: string): MeshBasicMaterial;
  /** Inverted-hull outline material that pushes vertices out by `thickness`. */
  outline(thickness: number): ShaderMaterial;
  /** Frees every material and texture the kit created. */
  dispose(): void;
}

/** Smoothstep as in GLSL. */
function smoothstep(edge0: number, edge1: number, x: number): number {
  const t = Math.min(Math.max((x - edge0) / (edge1 - edge0), 0), 1);
  return t * t * (3 - 2 * t);
}

/**
 * Brightness of the light ramp at `u = dot(N, L) * 0.5 + 0.5`: shadow, mid
 * and lit steps with short smoothed transitions, the terminator at 0.5.
 */
export function rampValue(u: number): number {
  return 0.6 + 0.24 * smoothstep(0.47, 0.52, u) + 0.16 * smoothstep(0.72, 0.77, u);
}

function createGradientMap(): DataTexture {
  const size = 64;
  const data = new Uint8Array(size * 4);
  for (let i = 0; i < size; i++) {
    const c = Math.round(rampValue(i / (size - 1)) * 255);
    data.set([c, c, c, 255], i * 4);
  }
  const texture = new DataTexture(data, size, 1, RGBAFormat);
  texture.minFilter = LinearFilter;
  texture.magFilter = LinearFilter;
  texture.generateMipmaps = false;
  texture.needsUpdate = true;
  return texture;
}

/** GLSL added before `#include <opaque_fragment>` of MeshToonMaterial. */
const TOON_FRAGMENT = /* glsl */ `
  {
    vec3 N = normalize(normal);
    vec3 V = normalize(vViewPosition);
    float ndl = dot(N, uKeyDir);
    float shade = 1.0 - smoothstep(-0.06, 0.04, ndl);
    outgoingLight = mix(outgoingLight, outgoingLight * uShadowTint * 1.6, shade * 0.35);
    float fres = 1.0 - max(dot(N, V), 0.0);
    float rimBand = smoothstep(0.56, 0.62, fres) * smoothstep(-0.1, 0.25, dot(N, uRimDir));
    outgoingLight += vec3(0.95, 0.98, 1.0) * rimBand * uRim * 0.55;
    vec3 H = normalize(uKeyDir + V);
    float sp = smoothstep(0.965, 0.975, dot(N, H));
    outgoingLight = mix(outgoingLight, vec3(1.0), sp * uSpec * 0.85);
  }
  #include <opaque_fragment>`;

const OUTLINE_VERTEX = /* glsl */ `
  uniform float thickness;
  void main() {
    vec3 n = normalize(normalMatrix * normal);
    vec4 mv = modelViewMatrix * vec4(position, 1.0);
    mv.xyz += n * thickness;
    gl_Position = projectionMatrix * mv;
  }`;

const OUTLINE_FRAGMENT = /* glsl */ `
  uniform vec3 color;
  void main() {
    gl_FragColor = vec4(color, 1.0);
    #include <colorspace_fragment>
  }`;

/** Creates an empty kit; materials are built on first request and then reused. */
export function createToonKit(): ToonKit {
  const lights: ToonLightUniforms = {
    uKeyDir: { value: new Vector3(0.4, 0.7, 0.6).normalize() },
    uRimDir: { value: new Vector3(0.6, 0.5, -0.6).normalize() },
  };
  const gradientMap = createGradientMap();
  const toons = new Map<string, MeshToonMaterial>();
  const flats = new Map<string, MeshBasicMaterial>();
  const outlines = new Map<number, ShaderMaterial>();

  function toon(color: string, options: ToonOptions = {}): MeshToonMaterial {
    const { rim = 0.35, spec = 0.5, shadowTint = '#6b4a5a' } = options;
    const key = `${color}|${rim}|${spec}|${shadowTint}`;
    const cached = toons.get(key);
    if (cached) return cached;
    const material = new MeshToonMaterial({ color, gradientMap });
    material.onBeforeCompile = (shader) => {
      Object.assign(shader.uniforms, lights, {
        uRim: { value: rim },
        uSpec: { value: spec },
        uShadowTint: { value: new Color(shadowTint) },
      });
      shader.fragmentShader = shader.fragmentShader
        .replace(
          'void main() {',
          'uniform vec3 uKeyDir; uniform vec3 uRimDir; uniform float uRim; uniform float uSpec; uniform vec3 uShadowTint;\nvoid main() {',
        )
        .replace('#include <opaque_fragment>', TOON_FRAGMENT);
    };
    material.customProgramCacheKey = () => `gobento-toon-${key}`;
    toons.set(key, material);
    return material;
  }

  function flat(color: string): MeshBasicMaterial {
    let material = flats.get(color);
    if (!material) {
      material = new MeshBasicMaterial({ color });
      flats.set(color, material);
    }
    return material;
  }

  function outline(thickness: number): ShaderMaterial {
    let material = outlines.get(thickness);
    if (!material) {
      material = new ShaderMaterial({
        uniforms: { thickness: { value: thickness }, color: { value: new Color(PALETTE.sumi) } },
        vertexShader: OUTLINE_VERTEX,
        fragmentShader: OUTLINE_FRAGMENT,
        side: BackSide,
      });
      outlines.set(thickness, material);
    }
    return material;
  }

  function dispose(): void {
    for (const cache of [toons, flats, outlines]) {
      cache.forEach((material) => material.dispose());
      cache.clear();
    }
    gradientMap.dispose();
  }

  return { lights, toon, flat, outline, dispose };
}

/**
 * Geometry for an inverted-hull outline: the same surface with seams welded
 * and normals averaged, so pushing vertices along the normal never cracks the
 * hull open at hard edges.
 */
export function createHullGeometry(source: BufferGeometry): BufferGeometry {
  const bare = new BufferGeometry();
  bare.setAttribute('position', source.getAttribute('position').clone());
  if (source.index) bare.setIndex(source.index.clone());
  const hull = mergeVertices(bare, 1e-4);
  bare.dispose();
  hull.computeVertexNormals();
  return hull;
}
