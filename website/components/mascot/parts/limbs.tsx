'use client';
/** Stubby arms and feet. */
import type { ThreeElements } from '@react-three/fiber';
import type { Ref } from 'react';
import { CapsuleGeometry, SphereGeometry, type Group } from 'three';
import { ToonMesh, useGeometry, useToonKit } from '../kit';
import { PALETTE } from '../toon';

type GroupProps = Omit<ThreeElements['group'], 'children'>;

const ARM_RADIUS = 0.15;
const ARM_LENGTH = 0.42;

/**
 * A short nub arm hanging from its shoulder pivot (the group origin) along
 * -Y, with a cream paw. Rotate the group to pose or wave it.
 */
export function Arm({ color, ref, ...props }: GroupProps & { color: string; ref?: Ref<Group> }) {
  const kit = useToonKit();
  const limb = useGeometry(() => new CapsuleGeometry(ARM_RADIUS, ARM_LENGTH, 8, 20));
  const paw = useGeometry(() => new SphereGeometry(ARM_RADIUS * 0.62, 20, 14));
  return (
    <group ref={ref} {...props}>
      <ToonMesh
        geometry={limb}
        material={kit.toon(color, { rim: 0.35, spec: 0.4 })}
        outline={0.026}
        position={[0, -ARM_LENGTH / 2, 0]}
      />
      <ToonMesh
        geometry={paw}
        material={kit.toon(PALETTE.cream, { rim: 0, spec: 0.2 })}
        outline={false}
        position={[0, -ARM_LENGTH - ARM_RADIUS * 0.55, ARM_RADIUS * 0.25]}
        scale={[1, 0.7, 0.8]}
      />
    </group>
  );
}

/** A round foot with a cream pad peeking out at the front. */
export function Foot({ color, ...props }: GroupProps & { color: string }) {
  const kit = useToonKit();
  const foot = useGeometry(() => new SphereGeometry(0.24, 28, 18));
  const pad = useGeometry(() => new SphereGeometry(0.12, 20, 12));
  return (
    <group {...props}>
      <ToonMesh
        geometry={foot}
        material={kit.toon(color, { rim: 0.3, spec: 0.3 })}
        outline={0.026}
        scale={[1, 0.55, 1.3]}
      />
      <ToonMesh
        geometry={pad}
        material={kit.toon(PALETTE.cream, { rim: 0, spec: 0 })}
        outline={false}
        position={[0, -0.02, 0.22]}
        scale={[1, 0.4, 0.8]}
      />
    </group>
  );
}
