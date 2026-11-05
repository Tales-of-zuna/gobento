// Lighthouse CI configuration: the performance and accessibility budget from
// the project-website spec ("Lighthouse budget"). Lighthouse emulates a mobile
// device by default, which is the mode the spec asks for.
//
// Run it against a finished static export (pnpm build) with:
//   npx --yes @lhci/cli@<version> autorun
// using the version pinned in .github/workflows/website.yml.
// LHCI_PORT picks the port of the local static server (default 4174).
const port = Number(process.env.LHCI_PORT ?? 4174);
const base = `http://127.0.0.1:${port}/gobento`;

module.exports = {
  ci: {
    collect: {
      startServerCommand: `node scripts/serve-static.mjs ${port}`,
      startServerReadyPattern: 'Serving',
      url: [`${base}/`, `${base}/configurator/`, `${base}/docs/getting-started/quick-start/`],
      // Lighthouse scores vary between runs, most on shared CI runners, where a
      // single run of a page scoring 0.94 locally has dropped to 0.89. The
      // assertions use the median run; with five runs per URL, three runs
      // have to be slow to fail the budget instead of two.
      numberOfRuns: 5,
      settings: {
        formFactor: 'mobile',
        // Containers (CI runners) often cannot use the Chrome sandbox.
        chromeFlags: '--no-sandbox --headless=new',
      },
    },
    assert: {
      assertions: {
        'categories:performance': ['error', { minScore: 0.9, aggregationMethod: 'median-run' }],
        'categories:accessibility': ['error', { minScore: 0.9, aggregationMethod: 'median-run' }],
        // Named on its own so a late font swap or a moving element fails with
        // its cause instead of only lowering the performance score.
        'cumulative-layout-shift': [
          'error',
          { maxNumericValue: 0.1, aggregationMethod: 'median-run' },
        ],
      },
    },
    upload: {
      target: 'filesystem',
      outputDir: '.lighthouseci/reports',
    },
  },
};
