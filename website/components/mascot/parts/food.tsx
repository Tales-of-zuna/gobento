'use client';
/**
 * Bento food, each piece standing on y = 0: a rice bed with an umeboshi,
 * tamagoyaki slices, an edamame pile, and an onigiri.
 */
import type { ThreeElements } from '@react-three/fiber';
import { useMemo } from 'react';
import { CapsuleGeometry, SphereGeometry, TorusGeometry } from 'three';
import { RoundedBoxGeometry } from 'three/addons/geometries/RoundedBoxGeometry.js';
import { createOnigiriGeometry } from '../geometry';
import { ToonMesh, useGeometry, useToonKit } from '../kit';
import { mulberry32, type Vec3 } from '../math';
import { PALETTE } from '../toon';

type GroupProps = Omit<ThreeElements['group'], 'children'>;

const RICE_SHADOW = '#c9b8a8';

/** Size of a rice bed. */
export interface RiceBedSize {
  readonly width: number;
  readonly depth: number;
  readonly height?: number;
}

/** Hinomaru rice: a white rice bed with a few grain bumps and one umeboshi. */
export function RiceBed({ width, depth, height = 0.26, ...props }: GroupProps & RiceBedSize) {
  const kit = useToonKit();
  const bed = useGeometry(
    () => new RoundedBoxGeometry(width, height, depth, 4, Math.min(0.1, height / 2 - 0.001)),
  );
  const grain = useGeometry(() => new SphereGeometry(0.05, 10, 8));
  const plum = useGeometry(() => new SphereGeometry(0.16, 28, 20));
  const grains = useMemo(() => {
    const random = mulberry32(7);
    return Array.from({ length: Math.round(width * depth * 22) }, () => ({
      position: [
        (random() - 0.5) * (width - 0.14),
        height - 0.005,
        (random() - 0.5) * (depth - 0.14),
      ] as Vec3,
      turn: random() * Math.PI,
    }));
  }, [width, depth, height]);
  const grainMaterial = kit.toon(PALETTE.rice, { rim: 0, spec: 0, shadowTint: RICE_SHADOW });
  return (
    <group {...props}>
      <ToonMesh
        geometry={bed}
        material={kit.toon(PALETTE.rice, { rim: 0.1, spec: 0, shadowTint: RICE_SHADOW })}
        outline={0.02}
        position={[0, height / 2, 0]}
      />
      {grains.map(({ position, turn }, i) => (
        <mesh
          key={i}
          geometry={grain}
          material={grainMaterial}
          position={position}
          rotation={[0, turn, 0]}
          scale={[1.4, 0.55, 0.8]}
        />
      ))}
      <ToonMesh
        geometry={plum}
        material={kit.toon(PALETTE.ume, { rim: 0.2, spec: 0.9, shadowTint: '#5a1a30' })}
        outline={0.02}
        position={[0, height + 0.06, 0]}
        scale={[1, 0.78, 1]}
      />
    </group>
  );
}

/** Two slices of rolled omelette, each with a darker swirl on its cut face. */
export function Tamagoyaki({ size = 0.34, ...props }: GroupProps & { size?: number }) {
  const kit = useToonKit();
  const slice = useGeometry(() => new RoundedBoxGeometry(size * 0.55, size * 0.72, size, 3, 0.06));
  const swirl = useGeometry(() => new TorusGeometry(size * 0.18, 0.012, 6, 24));
  const egg = kit.toon(PALETTE.tamago, { rim: 0.2, spec: 0.5, shadowTint: '#a0603a' });
  const swirlMaterial = kit.toon(PALETTE.tamagoDeep, { rim: 0, spec: 0 });
  return (
    <group {...props}>
      {[0, 1].map((i) => (
        <group
          key={i}
          position={[i * size * 0.6, size * 0.36, 0]}
          rotation={[0, 0, -0.12 + i * 0.05]}
        >
          <ToonMesh geometry={slice} material={egg} outline={0.02} />
          <mesh
            geometry={swirl}
            material={swirlMaterial}
            position={[size * 0.278, 0, 0]}
            rotation={[0, Math.PI / 2, 0]}
            scale={[1, 1.2, 1]}
          />
        </group>
      ))}
    </group>
  );
}

const POD_LENGTH = 0.42;

/** One edamame pod: a capsule with three bean bumps. */
function EdamamePod(props: GroupProps) {
  const kit = useToonKit();
  const pod = useGeometry(() => new CapsuleGeometry(0.075, POD_LENGTH, 6, 14));
  const bean = useGeometry(() => new SphereGeometry(0.095, 16, 12));
  const green = kit.toon(PALETTE.edamame, { rim: 0.25, spec: 0.5, shadowTint: '#2a4a30' });
  return (
    <group {...props}>
      <ToonMesh
        geometry={pod}
        material={green}
        outline={0.016}
        rotation={[0, 0, Math.PI / 2]}
        scale={[1, 1, 0.75]}
      />
      {[-1, 0, 1].map((i) => (
        <ToonMesh
          key={i}
          geometry={bean}
          material={green}
          outline={0.016}
          position={[i * POD_LENGTH * 0.36, 0.015, 0]}
          scale={[1, 0.9, 0.82]}
        />
      ))}
    </group>
  );
}

/** A loose pile of edamame pods, scattered the same way on every render. */
export function EdamamePile({
  count = 4,
  spread = 0.3,
  seed = 11,
  ...props
}: GroupProps & { count?: number; spread?: number; seed?: number }) {
  const pods = useMemo(() => {
    const random = mulberry32(seed);
    return Array.from({ length: count }, (_, i) => ({
      position: [(random() - 0.5) * spread, 0.08 + i * 0.03, (random() - 0.5) * spread] as Vec3,
      rotation: [(random() - 0.5) * 0.4, random() * Math.PI, (random() - 0.5) * 0.3] as Vec3,
    }));
  }, [count, spread, seed]);
  return (
    <group {...props}>
      {pods.map(({ position, rotation }, i) => (
        <EdamamePod key={i} position={position} rotation={rotation} />
      ))}
    </group>
  );
}

/** A triangle onigiri with a nori band, standing on its bottom edge. */
export function Onigiri({
  size = 0.31,
  depth = 0.18,
  ...props
}: GroupProps & { size?: number; depth?: number }) {
  const kit = useToonKit();
  const rice = useGeometry(() => createOnigiriGeometry(size, depth));
  const nori = useGeometry(
    () => new RoundedBoxGeometry(size * 0.75, size * 0.62, depth + 0.2, 2, 0.04),
  );
  return (
    <group {...props}>
      <ToonMesh
        geometry={rice}
        material={kit.toon(PALETTE.rice, { rim: 0.1, spec: 0.1, shadowTint: RICE_SHADOW })}
        outline={0.022}
      />
      <ToonMesh
        geometry={nori}
        material={kit.toon(PALETTE.nori, { rim: 0.2, spec: 0.4 })}
        outline={0.018}
        position={[0, size * 0.3, 0]}
      />
    </group>
  );
}
