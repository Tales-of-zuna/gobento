/** Destinations linked from the site chrome and pages. */
import { docsRoute, gitConfig } from '@/lib/shared';

const repo = `https://github.com/${gitConfig.user}/${gitConfig.repo}`;

/** The repository on GitHub. */
export const githubUrl = repo;

/** Route of the configurator page. */
export const configuratorRoute = '/configurator/';

/** Route of the docs' first page. */
export const docsHomeRoute = `${docsRoute}/`;

/** A file on the default branch, as shown on GitHub. */
export function repoFile(path: string): string {
  return `${repo}/blob/${gitConfig.branch}/${path}`;
}
