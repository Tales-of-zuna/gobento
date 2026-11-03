'use client';
/**
 * The mascot's body: a shu-lacquered bento box turned on its side, so its
 * opening faces the viewer like a belly, with four compartments of food.
 */
import { useFrame, type ThreeElements } from '@react-three/fiber';
import { useRef, type ReactNode } from 'react';
import type { Group } from 'three';
import { RoundedBoxGeometry } from 'three/addons/geometries/RoundedBoxGeometry.js';
import { createRingGeometry } from '../geometry';
import { ToonMesh, useGeometry, useToonKit } from '../kit';
import { clamp, springAtRest, stepSpring, type SpringState } from '../math';
import { PALETTE } from '../toon';
import { EdamamePile, Onigiri, RiceBed, Tamagoyaki } from './food';

type GroupProps = Omit<ThreeElements['group'], 'children'>;

/** Which of the four compartments hold food, in reading order (top-left first). */
export type Packed = readonly [boolean, boolean, boolean, boolean];

/** Outer size of the box. Box-local axes: +Y is the opening, -Z is up on screen. */
export const BOX = {
  width: 2.1,
  /** Height on screen (box-local depth). */
  height: 1.75,
  /** Front-to-back depth (box-local height). */
  depth: 1.3,
  wall: 0.13,
  radius: 0.32,
  /** The floor is thick, so the compartments are shallow shelves. */
  floor: 0.8,
} as const;

const INNER = {
  width: BOX.width - BOX.wall * 2,
  height: BOX.height - BOX.wall * 2,
  floorY: BOX.floor + 0.02,
} as const;

const DIVIDER_HEIGHT = BOX.depth - INNER.floorY - 0.12;

/** The lacquered shell: shu walls, a sumi lining and floor, and cross dividers. */
function BoxShell() {
  const kit = useToonKit();
  const walls = useGeometry(() =>
    createRingGeometry({
      width: BOX.width,
      depth: BOX.height,
      radius: BOX.radius,
      wall: BOX.wall,
      height: BOX.depth,
    }),
  );
  const lining = useGeometry(() =>
    createRingGeometry({
      width: INNER.width + 0.02,
      depth: INNER.height + 0.02,
      radius: BOX.radius - BOX.wall,
      wall: 0.05,
      height: BOX.depth - 0.05,
      bevel: 0.012,
    }),
  );
  const floor = useGeometry(
    () => new RoundedBoxGeometry(INNER.width + 0.04, BOX.floor, INNER.height + 0.04, 2, 0.04),
  );
  const verticalDivider = useGeometry(
    () => new RoundedBoxGeometry(0.06, DIVIDER_HEIGHT, INNER.height, 2, 0.025),
  );
  const horizontalDivider = useGeometry(
    () => new RoundedBoxGeometry(INNER.width, DIVIDER_HEIGHT, 0.06, 2, 0.025),
  );
  const shu = kit.toon(PALETTE.shu, { rim: 0.2, spec: 0.35, shadowTint: '#5a2a3a' });
  const lacquer = kit.toon(PALETTE.sumi, { rim: 0.25, spec: 0.7, shadowTint: '#3a2a40' });
  const divider = kit.toon(PALETTE.sumi, { rim: 0.25, spec: 0.6 });
  const dividerY = INNER.floorY + DIVIDER_HEIGHT / 2;
  return (
    <group>
      <ToonMesh geometry={walls} material={shu} />
      <ToonMesh geometry={lining} material={lacquer} outline={false} />
      <ToonMesh
        geometry={floor}
        material={lacquer}
        outline={false}
        position={[0, BOX.floor / 2 + 0.02, 0]}
      />
      <ToonMesh
        geometry={verticalDivider}
        material={divider}
        outline={0.018}
        position={[0, dividerY, 0]}
      />
      <ToonMesh
        geometry={horizontalDivider}
        material={divider}
        outline={0.018}
        position={[0, dividerY, 0]}
      />
    </group>
  );
}

const AT_REST: SpringState = { position: 0, velocity: 0 };

/**
 * One compartment's contents. When `packed` turns true the food drops in from
 * in front of the box with a spring; without animation it simply appears.
 */
function Compartment({
  packed,
  animate,
  children,
  ...props
}: GroupProps & { packed: boolean; animate: boolean; children: ReactNode }) {
  const piece = useRef<Group>(null);
  const spring = useRef<SpringState>(AT_REST);
  const wasPacked = useRef(packed);

  useFrame((_, delta) => {
    const group = piece.current;
    if (!group) return;
    if (packed && !wasPacked.current) spring.current = { position: 1, velocity: 0 };
    wasPacked.current = packed;
    if (!animate) spring.current = AT_REST;
    else if (!springAtRest(spring.current, 0))
      spring.current = stepSpring(spring.current, 0, delta);
    const p = spring.current.position;
    // Falls in along the opening (+Y) and slightly from above (-Z).
    group.position.set(0, p * 0.9, -p * 0.5);
    group.scale.setScalar(1 - clamp(p, 0, 1) * 0.4);
  });

  return (
    <group {...props}>
      <group ref={piece} visible={packed}>
        {children}
      </group>
    </group>
  );
}

const DEG = Math.PI / 180;

/** Props of {@link BentoBody}. */
export type BentoBodyProps = GroupProps & {
  /** Which compartments show food. */
  readonly packed: Packed;
  /** Spring the food in when a compartment fills; false snaps instantly. */
  readonly animate: boolean;
};

/**
 * The bento box body, centred on its middle, opening towards +Z, with rice
 * and umeboshi, tamagoyaki, edamame, and an onigiri.
 */
export function BentoBody({ packed, animate, ...props }: BentoBodyProps) {
  const cellX = INNER.width / 4;
  const cellZ = INNER.height / 4;
  return (
    <group {...props}>
      <group position={[0, 0, -BOX.depth / 2]} rotation={[90 * DEG, 0, 0]}>
        <BoxShell />
        <Compartment packed={packed[0]} animate={animate} position={[-cellX, INNER.floorY, -cellZ]}>
          <RiceBed width={INNER.width / 2 - 0.12} depth={INNER.height / 2 - 0.12} />
        </Compartment>
        <Compartment
          packed={packed[1]}
          animate={animate}
          position={[cellX - 0.18, INNER.floorY + 0.1, -cellZ]}
        >
          <Tamagoyaki rotation={[90 * DEG, 0, 0]} />
        </Compartment>
        <Compartment packed={packed[2]} animate={animate} position={[-cellX, INNER.floorY, cellZ]}>
          <EdamamePile />
        </Compartment>
        <Compartment
          packed={packed[3]}
          animate={animate}
          position={[cellX, INNER.floorY + 0.12, cellZ + 0.24]}
        >
          <Onigiri rotation={[-90 * DEG, 0, 0]} />
        </Compartment>
      </group>
    </group>
  );
}
