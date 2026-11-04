import { describe, expect, it } from 'vitest';
import { ROADMAP as CONFIGURATOR_ROADMAP } from '@/components/configurator/roadmap';
import { ROADMAP as HOME_ROADMAP, stackCells } from '@/lib/home/menu';
import { catalog } from './data';

/** Lower-cased names and titles of the components the real catalog ships. */
const shipped = new Set(
  catalog.components.flatMap((c) => [`${c.kind}/${c.name}`, `${c.kind}/${c.title}`.toLowerCase()]),
);

describe('the real catalog', () => {
  it('ships one or more components of every kind', () => {
    for (const kind of catalog.kinds) {
      expect(catalog.components.some((c) => c.kind === kind.id)).toBe(true);
    }
  });

  it('is not listed on the configurator roadmap', () => {
    for (const [kind, planned] of Object.entries(CONFIGURATOR_ROADMAP)) {
      for (const p of planned) {
        expect(
          shipped.has(`${kind}/${p.name}`),
          `${kind}/${p.name} ships; remove it from the roadmap`,
        ).toBe(false);
      }
    }
  });

  it('lists the Vite frontends as shipped frontends on the home page', () => {
    const frontend = stackCells(catalog).find((cell) => cell.kind === 'frontend')!;
    expect(frontend.options.map((o) => o.name)).toEqual(
      expect.arrayContaining(['htmx + templ', 'React', 'Vue', 'Svelte']),
    );
    for (const name of ['React', 'Vue', 'Svelte']) expect(frontend.roadmap).not.toContain(name);
  });

  it('is not listed on the home page roadmap', () => {
    for (const item of HOME_ROADMAP) {
      const id = `${item.kind}/${item.name}`.toLowerCase();
      expect(shipped.has(id), `${id} ships; remove it from the roadmap`).toBe(false);
    }
  });
});
