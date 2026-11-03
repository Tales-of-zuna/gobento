import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { parseReleaseManifest } from './release';

describe('parseReleaseManifest', () => {
  it('treats 0.0.0 as unreleased', () => {
    expect(parseReleaseManifest('{".": "0.0.0"}')).toEqual({ version: '0.0.0', released: false });
  });

  it('treats any other version as released', () => {
    expect(parseReleaseManifest('{".": "0.1.0"}')).toEqual({ version: '0.1.0', released: true });
  });

  it.each(['{}', '{".": 1}', '{".": "latest"}', '[]', 'null'])('rejects %s', (text) => {
    expect(() => parseReleaseManifest(text)).toThrow(/release manifest/);
  });

  it('rejects invalid JSON', () => {
    expect(() => parseReleaseManifest('not json')).toThrow();
  });

  it("reads the repository's manifest", () => {
    const text = readFileSync(
      join(import.meta.dirname, '..', '..', '.release-please-manifest.json'),
      'utf8',
    );
    expect(parseReleaseManifest(text).version).toMatch(/^\d+\.\d+\.\d+/);
  });
});
