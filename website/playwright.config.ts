import { defineConfig, devices } from '@playwright/test';

const port = 4173;

/**
 * End-to-end tests run against the static export served under /gobento/,
 * as on GitHub Pages, never against the dev server.
 */
export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    baseURL: `http://127.0.0.1:${port}/gobento/`,
    trace: 'on-first-retry',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    command: `pnpm build && node scripts/serve-static.mjs ${port}`,
    url: `http://127.0.0.1:${port}/gobento/`,
    reuseExistingServer: !process.env.CI,
    timeout: 300_000,
  },
});
