'use client';
/** Stubby arms. */
import type { ThreeElements } from '@react-three/fiber';
import type { Ref } from 'react';
import { CapsuleGeometry, SphereGeometry, type Group } from 'three';
import { ToonMesh, useGeometry, useToonKit } from '../kit';
import { PALETTE } from '../toon';

type GroupProps = Omit<ThreeElements['group'], 'children'>;

const ARM_RADIUS = 0.15;

/**
 * A short nub arm hanging from its shoulder pivot (the group origin) along
 * -Y, with a cream paw. Rotate the group to pose or wave it.
 */
export function Arm({
  color,
  length = 0.42,
  ref,
  ...props
}: GroupProps & { color: string; length?: number; ref?: Ref<Group> }) {
  const kit = useToonKit();
  const limb = useGeometry(() => new CapsuleGeometry(ARM_RADIUS, length, 8, 20));
  const paw = useGeometry(() => new SphereGeometry(ARM_RADIUS * 0.62, 20, 14));
  return (
    <group ref={ref} {...props}>
      <ToonMesh
        geometry={limb}
        material={kit.toon(color, { rim: 0.35, spec: 0.4 })}
        outline={0.026}
        position={[0, -length / 2, 0]}
      />
      <ToonMesh
        geometry={paw}
        material={kit.toon(PALETTE.cream, { rim: 0, spec: 0.2 })}
        outline={false}
        position={[0, -length - ARM_RADIUS * 0.55, ARM_RADIUS * 0.25]}
        scale={[1, 0.7, 0.8]}
      />
    </group>
  );
}
