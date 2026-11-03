'use client';
/**
 * The gopher's head and face features. Every feature is modelled facing +Z
 * and placed on the head with {@link ellipsoidPose}.
 */
import type { ThreeElements } from '@react-three/fiber';
import { SphereGeometry, type Group } from 'three';
import { RoundedBoxGeometry } from 'three/addons/geometries/RoundedBoxGeometry.js';
import type { Ref } from 'react';
import { createSproutGeometry } from '../geometry';
import { ToonMesh, useGeometry, useToonKit } from '../kit';
import { ellipsoidPose, type Vec3 } from '../math';
import { PALETTE } from '../toon';

type GroupProps = Omit<ThreeElements['group'], 'children'>;

/** Head size: a sphere of radius 0.98, squashed into a soft mochi shape. */
const HEAD_RADIUS = 0.98;
const HEAD_SCALE: Vec3 = [1.08, 0.94, 1];
/** Radii of the head ellipsoid, used to stick features onto its surface. */
export const HEAD_RADII: Vec3 = [
  HEAD_RADIUS * HEAD_SCALE[0],
  HEAD_RADIUS * HEAD_SCALE[1],
  HEAD_RADIUS * HEAD_SCALE[2],
];

/** Group props that put a feature on the head along `direction`. */
function onHead(direction: Vec3, push: number): Pick<GroupProps, 'position' | 'quaternion'> {
  const { position, quaternion } = ellipsoidPose(HEAD_RADII, direction, push);
  return { position: [...position], quaternion: [...quaternion] };
}

/** The round, slightly squashed head. */
export function Head({ color }: { color: string }) {
  const kit = useToonKit();
  const geometry = useGeometry(() => new SphereGeometry(HEAD_RADIUS, 64, 48));
  return (
    <ToonMesh
      geometry={geometry}
      material={kit.toon(color, { rim: 0.4, spec: 0.55 })}
      scale={HEAD_SCALE}
    />
  );
}

const EYE_SIZE = 0.17;

/**
 * A glossy sumi eye with two highlights. The group's Y scale is the blink:
 * the mascot squashes it towards zero to close the eye.
 */
export function Eye({ ref, ...props }: GroupProps & { ref?: Ref<Group> }) {
  const kit = useToonKit();
  const ball = useGeometry(() => new SphereGeometry(EYE_SIZE, 32, 24));
  const glint = useGeometry(() => new SphereGeometry(EYE_SIZE * 0.34, 20, 16));
  const sparkle = useGeometry(() => new SphereGeometry(EYE_SIZE * 0.15, 16, 12));
  const white = kit.flat('#ffffff');
  const z = EYE_SIZE * 0.36;
  return (
    <group ref={ref} {...props}>
      <ToonMesh
        geometry={ball}
        material={kit.toon(PALETTE.sumi, { rim: 0, spec: 0 })}
        outline={false}
        scale={[0.82, 1.08, 0.42]}
      />
      <mesh
        geometry={glint}
        material={white}
        position={[-EYE_SIZE * 0.28, EYE_SIZE * 0.42, z]}
        scale={[1, 1, 0.4]}
      />
      <mesh
        geometry={sparkle}
        material={white}
        position={[EYE_SIZE * 0.25, -EYE_SIZE * 0.38, z]}
        scale={[1, 1, 0.4]}
      />
    </group>
  );
}

/** A soft pink cheek. */
export function Blush(props: GroupProps) {
  const kit = useToonKit();
  const geometry = useGeometry(() => new SphereGeometry(0.13, 24, 16));
  return (
    <group {...props}>
      <mesh geometry={geometry} material={kit.flat(PALETTE.blush)} scale={[1.25, 0.72, 0.12]} />
    </group>
  );
}

