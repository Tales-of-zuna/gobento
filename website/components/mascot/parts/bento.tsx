'use client';
/**
 * The open bento box the gopher sits in: shu lacquer outside, sumi inside, a
 * large compartment on the left for the gopher, and a 2x2 block of food on
 * the right. Also the lid, with its washi band and hanko seal.
 */
import { useFrame, type ThreeElements } from '@react-three/fiber';
import { useRef, type ReactNode, type Ref } from 'react';
import type { Group } from 'three';
import { RoundedBoxGeometry } from 'three/addons/geometries/RoundedBoxGeometry.js';
import { createBaranGeometry, createRingGeometry } from '../geometry';
import { ToonMesh, useGeometry, useToonKit } from '../kit';
import { clamp, springAtRest, stepSpring, type SpringState } from '../math';
import { PALETTE } from '../toon';
import { EdamamePile, Onigiri, RiceBed, Tamagoyaki } from './food';

type GroupProps = Omit<ThreeElements['group'], 'children'>;

/** Which food compartments are filled: back-left, back-right, front-left, front-right. */
export type Packed = readonly [boolean, boolean, boolean, boolean];

/** Outer size of the box. Its base is at y = 0 and the opening faces +Y. */
const BOX = {
  width: 4,
  depth: 2.5,
  height: 0.82,
  wall: 0.15,
  radius: 0.34,
  floor: 0.44,
} as const;

const INNER = {
  width: BOX.width - BOX.wall * 2,
  depth: BOX.depth - BOX.wall * 2,
  floorY: BOX.floor + 0.02,
} as const;

/** The food block starts this far right of centre; the gopher takes the rest. */
const SPLIT_X = 0.62;
const FOOD_WIDTH = INNER.width / 2 - SPLIT_X;
const DIVIDER_HEIGHT = 0.24;
const DEG = Math.PI / 180;

/** Shu walls, sumi lining and floor. */
function Shell() {
  const kit = useToonKit();
  const walls = useGeometry(() =>
    createRingGeometry({
      width: BOX.width,
      depth: BOX.depth,
      radius: BOX.radius,
      wall: BOX.wall,
      height: BOX.height,
    }),
  );
  const lining = useGeometry(() =>
    createRingGeometry({
      width: INNER.width + 0.02,
      depth: INNER.depth + 0.02,
      radius: BOX.radius - BOX.wall,
      wall: 0.05,
      height: BOX.height - 0.05,
      bevel: 0.012,
    }),
  );
  const floor = useGeometry(
    () => new RoundedBoxGeometry(INNER.width + 0.04, BOX.floor, INNER.depth + 0.04, 2, 0.04),
  );
  const lacquer = kit.toon(PALETTE.sumi, { rim: 0.25, spec: 0.7, shadowTint: '#3a2a40' });
  return (
    <group>
      <ToonMesh
        geometry={walls}
        material={kit.toon(PALETTE.shu, { rim: 0.2, spec: 0.35, shadowTint: '#5a2a3a' })}
      />
      <ToonMesh geometry={lining} material={lacquer} outline={false} />
      <ToonMesh
        geometry={floor}
        material={lacquer}
        outline={false}
        position={[0, BOX.floor / 2 + 0.02, 0]}
      />
    </group>
  );
}

/** Thin lacquer dividers between the gopher and the food, and within the food block. */
function Dividers() {
  const kit = useToonKit();
  const fullDepth = useGeometry(
    () => new RoundedBoxGeometry(0.06, DIVIDER_HEIGHT, INNER.depth, 2, 0.025),
  );
  const across = useGeometry(
    () => new RoundedBoxGeometry(FOOD_WIDTH, DIVIDER_HEIGHT, 0.06, 2, 0.025),
  );
  const material = kit.toon(PALETTE.sumi, { rim: 0.25, spec: 0.6 });
  const y = INNER.floorY + DIVIDER_HEIGHT / 2;
  return (
    <group>
      <ToonMesh
        geometry={fullDepth}
        material={material}
        outline={0.018}
        position={[SPLIT_X, y, 0]}
      />
      <ToonMesh
        geometry={fullDepth}
        material={material}
        outline={0.018}
        position={[SPLIT_X + FOOD_WIDTH / 2, y, 0]}
      />
      <ToonMesh
        geometry={across}
        material={material}
        outline={0.018}
        position={[SPLIT_X + FOOD_WIDTH / 2, y, 0]}
      />
    </group>
  );
}

/** The green baran sheet standing where the food dividers cross. */
function Baran(props: GroupProps) {
  const kit = useToonKit();
  const geometry = useGeometry(() => createBaranGeometry());
  return (
    <group {...props}>
      <ToonMesh
        geometry={geometry}
        material={kit.toon(PALETTE.matcha, { rim: 0.2, spec: 0.3, shadowTint: '#2b4a40' })}
        outline={0.016}
      />
    </group>
  );
}

const AT_REST: SpringState = { position: 0, velocity: 0 };
/** How high above its compartment a piece starts when it drops in. */
const DROP_HEIGHT = 1.4;

