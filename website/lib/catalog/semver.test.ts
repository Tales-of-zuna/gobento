import { describe, expect, it } from 'vitest';
import {
  compareVersions,
  parseGoVersion,
  parseModuleVersion,
  parseRange,
  sameVersion,
  satisfies,
  type VersionKind,
} from './semver';

describe('parseModuleVersion', () => {
  it.each([
    ['v1.61.1', { major: 1, minor: 61, patch: 1, prerelease: [] }],
    ['v1', { major: 1, minor: 0, patch: 0, prerelease: [] }],
    ['v1.2', { major: 1, minor: 2, patch: 0, prerelease: [] }],
    ['v1.2.3-rc.1', { major: 1, minor: 2, patch: 3, prerelease: ['rc', '1'] }],
    ['v1.2.3+build.5', { major: 1, minor: 2, patch: 3, prerelease: [] }],
  ])('parses %s', (text, expected) => {
    expect(parseModuleVersion(text)).toEqual(expected);
  });

  it.each(['1.2.3', 'v01.2.3', 'v1.2.3-01', 'v1.2-rc.1', 'v1.2.3-', 'v', '', 'latest'])(
    'rejects %j',
    (text) => {
      expect(parseModuleVersion(text)).toBeUndefined();
    },
  );
});

describe('parseGoVersion', () => {
  it('normalizes 1.27 to v1.27.0', () => {
    expect(parseGoVersion('1.27')).toEqual(parseModuleVersion('v1.27.0'));
  });

  it.each(['1.27.1', '1.26'])('accepts %s', (text) => {
    expect(parseGoVersion(text)).toBeDefined();
  });

  it.each(['v1.27', '1', '1.27rc1', 'go1.27', '1.027'])('rejects %j', (text) => {
    expect(parseGoVersion(text)).toBeUndefined();
  });
});

describe('compareVersions', () => {
  it.each([
    ['v1.2.3', 'v1.2.3', 0],
    ['v1.2.3', 'v1.2.4', -1],
    ['v1.10.0', 'v1.9.0', 1],
    ['v2.0.0', 'v1.99.99', 1],
    ['v1.2', 'v1.2.0', 0],
    ['v1.2.3-rc.1', 'v1.2.3', -1],
    ['v1.2.3-alpha', 'v1.2.3-beta', -1],
    ['v1.2.3-rc.2', 'v1.2.3-rc.10', -1],
    ['v1.2.3-1', 'v1.2.3-alpha', -1],
    ['v1.2.3-rc', 'v1.2.3-rc.1', -1],
    ['v1.2.3+a', 'v1.2.3+b', 0],
  ])('%s vs %s → %i', (a, b, expected) => {
    expect(compareVersions(parseModuleVersion(a)!, parseModuleVersion(b)!)).toBe(expected);
  });
});

describe('satisfies', () => {
  const cases: [version: string, range: string, kind: VersionKind, expected: boolean][] = [
    // Each comparator, at and around its boundary.
    ['v1.61.0', '>=v1.61.0', 'module', true],
    ['v1.60.9', '>=v1.61.0', 'module', false],
    ['v1.61.0', '>v1.61.0', 'module', false],
    ['v1.61.1', '>v1.61.0', 'module', true],
    ['v1.61.0', '<=v1.61.0', 'module', true],
    ['v1.61.1', '<=v1.61.0', 'module', false],
    ['v1.99.0', '<v2.0.0', 'module', true],
    ['v2.0.0', '<v2.0.0', 'module', false],
    ['v1.31.1', '=v1.31.1', 'module', true],
    ['v1.31.2', '=v1.31.1', 'module', false],
    // Comparators are ANDed.
    ['v1.61.1', '>=v1.61.0 <v2.0.0', 'module', true],
    ['v2.0.0', '>=v1.61.0 <v2.0.0', 'module', false],
    ['v1.60.0', '>=v1.61.0 <v2.0.0', 'module', false],
    // A prerelease sorts below its release.
    ['v2.0.0-rc.1', '<v2.0.0', 'module', true],
    // Go versions are normalized: 1.27 is v1.27.0.
    ['1.27', '>=1.26 <1.28', 'go', true],
    ['1.26', '>=1.26 <1.28', 'go', true],
    ['1.28', '>=1.26 <1.28', 'go', false],
    ['1.25', '>=1.26 <1.28', 'go', false],
    ['1.27.3', '>=1.26 <1.28', 'go', true],
    ['1.26.0', '=1.26', 'go', true],
    // Malformed input never satisfies.
    ['v1.61.1', '~v1.61.0', 'module', false],
    ['latest', '>=v1.0.0', 'module', false],
    ['1.27', '>=v1.26.0', 'go', false],
    ['1.27', '', 'go', false],
  ];

  it.each(cases)('%s in %j (%s) → %s', (version, range, kind, expected) => {
    expect(satisfies(version, range, kind)).toBe(expected);
  });
});

describe('parseRange', () => {
  it('splits on any whitespace', () => {
    expect(parseRange('  >=v1.0.0   <v2.0.0 ', 'module')).toHaveLength(2);
  });

  it.each(['', '>=', '>>v1.0.0', 'v1.0.0', '>=v1.0.0,<v2.0.0'])('rejects %j', (range) => {
    expect(parseRange(range, 'module')).toBeUndefined();
  });
});

describe('sameVersion', () => {
  it.each([
    ['1.27', '1.27.0', 'go', true],
    ['1.27', '1.26', 'go', false],
    ['v1.61.1', 'v1.61.1', 'module', true],
    ['oops', 'oops', 'module', false],
  ] as const)('%s and %s (%s) → %s', (a, b, kind, expected) => {
    expect(sameVersion(a, b, kind)).toBe(expected);
  });
});
