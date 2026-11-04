/**
 * Components planned for later releases, shown as disabled "Coming soon"
 * cards so visitors can see where gobento is heading. They are never
 * selectable and never reach the command. As soon as the catalog contains a
 * component of the same kind and name, the catalog entry replaces the card.
 */
import type { KindId } from '@/lib/catalog/schema';

/** A planned component. */
export interface PlannedComponent {
  readonly name: string;
  readonly title: string;
}

export const ROADMAP: Readonly<Record<KindId, readonly PlannedComponent[]>> = {
  architecture: [],
  framework: [],
  database: [],
  data: [],
  frontend: [
    { name: 'react', title: 'React via Vite' },
    { name: 'vue', title: 'Vue via Vite' },
    { name: 'svelte', title: 'Svelte via Vite' },
  ],
  css: [],
};
