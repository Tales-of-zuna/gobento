import { Figtree, JetBrains_Mono } from 'next/font/google';
import localFont from 'next/font/local';
import type { Metadata } from 'next';
import { headStill, stillUrl } from '@/components/mascot/stills';
import { Provider } from '@/components/provider';
import { cn } from '@/lib/utils';
import './global.css';

/**
 * Display type: headings and the wordmark. Zen Maru Gothic is a Japanese
 * font: from Google Fonts it arrives as ~120 slices per weight, which
 * next/font/google preloads all of (about 4 MB). The site shows no Japanese
 * text, so only its Latin slice is self-hosted (app/fonts, SIL OFL 1.1).
 */
const display = localFont({
  src: [
    { path: './fonts/zen-maru-gothic-latin-500.woff2', weight: '500', style: 'normal' },
    { path: './fonts/zen-maru-gothic-latin-700.woff2', weight: '700', style: 'normal' },
    { path: './fonts/zen-maru-gothic-latin-900.woff2', weight: '900', style: 'normal' },
  ],
  variable: '--font-display',
  display: 'swap',
  fallback: ['ui-rounded', 'sans-serif'],
});

/** Body text. */
const body = Figtree({ subsets: ['latin'], variable: '--font-body', display: 'swap' });

/**
 * Code; ligatures are turned off in global.css. Not preloaded: its 40 KB file
 * would compete with the text. `optional` instead of `swap`: a late swap
 * reflows paragraphs full of inline code (a layout shift Lighthouse counts),
 * so on a slow first load code stays in the fallback monospace font.
 */
const mono = JetBrains_Mono({
  subsets: ['latin'],
  variable: '--font-mono',
  display: 'optional',
  preload: false,
});

export const metadata: Metadata = {
  title: { default: 'gobento', template: '%s | gobento' },
  description: 'Compose a ready-to-build Go project from the stack you choose.',
  // Next does not prefix metadata icons with basePath; stillUrl does.
  icons: {
    icon: [{ url: stillUrl(headStill(64), 'png'), sizes: '64x64', type: 'image/png' }],
    apple: [{ url: stillUrl(headStill(128), 'png'), sizes: '128x128', type: 'image/png' }],
  },
};

export default function Layout({ children }: LayoutProps<'/'>) {
  return (
    <html
      lang="en"
      className={cn('font-sans', display.variable, body.variable, mono.variable)}
      suppressHydrationWarning
    >
      <head>
        {/*
         * Marks <html> before the first paint when scripts run, so entrance
         * animations can play. Without JavaScript the class never appears and
         * the server-rendered page stays fully visible (components/home/motion).
         */}
        <script
          dangerouslySetInnerHTML={{
            __html: "document.documentElement.classList.add('motion-ready')",
          }}
        />
      </head>
      <body className="flex min-h-screen flex-col">
        <Provider>{children}</Provider>
      </body>
    </html>
  );
}
