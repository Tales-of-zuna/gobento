'use client';
/**
 * The bare R3F canvas for the mascot: camera, lights, contact shadow, and the
 * mascot itself. Runtime policy (when to pause, when to animate) lives in
 * MascotScene; the stills generator renders this component directly.
 */
import { Canvas, useFrame, useThree, type RootState } from '@react-three/fiber';
import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { CanvasTexture, PlaneGeometry, SRGBColorSpace, Vector3 } from 'three';
import { useToonKit, ToonKitProvider, useGeometry } from './kit';
import { Mascot, type MascotProps } from './Mascot';

/** Camera framing of the approved three-quarter view. */
const CAMERA = {
  fov: 26,
  position: [4.675, 4.626, 9.525] as const,
  target: new Vector3(0.064, 2.004, 0.072),
};
/** Key light (upper left front) and the rim light that drives the toon rim band. */
const KEY_LIGHT = new Vector3(-5, 5.5, 3.5);
const RIM_LIGHT = new Vector3(5, 4, -4);
const SHADOW = { radius: 1.84, x: 0.064, z: 0.072 };

/** Points the camera at the mascot and keeps the toon light directions in view space. */
function Rig() {
  const camera = useThree((state) => state.camera);
  const kit = useToonKit();
  useLayoutEffect(() => {
    camera.lookAt(CAMERA.target);
    camera.updateMatrixWorld();
  }, [camera]);
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

/** A soft radial contact shadow under the feet, painted into a small texture. */
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
      camera={{ fov: CAMERA.fov, near: 0.1, far: 100, position: [...CAMERA.position] }}
      onCreated={handleCreated}
    >
      <ToonKitProvider>
        <Rig />
        <ContactShadow />
        <Mascot {...mascot} />
        <FirstFrame onReady={onReady} />
      </ToonKitProvider>
    </Canvas>
  );
}
