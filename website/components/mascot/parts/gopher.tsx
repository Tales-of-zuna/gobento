'use client';
/**
 * The gopher of the approved mascot: a chubby daifuku body with the face on
 * its upper half and two paws reaching forward over the rim of the box.
 */
import type { ThreeElements } from '@react-three/fiber';
import type { Ref } from 'react';
import { LatheGeometry, Vector2, type Group } from 'three';
import { ToonMesh, useGeometry, useToonKit } from '../kit';
import { daifukuProfile, lathePose, type LatheProfile, type Vec3 } from '../math';
import { BODY_FACE, Face, type FaceRefs, type Surface } from './face';
import { Arm } from './limbs';

type GroupProps = Omit<ThreeElements['group'], 'children'>;

const DEG = Math.PI / 180;
/** Body size, and the height of the face centre on it. */
const BODY = { height: 2.6, width: 1.18, faceY: 1.74 } as const;
const PROFILE: LatheProfile = daifukuProfile(BODY.height, BODY.width);
const FACE_CENTER: Vec3 = [0, BODY.faceY, 0];
const bodySurface: Surface = (direction, push) => lathePose(PROFILE, FACE_CENTER, direction, push);

/** Where the shoulders sit and how the paws rest on the box rim. */
export const PAW = {
  x: 0.76,
  y: 1.26,
  z: 0.74,
  rest: { x: -100 * DEG, z: -8 * DEG },
  length: 0.36,
} as const;

/** The egg-shaped daifuku body, base at y = 0. */
function DaifukuBody({ color }: { color: string }) {
  const kit = useToonKit();
  const geometry = useGeometry(
    () =>
      new LatheGeometry(
        PROFILE.radius.map((r, i) => new Vector2(r, PROFILE.height[i] ?? 0)),
        72,
      ),
  );
  return <ToonMesh geometry={geometry} material={kit.toon(color, { rim: 0.4, spec: 0.55 })} />;
}

/** Props of {@link Gopher}. */
export type GopherProps = GroupProps &
  FaceRefs & {
    readonly color: string;
    /** The right paw, which waves. */
    readonly wavePaw?: Ref<Group>;
    /**
     * Group around the face. Turning it on Y slides the whole face around the
     * body, which is round, so the gopher looks left and right.
     */
    readonly face?: Ref<Group>;
    /** The whole gopher, for the idle bob and lean. */
    readonly ref?: Ref<Group>;
  };

/** Body, face, and paws. Base at y = 0, facing +Z. */
export function Gopher({ color, leftEye, rightEye, wavePaw, face, ref, ...props }: GopherProps) {
  return (
    <group ref={ref} {...props}>
      <DaifukuBody color={color} />
      <group ref={face}>
        <Face
          color={color}
          surface={bodySurface}
          layout={BODY_FACE}
          leftEye={leftEye}
          rightEye={rightEye}
        />
      </group>
      {([-1, 1] as const).map((side) => (
        <Arm
          key={side}
          ref={side > 0 ? wavePaw : undefined}
          color={color}
          length={PAW.length}
          position={[side * PAW.x, PAW.y, PAW.z]}
          rotation={[PAW.rest.x, 0, side * PAW.rest.z]}
        />
      ))}
    </group>
  );
}
