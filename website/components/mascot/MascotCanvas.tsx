'use client';
/**
 * The bare R3F canvas for the mascot: camera, lights, contact shadow, and the
 * mascot itself (or, for the `head` view, just the gopher's round head). Runtime policy (when to pause, when to animate) lives in
 * MascotScene; the stills generator renders this component directly.
 */
import { Canvas, useFrame, useThree, type RootState } from '@react-three/fiber';
import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { CanvasTexture, PlaneGeometry, SRGBColorSpace, Vector3 } from 'three';
import { useToonKit, ToonKitProvider, useGeometry } from './kit';
import { Mascot, type MascotProps } from './Mascot';
import { GopherHead } from './parts/face';
import { PALETTE } from './toon';

/** A camera framing: where the camera sits and what it looks at. */
interface Framing {
  readonly position: readonly [number, number, number];
  readonly target: readonly [number, number, number];
}

/**
 * Camera framings, all with a 26° vertical field of view. `hero` is the
 * approved three-quarter view and `front` the straight-on view, both for a
 * 3:2 canvas; `head` is a slight three-quarter portrait of the head for a
 * square canvas.
 */
export const VIEWS = {
  hero: { position: [3.934, 4.073, 7.596], target: [-0.051, 0.765, -0.572] },
  front: { position: [0, 2.63, 7.353], target: [0, 0.919, -0.693] },
  head: { position: [1.86, 1.12, 5.75], target: [0, 0.22, 0] },
} as const satisfies Record<string, Framing>;

/** Name of a camera framing. */
export type MascotView = keyof typeof VIEWS;

const FOV = 26;
/** Key light (upper left front) and the rim light that drives the toon rim band. */
const KEY_LIGHT = new Vector3(-5, 5.5, 3.5);
const RIM_LIGHT = new Vector3(5, 4, -4);
const SHADOW = { radius: 2.52, x: 0, z: -0.73 };

/** Frames the subject and keeps the toon light directions in view space. */
function Rig({ view }: { view: MascotView }) {
  const camera = useThree((state) => state.camera);
  const kit = useToonKit();
  useLayoutEffect(() => {
    const { position, target } = VIEWS[view];
    camera.position.fromArray(position);
    camera.lookAt(new Vector3().fromArray(target));
    camera.updateMatrixWorld();
  }, [camera, view]);
  useFrame(() => {
    const view = camera.matrixWorldInverse;
    kit.lights.uKeyDir.value.copy(KEY_LIGHT).normalize().transformDirection(view);
    kit.lights.uRimDir.value.copy(RIM_LIGHT).normalize().transformDirection(view);
  });
  return (
    <>
      <hemisphereLight args={['#fff6ea', '#e9d6c4', 1.7]} />
      <directionalLight color="#fff4e6" intensity={2.4} position={KEY_LIGHT.toArray()} />
    </>
  );
}

/** A soft radial contact shadow under the box, painted into a small texture. */
function ContactShadow() {
  const geometry = useGeometry(() => new PlaneGeometry(SHADOW.radius * 2, SHADOW.radius * 2));
  const [texture] = useState(() => {
    const canvas = document.createElement('canvas');
    canvas.width = canvas.height = 128;
    const context = canvas.getContext('2d');
    if (context) {
      const gradient = context.createRadialGradient(64, 64, 0, 64, 64, 64);
      gradient.addColorStop(0, 'rgba(58,42,34,0.30)');
      gradient.addColorStop(0.55, 'rgba(58,42,34,0.14)');
      gradient.addColorStop(1, 'rgba(58,42,34,0)');
      context.fillStyle = gradient;
      context.fillRect(0, 0, 128, 128);
    }
    const map = new CanvasTexture(canvas);
    map.colorSpace = SRGBColorSpace;
    return map;
  });
  useEffect(() => () => texture.dispose(), [texture]);
  return (
    <mesh geometry={geometry} rotation-x={-Math.PI / 2} position={[SHADOW.x, 0.002, SHADOW.z]}>
      <meshBasicMaterial map={texture} transparent depthWrite={false} />
    </mesh>
  );
}

/** Calls `onReady` once, after the first frame has been drawn. */
function FirstFrame({ onReady }: { onReady?: () => void }) {
  const done = useRef(false);
  useFrame(() => {
    if (done.current) return;
    done.current = true;
    // useFrame runs before the frame is drawn; report on the next tick.
    requestAnimationFrame(() => onReady?.());
  });
  return null;
}

/** Props of {@link MascotCanvas}. */
export type MascotCanvasProps = MascotProps & {
  /** Camera framing; `head` shows only the round head. Defaults to `hero`. */
  readonly view?: MascotView;
  /** R3F frame loop: "always" animates, "demand" draws only on change, "never" pauses. */
  readonly frameloop: 'always' | 'demand' | 'never';
  /** Device pixel ratio range. */
  readonly dpr?: number | [number, number];
  /** Keep the drawing buffer readable (the stills generator reads pixels back). */
  readonly preserveDrawingBuffer?: boolean;
  readonly onReady?: () => void;
  /** The GPU dropped the WebGL context; the caller should fall back to the still. */
  readonly onContextLost?: () => void;
  readonly className?: string;
};

/** Transparent canvas with the mascot in its approved framing. */
export function MascotCanvas({
  view = 'hero',
  frameloop,
  dpr = [1, 2],
  preserveDrawingBuffer = false,
  onReady,
  onContextLost,
  className,
  ...mascot
}: MascotCanvasProps) {
  const handleCreated = ({ gl }: RootState) => {
    gl.setClearColor(0x000000, 0);
    if (onContextLost) {
      gl.domElement.addEventListener('webglcontextlost', onContextLost, { once: true });
    }
  };
  return (
    <Canvas
      className={className}
      aria-hidden
      frameloop={frameloop}
      dpr={dpr}
      flat
      gl={{ antialias: true, alpha: true, powerPreference: 'low-power', preserveDrawingBuffer }}
      camera={{ fov: FOV, near: 0.1, far: 100, position: [...VIEWS[view].position] }}
      onCreated={handleCreated}
    >
      <ToonKitProvider>
        <Rig view={view} />
        {view === 'head' ? (
          <GopherHead color={mascot.color ?? PALETTE.sora} />
        ) : (
          <>
            <ContactShadow />
            <Mascot {...mascot} />
          </>
        )}
        <FirstFrame onReady={onReady} />
      </ToonKitProvider>
    </Canvas>
  );
}
