/**
 * Release status of gobento, read from release-please's manifest
 * (`.release-please-manifest.json` at the repository root). Version `0.0.0`
 * means nothing has been released yet (design: pre-release notice).
 */

/** Whether gobento has a published release, and which version is current. */
export interface ReleaseStatus {
  readonly version: string;
  readonly released: boolean;
}

/** Version release-please starts from before the first release. */
export const UNRELEASED_VERSION = '0.0.0';

/**
 * Parses the manifest's text. The root package (`.`) carries the version; a
 * manifest that is missing it or is not valid JSON fails loudly so a broken
 * file cannot silently hide or show the notice.
 */
export function parseReleaseManifest(text: string): ReleaseStatus {
  const data: unknown = JSON.parse(text);
  const version =
    typeof data === 'object' && data !== null ? (data as Record<string, unknown>)['.'] : undefined;
  if (typeof version !== 'string' || !/^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$/.test(version)) {
    throw new Error('release manifest: missing or malformed version for "."');
  }
  return { version, released: version !== UNRELEASED_VERSION };
}
