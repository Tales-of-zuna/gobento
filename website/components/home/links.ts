/** Destinations linked from the home page. */
import { gitConfig } from '@/lib/shared';

const repo = `https://github.com/${gitConfig.user}/${gitConfig.repo}`;

/** The repository on GitHub. */
export const githubUrl = repo;

/** Route of the configurator page. */
export const configuratorRoute = '/configurator/';

/** A file on the default branch, as shown on GitHub. */
export function repoFile(path: string): string {
  return `${repo}/blob/${gitConfig.branch}/${path}`;
}

/** The quick-start command shown in the hero (the planned CLI interface). */
export const QUICK_START = 'gobento new shop --module example.com/shop';
