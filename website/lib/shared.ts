/** Display name used in the navigation bar. */
export const appName = 'gobento';

/** Route of the documentation section. */
export const docsRoute = '/docs';

/** GitHub repository the site documents and edits link to. */
export const gitConfig = {
  user: 'Tales-of-zuna',
  repo: 'gobento',
  branch: 'main',
};

/** Directory of the docs sources, relative to the repository root. */
export const docsContentDir = 'website/content/docs';

/** GitHub editor URL for a docs source file, given its path inside {@link docsContentDir}. */
export function editUrl(path: string): string {
  const { user, repo, branch } = gitConfig;
  return `https://github.com/${user}/${repo}/edit/${branch}/${docsContentDir}/${path}`;
}
