import { describe, expect, it } from 'vitest';
import { catalogWith, moduleOf } from '@/lib/test/catalog';
import { beltPlates, ROADMAP, shortTitle, stackCells, versionShowcase } from './menu';

describe('shortTitle', () => {
  it('keeps short titles and abbreviates long ones to initials', () => {
    expect(shortTitle({ title: 'net/http' })).toBe('net/http');
    expect(shortTitle({ title: 'Model-View-Controller' })).toBe('MVC');
  });
});

describe('beltPlates', () => {
  it('lists the shipping components in kind order, then the roadmap', () => {
    const plates = beltPlates(catalogWith());
    expect(plates.filter((p) => p.status === 'ready').map((p) => p.name)).toEqual([
      'MVC',
      'net/http',
      'SQLite',
      'sqlc',
    ]);
    expect(plates.filter((p) => p.status === 'roadmap')).toHaveLength(ROADMAP.length);
  });

  it('follows the catalog when a component is added', () => {
    const catalog = catalogWith((data) => {
      data.components.push({
        id: 'framework/echo',
        kind: 'framework',
        name: 'echo',
        title: 'Echo',
        description: 'A minimalist web framework.',
        provides: {},
        requires: {},
      });
    });
    expect(beltPlates(catalog, []).map((p) => p.name)).toContain('Echo');
  });
});

describe('stackCells', () => {
  it('gives one compartment per kind with its options and roadmap', () => {
    const cells = stackCells(catalogWith());
    expect(cells.map((c) => c.title)).toEqual([
      'Architecture',
      'HTTP framework',
      'Database',
      'Data access',
    ]);
    expect(cells[1]!.options.map((o) => o.name)).toEqual(['net/http']);
    expect(cells[1]!.roadmap).toEqual(['gin', 'chi']);
  });
});

describe('versionShowcase', () => {
  it('derives the rule and the ruled-out Go version from the catalog', () => {
    expect(versionShowcase(catalogWith())).toEqual({
      module: 'modernc.org/sqlite',
      version: 'v1.61.1',
      role: 'library',
      goRange: '>=1.26',
      goMinimum: '1.26',
      goDefault: '1.27',
      goDisabled: '1.25',
      reason: 'modernc.org/sqlite v1.61.1 needs Go 1.26 or newer',
    });
  });

  it('tracks a changed default version', () => {
    const catalog = catalogWith((data) => {
      moduleOf(data, 'modernc.org/sqlite').default = 'v1.62.0';
    });
    expect(versionShowcase(catalog)?.reason).toBe(
      'modernc.org/sqlite v1.62.0 needs Go 1.26 or newer',
    );
  });

  it('returns undefined when no library constrains Go', () => {
    const catalog = catalogWith((data) => {
      for (const mod of data.modules) mod.constraints = [];
    });
    expect(versionShowcase(catalog)).toBeUndefined();
  });
});
