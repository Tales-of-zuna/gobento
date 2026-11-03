import { Geist } from 'next/font/google';
import type { Metadata } from 'next';
import { headStill, stillUrl } from '@/components/mascot/stills';
import { Provider } from '@/components/provider';
import { cn } from '@/lib/utils';
import './global.css';

const geist = Geist({ subsets: ['latin'], variable: '--font-sans' });

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
    <html lang="en" className={cn('font-sans', geist.variable)} suppressHydrationWarning>
      <body className="flex min-h-screen flex-col">
        <Provider>{children}</Provider>
      </body>
    </html>
  );
}