/**
 * One compartment's food. When `packed` turns true the food drops in from
 * above and settles with a spring; without animation it simply appears.
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
    group.position.setY(p * DROP_HEIGHT);
    group.scale.setScalar(1 - clamp(p, 0, 1) * 0.3);
  });

  return (
    <group {...props}>
      <group ref={piece} visible={packed}>
        {children}
      </group>
    </group>
  );
}

/** Props of {@link BentoBox}. */
export type BentoBoxProps = GroupProps & {
  /** Which food compartments are filled. */
  readonly packed: Packed;
  /** Spring the food in when a compartment fills; false snaps instantly. */
  readonly animate: boolean;
};

/** The open box with dividers, baran, and the four food compartments. */
export function BentoBox({ packed, animate, ...props }: BentoBoxProps) {
  const left = SPLIT_X + FOOD_WIDTH * 0.25;
  const right = SPLIT_X + FOOD_WIDTH * 0.75;
  const back = -INNER.depth / 4;
  const front = INNER.depth / 4;
  const y = INNER.floorY;
  return (
    <group {...props}>
      <Shell />
      <Dividers />
      <Compartment packed={packed[0]} animate={animate} position={[left, y, back]}>
        <RiceBed width={FOOD_WIDTH / 2 - 0.1} depth={INNER.depth / 2 - 0.1} height={0.24} />
      </Compartment>
      <Compartment packed={packed[1]} animate={animate} position={[right, y, back - 0.05]}>
        <Onigiri size={0.27} rotation={[-8 * DEG, -25 * DEG, 0]} />
      </Compartment>
      <Compartment packed={packed[2]} animate={animate} position={[left - 0.08, y + 0.1, front]}>
        <Tamagoyaki size={0.32} rotation={[0, -60 * DEG, 0]} />
      </Compartment>
      <Compartment packed={packed[3]} animate={animate} position={[right, y + 0.14, front]}>
        <EdamamePile count={3} spread={0.18} seed={4} />
      </Compartment>
      <Baran position={[SPLIT_X + FOOD_WIDTH / 2, y, 0.02]} rotation={[0, -10 * DEG, 0]} />
    </group>
  );
}

const LID_HEIGHT = 0.26;
const BAND_WIDTH = 0.5;

/**
 * The lid, modelled lying flat with its top at y = LID_HEIGHT: a shallow
 * shu box with a washi paper band, a matcha stripe, and a little hanko seal.
 */
function LidBody() {
  const kit = useToonKit();
  const top = useGeometry(
    () =>
      new RoundedBoxGeometry(
        BOX.width + 0.06,
        LID_HEIGHT,
        BOX.depth + 0.06,
        4,
        LID_HEIGHT / 2 - 0.001,
      ),
  );
  const paper = useGeometry(
    () => new RoundedBoxGeometry(BAND_WIDTH, LID_HEIGHT + 0.05, BOX.depth + 0.12, 2, 0.02),
  );
  const stripe = useGeometry(
    () => new RoundedBoxGeometry(BAND_WIDTH * 0.22, LID_HEIGHT + 0.06, BOX.depth + 0.13, 1, 0.01),
  );
  const seal = useGeometry(
    () => new RoundedBoxGeometry(BAND_WIDTH * 0.62, 0.03, BAND_WIDTH * 0.62, 2, 0.04),
  );
  const bandX = BOX.width * 0.28;
  return (
    <group>
      <ToonMesh
        geometry={top}
        material={kit.toon(PALETTE.shu, { rim: 0, spec: 0.12, shadowTint: '#5a2a3a' })}
        position={[0, LID_HEIGHT / 2, 0]}
      />
      <ToonMesh
        geometry={paper}
        material={kit.toon(PALETTE.band, { rim: 0.2, spec: 0.1 })}
        outline={0.02}
        position={[bandX, LID_HEIGHT / 2, 0]}
      />
      <ToonMesh
        geometry={stripe}
        material={kit.toon(PALETTE.matcha, { rim: 0, spec: 0 })}
        outline={false}
        position={[bandX, LID_HEIGHT / 2, 0]}
      />
      <ToonMesh
        geometry={seal}
        material={kit.toon(PALETTE.shu, { rim: 0, spec: 0.3 })}
        outline={0.014}
        position={[bandX, LID_HEIGHT + 0.035, BOX.depth * 0.18]}
      />
    </group>
  );
}

/** How the lid leans behind the box as a backrest. */
export const LID_POSE = {
  /** Hinge line behind the box. */
  hinge: [0, 0, -BOX.depth / 2 - 0.3] as const,
  /** Tilt around the hinge; the lid's top faces the viewer. */
  tilt: 64 * DEG,
} as const;

/**
 * The lid leaning behind the box. The group is the hinge: rotate it on X to
 * tilt the lid, move it on Y to lift it.
 */
export function Lid({ ref }: { ref?: Ref<Group> }) {
  return (
    <group ref={ref} position={[...LID_POSE.hinge]} rotation={[LID_POSE.tilt, 0, 0]}>
      <group position={[0, 0, -BOX.depth / 2 - 0.03]}>
        <LidBody />
      </group>
    </group>
  );
}
