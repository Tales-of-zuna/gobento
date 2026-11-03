'use client';
/**
 * React glue for the toon look: a context that shares one {@link ToonKit} per
 * scene, a hook that owns a geometry's lifetime, and {@link ToonMesh}, a mesh
 * with its inverted-hull outline attached.
 */
import type { ThreeElements } from '@react-three/fiber';
import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import type { BufferGeometry, Material } from 'three';
import { createHullGeometry, createToonKit, DEFAULT_OUTLINE, type ToonKit } from './toon';

const ToonKitContext = createContext<ToonKit | null>(null);

/** Provides a fresh {@link ToonKit} to the parts below and disposes it on unmount. */
export function ToonKitProvider({ children }: { children: ReactNode }) {
  const [kit] = useState(createToonKit);
  useEffect(() => () => kit.dispose(), [kit]);
  return <ToonKitContext value={kit}>{children}</ToonKitContext>;
}

/** The scene's shared materials. Must be used under a {@link ToonKitProvider}. */
export function useToonKit(): ToonKit {
  const kit = useContext(ToonKitContext);
  if (!kit) throw new Error('useToonKit must be used inside <ToonKitProvider>');
  return kit;
}

/**
 * Creates a geometry once per mount and disposes it on unmount. Parts pass
 * constant dimensions, so the geometry never needs rebuilding; remount the
 * component with a new `key` if it ever must change.
 */
export function useGeometry<T extends BufferGeometry>(create: () => T): T {
  const [geometry] = useState(create);
  useEffect(() => () => geometry.dispose(), [geometry]);
  return geometry;
}

/** Props of {@link ToonMesh}: a mesh plus its outline thickness. */
export type ToonMeshProps = Omit<ThreeElements['mesh'], 'geometry' | 'material' | 'children'> & {
  readonly geometry: BufferGeometry;
  readonly material: Material;
  /** Outline thickness in world units, or `false` for no outline. */
  readonly outline?: number | false;
};

/**
 * A mesh with an inverted-hull outline: a slightly inflated, back-faced copy
 * drawn in sumi ink behind it.
 */
export function ToonMesh({
  geometry,
  material,
  outline = DEFAULT_OUTLINE,
  ...props
}: ToonMeshProps) {
  const kit = useToonKit();
  const hasOutline = outline !== false;
  const hull = useMemo(
    () => (hasOutline ? createHullGeometry(geometry) : null),
    [geometry, hasOutline],
  );
  useEffect(() => () => hull?.dispose(), [hull]);
  return (
    <mesh geometry={geometry} material={material} {...props}>
      {hull && outline !== false && (
        <mesh geometry={hull} material={kit.outline(outline)} renderOrder={-1} />
      )}
    </mesh>
  );
}
