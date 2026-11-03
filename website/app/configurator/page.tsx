import type { Metadata } from 'next';
import { Configurator } from '@/components/configurator/Configurator';
import { catalog } from '@/lib/catalog/data';
import { releaseStatus } from '@/lib/release.server';

export const metadata: Metadata = {
  title: 'Configurator',
  description: 'Choose your stack and copy the exact gobento new command for it.',
};

export default function ConfiguratorPage() {
  const { released } = releaseStatus();
  return <Configurator catalog={catalog} released={released} />;
}
