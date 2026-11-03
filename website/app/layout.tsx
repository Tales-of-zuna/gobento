import { Figtree, JetBrains_Mono, Zen_Maru_Gothic } from 'next/font/google';
import type { Metadata } from 'next';
import { headStill, stillUrl } from '@/components/mascot/stills';
import { Provider } from '@/components/provider';
import { cn } from '@/lib/utils';
import './global.css';

/** Display type: headings and the wordmark. */
const display = Zen_Maru_Gothic({
  subsets: ['latin'],
  weight: ['500', '700', '900'],
  variable: '--font-display',
  display: 'swap',
});

/** Body text. */
const body = Figtree({ subsets: ['latin'], variable: '--font-body', display: 'swap' });

/** Code; ligatures are turned off in global.css. */
const mono = JetBrains_Mono({ subsets: ['latin'], variable: '--font-mono', display: 'swap' });

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
      <body className="flex min-h-screen flex-col">
        <Provider>{children}</Provider>
      </body>
    </html>
  );
}
