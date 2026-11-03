'use client';
/**
 * The gopher's face features, and a round head for the small icon. Every
 * feature is modelled facing +Z and placed by a {@link Surface} function, so
 * the same face fits the daifuku body and the round head.
 */
import type { ThreeElements } from '@react-three/fiber';
import { SphereGeometry, type Group } from 'three';
import { RoundedBoxGeometry } from 'three/addons/geometries/RoundedBoxGeometry.js';
import type { Ref } from 'react';
import { createSproutGeometry } from '../geometry';
import { ToonMesh, useGeometry, useToonKit } from '../kit';
import { ellipsoidPose, type SurfacePose, type Vec3 } from '../math';
import { PALETTE } from '../toon';

type GroupProps = Omit<ThreeElements['group'], 'children'>;

/** Head size: a sphere of radius 0.98, squashed into a soft mochi shape. */
const HEAD_RADIUS = 0.98;
const HEAD_SCALE: Vec3 = [1.08, 0.94, 1];
/** Radii of the head ellipsoid, used to stick features onto its surface. */
const HEAD_RADII: Vec3 = [
  HEAD_RADIUS * HEAD_SCALE[0],
  HEAD_RADIUS * HEAD_SCALE[1],
  HEAD_RADIUS * HEAD_SCALE[2],
];

/**
 * Where a feature lands on the surface it sits on: the point reached from the
 * face centre along `direction`, moved `push` outward.
 */
export type Surface = (direction: Vec3, push: number) => SurfacePose;

/** Surface of the round head, centred on the origin. */
const headSurface: Surface = (direction, push) => ellipsoidPose(HEAD_RADII, direction, push);

/** The round, slightly squashed head of the icon. */
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

/**
 * A glossy sumi eye with two highlights. The group's Y scale is the blink:
 * the mascot squashes it towards zero to close the eye.
 */
export function Eye({ ref, size, ...props }: GroupProps & { ref?: Ref<Group>; size: number }) {
  const kit = useToonKit();
  const ball = useGeometry(() => new SphereGeometry(size, 32, 24));
  const glint = useGeometry(() => new SphereGeometry(size * 0.34, 20, 16));
  const sparkle = useGeometry(() => new SphereGeometry(size * 0.15, 16, 12));
  const white = kit.flat('#ffffff');
  const z = size * 0.36;
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
        position={[-size * 0.28, size * 0.42, z]}
        scale={[1, 1, 0.4]}
      />
      <mesh
        geometry={sparkle}
        material={white}
        position={[size * 0.25, -size * 0.38, z]}
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

/** Where the features sit: directions from the face centre, and eye size. */
export interface FaceLayout {
  readonly eyeSize: number;
  readonly eye: Vec3;
  readonly blush: Vec3;
  readonly ear: Vec3;
  readonly muzzle: Vec3;
}

/** Layout of the approved mascot's face on its daifuku body. */
export const BODY_FACE: FaceLayout = {
  eyeSize: 0.18,
  eye: [0.36, 0.12, 1],
  blush: [0.62, -0.12, 1],
  ear: [0.72, 0.62, 0.15],
  muzzle: [0, -0.2, 1],
};

/** Layout on the round icon head. */
const HEAD_FACE: FaceLayout = {
  eyeSize: 0.17,
  eye: [0.36, 0.08, 1],
  blush: [0.62, -0.16, 1],
  ear: [0.72, 0.66, 0.15],
  muzzle: [0, -0.26, 1],
};

/** How far each feature sinks into (negative) or floats off the surface. */
const PUSH = { eye: -0.01, blush: -0.004, ear: -0.06, muzzle: -0.03 } as const;

/** Refs the mascot animates inside the face. */
export interface FaceRefs {
  readonly leftEye?: Ref<Group>;
  readonly rightEye?: Ref<Group>;
}

/** Props of {@link Face}. */
export type FaceProps = FaceRefs & {
  readonly color: string;
  readonly surface: Surface;
  readonly layout: FaceLayout;
};

const mirrored = (side: -1 | 1, [x, y, z]: Vec3): Vec3 => [side * x, y, z];

/** Group props for a feature placed on `surface`. */
function placed(surface: Surface, direction: Vec3, push: number) {
  const { position, quaternion } = surface(direction, push);
  return { position: [...position] as Vec3, quaternion: [...quaternion] as const };
}

/** Eyes, cheeks, ears, muzzle, and the sprout on top, placed on `surface`. */
export function Face({ color, surface, layout, leftEye, rightEye }: FaceProps) {
  const sprout = surface([0.05, 1, 0.05], 0).position;
  return (
    <group>
      <Eye
        ref={leftEye}
        size={layout.eyeSize}
        {...placed(surface, mirrored(-1, layout.eye), PUSH.eye)}
      />
      <Eye ref={rightEye} size={layout.eyeSize} {...placed(surface, layout.eye, PUSH.eye)} />
      {([-1, 1] as const).map((side) => (
        <group key={side}>
          <Blush {...placed(surface, mirrored(side, layout.blush), PUSH.blush)} />
          <Ear color={color} {...placed(surface, mirrored(side, layout.ear), PUSH.ear)} />
        </group>
      ))}
      <Muzzle {...placed(surface, layout.muzzle, PUSH.muzzle)} />
      <Sprout position={[sprout[0], sprout[1] - 0.06, sprout[2]]} rotation={[0, 0.5, -0.25]} />
    </group>
  );
}

/** The round head with its face: the small icon and favicon subject. */
export function GopherHead({ color }: { color: string }) {
  return (
    <group>
      <Head color={color} />
      <Face color={color} surface={headSurface} layout={HEAD_FACE} />
    </group>
  );
}
