import type { Metadata } from 'next';
import { Configurator } from '@/components/configurator/Configurator';
import { catalog } from '@/lib/catalog/data';

export const metadata: Metadata = {
  title: 'Configurator',
  description: 'Choose your stack and copy the exact gobento new command for it.',
};

export default function ConfiguratorPage() {
  return <Configurator catalog={catalog} />;
}