/** Puffy two-lobed muzzle with a button nose and two buck teeth. */
export function Muzzle(props: GroupProps) {
  const kit = useToonKit();
  const lobe = useGeometry(() => new SphereGeometry(0.19, 32, 20));
  const nose = useGeometry(() => new SphereGeometry(0.085, 24, 16));
  const tooth = useGeometry(() => new RoundedBoxGeometry(0.095, 0.12, 0.05, 3, 0.022));
  const cream = kit.toon(PALETTE.cream, { rim: 0.1, spec: 0.3 });
  const enamel = kit.toon(PALETTE.tooth, { rim: 0, spec: 0 });
  return (
    <group {...props}>
      {[-1, 1].map((side) => (
        <ToonMesh
          key={`lobe${side}`}
          geometry={lobe}
          material={cream}
          outline={0.022}
          position={[side * 0.13, 0, 0]}
          scale={[1, 0.8, 0.55]}
        />
      ))}
      <ToonMesh
        geometry={nose}
        material={kit.toon(PALETTE.sumi, { rim: 0, spec: 0.9 })}
        outline={false}
        position={[0, 0.11, 0.09]}
        scale={[1.35, 0.9, 0.8]}
      />
      {[-1, 1].map((side) => (
        <ToonMesh
          key={`tooth${side}`}
          geometry={tooth}
          material={enamel}
          outline={0.016}
          position={[side * 0.051, -0.17, 0.035]}
          rotation={[0, 0, side * -0.04]}
        />
      ))}
    </group>
  );
}

/** A small round ear with a pink inner. */
export function Ear({ color, ...props }: GroupProps & { color: string }) {
  const kit = useToonKit();
  const outer = useGeometry(() => new SphereGeometry(0.2, 28, 20));
  const inner = useGeometry(() => new SphereGeometry(0.12, 24, 16));
  return (
    <group {...props}>
      <ToonMesh
        geometry={outer}
        material={kit.toon(color, { rim: 0.3, spec: 0.2 })}
        outline={0.026}
        scale={[1, 0.92, 0.55]}
      />
      <ToonMesh
        geometry={inner}
        material={kit.toon(PALETTE.blush, { rim: 0, spec: 0 })}
        outline={false}
        position={[0, 0, 0.085]}
        scale={[1, 0.9, 0.3]}
      />
    </group>
  );
}

/** A baran leaf growing from the top of the head like a hair sprout. */
export function Sprout(props: GroupProps) {
  const kit = useToonKit();
  const geometry = useGeometry(() => createSproutGeometry());
  return (
    <group {...props}>
      <ToonMesh
        geometry={geometry}
        material={kit.toon(PALETTE.matcha, { rim: 0.3, spec: 0.4, shadowTint: '#2b4a40' })}
        outline={0.02}
      />
    </group>
  );
}

/** Where the face features sit on the head (directions from its centre). */
const FACE = {
  eye: { direction: [0.36, 0.08, 1] as Vec3, push: -0.01 },
  blush: { direction: [0.62, -0.16, 1] as Vec3, push: -0.004 },
  ear: { direction: [0.72, 0.66, 0.15] as Vec3, push: -0.06 },
  muzzle: { direction: [0, -0.26, 1] as Vec3, push: -0.03 },
} as const;

const SPROUT_ROOT = ellipsoidPose(HEAD_RADII, [0.05, 1, 0.05]).position;

/** Refs the mascot animates inside the face. */
export interface FaceRefs {
  readonly leftEye?: Ref<Group>;
  readonly rightEye?: Ref<Group>;
}

/** Head plus all face features, centred on the head. */
export function GopherHead({ color, leftEye, rightEye }: { color: string } & FaceRefs) {
  const mirrored = (side: -1 | 1, [x, y, z]: Vec3): Vec3 => [side * x, y, z];
  return (
    <group>
      <Head color={color} />
      <Eye ref={leftEye} {...onHead(mirrored(-1, FACE.eye.direction), FACE.eye.push)} />
      <Eye ref={rightEye} {...onHead(FACE.eye.direction, FACE.eye.push)} />
      {([-1, 1] as const).map((side) => (
        <group key={side}>
          <Blush {...onHead(mirrored(side, FACE.blush.direction), FACE.blush.push)} />
          <Ear color={color} {...onHead(mirrored(side, FACE.ear.direction), FACE.ear.push)} />
        </group>
      ))}
      <Muzzle {...onHead(FACE.muzzle.direction, FACE.muzzle.push)} />
      <Sprout
        position={[SPROUT_ROOT[0], SPROUT_ROOT[1] - 0.06, SPROUT_ROOT[2]]}
        rotation={[0, 0.5, -0.25]}
      />
    </group>
  );
}
