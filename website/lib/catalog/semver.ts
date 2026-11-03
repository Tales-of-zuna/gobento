/**
 * Semantic version comparison and range evaluation, mirroring the CLI's use
 * of golang.org/x/mod/semver (see add-generator-foundation design D4).
 *
 * Module versions use Go's semver form (`v1.2.3`, `v1.2.3-rc.1`, shorthand
 * `v1` and `v1.2`). Go versions use the toolchain form (`1.27`, `1.27.1`) and
 * are normalized to semver (`v1.27.0`) before comparison.
 */

/** A parsed semantic version. Build metadata is dropped, as it never affects order. */
export interface Version {
  readonly major: number;
  readonly minor: number;
  readonly patch: number;
  readonly prerelease: readonly string[];
}

/** Comparison operators allowed in a range. */
export type Operator = '>=' | '>' | '<=' | '<' | '=';

/** One `<operator><version>` term of a range. */
export interface Comparator {
  readonly operator: Operator;
  readonly version: Version;
}

/** Which version syntax a range or version uses. */
export type VersionKind = 'module' | 'go';

const num = '(0|[1-9]\\d*)';
const ident = '[0-9A-Za-z-]+';
const semverPattern = new RegExp(
  `^v${num}(?:\\.${num}(?:\\.${num}(?:-(${ident}(?:\\.${ident})*))?(?:\\+${ident}(?:\\.${ident})*)?)?)?$`,
);
const goPattern = new RegExp(`^${num}\\.${num}(?:\\.${num})?$`);
const comparatorPattern = /^(>=|<=|>|<|=)(.+)$/;

/**
 * Parses a module version such as `v1.61.1`. Shorthands `v1` and `v1.2` mean
 * `v1.0.0` and `v1.2.0`, as in golang.org/x/mod/semver. Returns undefined for
 * anything else, including numeric prerelease identifiers with leading zeros.
 */
export function parseModuleVersion(text: string): Version | undefined {
  const match = semverPattern.exec(text);
  if (!match) return undefined;
  const [, major, minor = '0', patch = '0', prerelease] = match;
  const parts = prerelease === undefined ? [] : prerelease.split('.');
  if (parts.some((part) => /^0\d+$/.test(part))) return undefined;
  return { major: Number(major), minor: Number(minor), patch: Number(patch), prerelease: parts };
}

/** Parses a Go version such as `1.27` (normalized to v1.27.0) or `1.27.1`. */
export function parseGoVersion(text: string): Version | undefined {
  const match = goPattern.exec(text);
  if (!match) return undefined;
  const [, major, minor, patch = '0'] = match;
  return { major: Number(major), minor: Number(minor), patch: Number(patch), prerelease: [] };
}

/** Parses a version of the given kind. */
export function parseVersion(text: string, kind: VersionKind): Version | undefined {
  return kind === 'go' ? parseGoVersion(text) : parseModuleVersion(text);
}

function compareIdentifiers(a: string, b: string): number {
  const aNumeric = /^\d+$/.test(a);
  const bNumeric = /^\d+$/.test(b);
  if (aNumeric && bNumeric) return Math.sign(Number(a) - Number(b));
  if (aNumeric !== bNumeric) return aNumeric ? -1 : 1;
  return a < b ? -1 : a > b ? 1 : 0;
}

/** Orders two versions: negative if a < b, zero if equal, positive if a > b. */
export function compareVersions(a: Version, b: Version): number {
  const core = a.major - b.major || a.minor - b.minor || a.patch - b.patch;
  if (core !== 0) return Math.sign(core);
  if (a.prerelease.length === 0 || b.prerelease.length === 0) {
    // A release sorts after any of its prereleases.
    return Math.sign(b.prerelease.length - a.prerelease.length);
  }
  for (let i = 0; i < Math.min(a.prerelease.length, b.prerelease.length); i++) {
    const order = compareIdentifiers(a.prerelease[i]!, b.prerelease[i]!);
    if (order !== 0) return order;
  }
  return Math.sign(a.prerelease.length - b.prerelease.length);
}

/**
 * Parses a range: comparators separated by spaces, all of which must hold
 * (for example `>=v1.61.0 <v2.0.0`). Returns undefined if the range is empty
 * or any comparator is malformed.
 */
export function parseRange(range: string, kind: VersionKind): Comparator[] | undefined {
  const terms = range.trim().split(/\s+/).filter(Boolean);
  if (terms.length === 0) return undefined;
  const comparators: Comparator[] = [];
  for (const term of terms) {
    const match = comparatorPattern.exec(term);
    const version = match && parseVersion(match[2]!, kind);
    if (!match || !version) return undefined;
    comparators.push({ operator: match[1] as Operator, version });
  }
  return comparators;
}

function holds({ operator, version }: Comparator, candidate: Version): boolean {
  const order = compareVersions(candidate, version);
  switch (operator) {
    case '>=':
      return order >= 0;
    case '>':
      return order > 0;
    case '<=':
      return order <= 0;
    case '<':
      return order < 0;
    case '=':
      return order === 0;
  }
}

/**
 * Reports whether a version lies inside a range, both written in the given
 * kind's syntax. A malformed version or range never satisfies.
 */
export function satisfies(version: string, range: string, kind: VersionKind): boolean {
  const candidate = parseVersion(version, kind);
  const comparators = parseRange(range, kind);
  if (!candidate || !comparators) return false;
  return comparators.every((comparator) => holds(comparator, candidate));
}

/** Reports whether two version strings of the given kind denote the same version. */
export function sameVersion(a: string, b: string, kind: VersionKind): boolean {
  const left = parseVersion(a, kind);
  const right = parseVersion(b, kind);
  return left !== undefined && right !== undefined && compareVersions(left, right) === 0;
}
