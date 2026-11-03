'use client';
import { RootProvider } from 'fumadocs-ui/provider/next';
import dynamic from 'next/dynamic';
import { type ReactNode } from 'react';

/**
 * The search dialog and its Orama client are only downloaded when search is
 * first opened, so no page pays for them at load.
 */
const SearchDialog = dynamic(() => import('@/components/search'), { ssr: false });

export function Provider({ children }: { children: ReactNode }) {
  return <RootProvider search={{ SearchDialog }}>{children}</RootProvider>;
}
