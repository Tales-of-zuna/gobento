import { cn } from '@/lib/utils';
import s from './site.module.css';

/** The notice's text, shown on every page until the first release (spec "Pre-release notice"). */
export const PRERELEASE_TEXT = 'Early development · not released yet';

/** Props of {@link PrereleaseNotice}. */
export interface PrereleaseNoticeProps {
  /** Whether gobento has a release; from `releaseStatus()` in lib/release.server. */
  readonly released: boolean;
  readonly className?: string;
}

/**
 * The pre-release pill. Pages read the release manifest at build time and
 * pass `released`; the notice renders nothing once a version is out.
 */
export function PrereleaseNotice({ released, className }: PrereleaseNoticeProps) {
  if (released) return null;
  return (
    <p className={cn(s.pill, className)} data-prerelease-notice>
      <i aria-hidden="true" />
      {PRERELEASE_TEXT}
    </p>
  );
}
