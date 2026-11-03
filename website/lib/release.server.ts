/**
 * Reads the release manifest at build time (server components only). The static export
 * bakes the result into the HTML, so a release needs a rebuild to show.
 */
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { parseReleaseManifest, type ReleaseStatus } from './release';

/** Path of the manifest; the build runs from `website/`, the manifest is at the repo root. */
const manifestPath = join(process.cwd(), '..', '.release-please-manifest.json');

/** The current release status, read once per build. */
export function releaseStatus(): ReleaseStatus {
  return parseReleaseManifest(readFileSync(manifestPath, 'utf8'));
}
