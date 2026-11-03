import Link from 'next/link';
import type { ReleaseStatus } from '@/lib/release';
import { installationRoute, releasesUrl } from './links';
import s from './home.module.css';

/** Hero note before the first release (spec "Pre-release notice"). */
export const PLANNED_NOTE = 'Planned interface. There is nothing to install yet.';

/** Closing lead before the first release. */
export const PLANNED_CLOSING =
  'Try the configurator now; the CLI follows when the first release is ready.';

/** Closing lead once a version is out. */
export const RELEASED_CLOSING =
  'Pick your stack in the configurator, then paste the command it builds into your terminal.';

/**
 * The note under the hero's quick-start command. It follows the release
 * manifest (lib/release.server), so the copy is right before and after a
 * release without editing it (spec "Release-aware install guidance").
 */
export function HeroInstallNote({ release }: { release: ReleaseStatus }) {
  if (!release.released) return <p className={s.note}>{PLANNED_NOTE}</p>;
  return (
    <p className={s.note} data-install-note>
      gobento v{release.version} is out. <Link href={installationRoute}>Install it</Link> with{' '}
      <code>go install</code> or a <a href={releasesUrl}>prebuilt binary</a>.
    </p>
  );
}

/** The lead of the closing call to action, worded for the release status. */
export function closingLead(release: ReleaseStatus): string {
  return release.released ? RELEASED_CLOSING : PLANNED_CLOSING;
}
