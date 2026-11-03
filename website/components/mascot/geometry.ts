/**
 * Geometry builders for mascot parts that are not plain primitives: the
 * lacquered box walls, the onigiri, and the baran leaf sprout.
 */
import { ExtrudeGeometry, Path, Shape, Vector2 } from 'three';

/** Traces a rounded rectangle centred on the origin into `path`. */
function traceRoundedRect<T extends Path>(
  path: T,
  width: number,
  depth: number,
  radius: number,
): T {
  const x = -width / 2;
  const y = -depth / 2;
  path.moveTo(x + radius, y);
  path.lineTo(x + width - radius, y);
  path.quadraticCurveTo(x + width, y, x + width, y + radius);
  path.lineTo(x + width, y + depth - radius);
  path.quadraticCurveTo(x + width, y + depth, x + width - radius, y + depth);
  path.lineTo(x + radius, y + depth);
  path.quadraticCurveTo(x, y + depth, x, y + depth - radius);
  path.lineTo(x, y + radius);
  path.quadraticCurveTo(x, y, x + radius, y);
  return path;
}

/** Dimensions of a hollow rounded-rectangle ring. */
export interface RingSize {
  readonly width: number;
  readonly depth: number;
  readonly radius: number;
  readonly wall: number;
  readonly height: number;
  readonly bevel?: number;
}

/**
 * Hollow rounded-rectangle ring (the walls of a box), standing on y = 0 and
 * open along +Y, with bevelled edges.
 */
export function createRingGeometry({
  width,
  depth,
  radius,
  wall,
  height,
  bevel = 0.03,
}: RingSize): ExtrudeGeometry {
  const outer = traceRoundedRect(
    new Shape(),
    width - bevel * 2,
    depth - bevel * 2,
    Math.max(radius - bevel, 0.02),
  );
  outer.holes.push(
    traceRoundedRect(
      new Path(),
      width - wall * 2 + bevel * 2,
      depth - wall * 2 + bevel * 2,
      Math.max(radius - wall + bevel, 0.02),
    ),
  );
  const geometry = new ExtrudeGeometry(outer, {
    depth: height - bevel * 2,
    bevelEnabled: true,
    bevelThickness: bevel,
    bevelSize: bevel,
    bevelSegments: 3,
    curveSegments: 10,
  });
  geometry.rotateX(-Math.PI / 2);
  geometry.translate(0, bevel, 0);
  return geometry;
}

/**
 * Rice triangle for an onigiri: a rounded equilateral triangle extruded to
 * `depth`, centred in depth and standing on its bottom edge at y = 0.
 */
export function createOnigiriGeometry(size: number, depth: number): ExtrudeGeometry {
  const corners = [0, 1, 2].map((i) => {
    const angle = Math.PI / 2 + (i * 2 * Math.PI) / 3;
    return new Vector2(Math.cos(angle) * size, Math.sin(angle) * size);
  }) as [Vector2, Vector2, Vector2];
  const k = 0.32;
  const towards = (from: Vector2, to: Vector2) => from.clone().lerp(to, k);
  const shape = new Shape();
  const start = towards(corners[0], corners[1]);
  shape.moveTo(start.x, start.y);
  for (let i = 0; i < 3; i++) {
    const corner = corners[(i + 1) % 3] as Vector2;
    const next = corners[(i + 2) % 3] as Vector2;
    const before = towards(corner, corners[i] as Vector2);
    const after = towards(corner, next);
    shape.lineTo(before.x, before.y);
    shape.quadraticCurveTo(corner.x, corner.y, after.x, after.y);
  }
  const geometry = new ExtrudeGeometry(shape, {
    depth,
    bevelEnabled: true,
    bevelThickness: 0.08,
    bevelSize: 0.08,
    bevelSegments: 4,
    curveSegments: 12,
  });
  geometry.translate(0, 0, -depth / 2);
  geometry.computeBoundingBox();
  geometry.translate(0, -(geometry.boundingBox?.min.y ?? 0), 0);
  return geometry;
}

/** Baran (the green grass divider in a bento) cut as a jagged leaf sprout. */
export function createSproutGeometry(height = 0.42): ExtrudeGeometry {
  const w = 0.13;
  const outline: readonly (readonly [number, number])[] = [
    [-w, 0],
    [-w * 1.05, height * 0.55],
    [-w * 0.55, height * 0.75],
    [-w * 0.35, height],
    [0, height * 0.8],
    [w * 0.4, height * 1.08],
    [w * 0.6, height * 0.72],
    [w * 1.1, height * 0.62],
    [w, 0],
  ];
  const shape = new Shape(outline.map(([x, y]) => new Vector2(x, y)));
  shape.closePath();
  const geometry = new ExtrudeGeometry(shape, {
    depth: 0.035,
    bevelEnabled: true,
    bevelThickness: 0.02,
    bevelSize: 0.02,
    bevelSegments: 2,
  });
  geometry.translate(0, 0, -0.0175);
  return geometry;
}

/** A baran grass divider: a thin zigzag sheet `width` wide, standing on y = 0. */
export function createBaranGeometry(width = 0.42, height = 0.34): ExtrudeGeometry {
  const teeth = 5;
  const points = [new Vector2(-width / 2, 0)];
  for (let i = 0; i <= teeth; i++) {
    points.push(new Vector2(-width / 2 + (i / teeth) * width, height * (i % 2 ? 0.7 : 1)));
  }
  points.push(new Vector2(width / 2, 0));
  const shape = new Shape(points);
  shape.closePath();
  return new ExtrudeGeometry(shape, {
    depth: 0.02,
    bevelEnabled: true,
    bevelThickness: 0.01,
    bevelSize: 0.01,
    bevelSegments: 1,
  });
}
