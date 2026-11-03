import { describe, expect, it } from 'vitest';
import { buildCommand, inputErrors, type ConfigState } from './build';
import { defaultState } from './url-state';
import { addPostgres, catalogWith } from '@/lib/test/catalog';

const fixture = catalogWith();
const NAME_RULE = fixture.rules.projectName.message;
const MODULE_RULE = fixture.rules.modulePath.message;

const state = (overrides: Partial<ConfigState> = {}): ConfigState => ({
  ...defaultState(fixture),
  ...overrides,
});

describe('buildCommand', () => {
  const cases: { name: string; state: ConfigState; command: string }[] = [
    { name: 'defaults', state: state(), command: 'gobento new myapp' },
    {
      name: 'changing a choice (spec: Changing a choice)',
      state: state({ name: 'shop', module: 'example.com/shop' }),
      command: 'gobento new shop --module example.com/shop',
    },
    {
      name: 'module equal to the name is the default',
      state: state({ name: 'shop', module: 'shop' }),
      command: 'gobento new shop',
    },
    {
      name: 'default Go version written another way',
      state: state({ go: '1.27.0' }),
      command: 'gobento new myapp',
    },
    {
      name: 'non-default Go version',
      state: state({ go: '1.26' }),
      command: 'gobento new myapp --go 1.26',
    },
    {
      name: 'pin equal to the default',
      state: state({ pins: { 'modernc.org/sqlite': 'v1.61.1' } }),
      command: 'gobento new myapp',
    },
    {
      name: 'fixed flag order',
      state: state({
        name: 'shop',
        module: 'example.com/shop',
        go: '1.26',
        pins: { 'modernc.org/sqlite': 'v1.62.0' },
      }),
      command:
        'gobento new shop --module example.com/shop --go 1.26 --pin modernc.org/sqlite=v1.62.0',
    },
  ];

  it.each(cases)('$name', ({ state: input, command }) => {
    expect(buildCommand(fixture, input)).toEqual({ ok: true, command });
  });

  it('includes the flag of a kind with several options, after --module', () => {
    const catalog = catalogWith(addPostgres);
    const input: ConfigState = {
      ...defaultState(catalog),
      name: 'shop',
      module: 'example.com/shop',
      components: { ...defaultState(catalog).components, database: 'sqlite' },
      go: '1.26',
    };
    expect(buildCommand(catalog, input)).toEqual({
      ok: true,
      command: 'gobento new shop --module example.com/shop --db sqlite --go 1.26',
    });
  });

  it('requires a choice for a kind with several options', () => {
    const catalog = catalogWith(addPostgres);
    expect(buildCommand(catalog, defaultState(catalog))).toEqual({
      ok: false,
      errors: [{ field: 'database', message: 'Choose a database: sqlite, postgres.' }],
    });
  });

  it('offers no command for an invalid project name (spec: Invalid project name)', () => {
    expect(buildCommand(fixture, state({ name: 'My App' }))).toEqual({
      ok: false,
      errors: [{ field: 'name', message: NAME_RULE }],
    });
  });

  it('offers no command for an incompatible selection', () => {
    const result = buildCommand(fixture, state({ pins: { 'modernc.org/sqlite': 'v2.0.0' } }));
    expect(result.ok).toBe(false);
  });

  it('reports every problem at once', () => {
    const result = buildCommand(
      fixture,
      state({ name: 'My App', module: 'Not A Path', go: '1.30' }),
    );
    expect(result.ok ? [] : result.errors.map((e) => e.field)).toEqual(['name', 'module', 'go']);
  });
});

describe('inputErrors', () => {
  const names: [string, boolean][] = [
    ['shop', true],
    ['my-app_2', true],
    ['a', true],
    ['My App', false],
    ['Shop', false],
    ['2shop', false],
    ['-shop', false],
    ['shop.v2', false],
    ['', false],
  ];

  it.each(names)('project name %j valid: %s', (name, valid) => {
    const errors = inputErrors(fixture, state({ name }));
    expect(errors).toEqual(valid ? [] : [{ field: 'name', message: NAME_RULE }]);
  });

  const modules: [string, boolean][] = [
    ['', true], // empty means the project name
    ['shop', true],
    ['example.com/shop', true],
    ['github.com/Tales-of-zuna/gobento', true],
    ['example.com/shop/v2', true],
    ['gopkg.in/yaml.v3', true],
    ['Not A Path', false],
    ['/example.com/shop', false],
    ['example.com/shop/', false],
    ['example.com//shop', false],
    ['example.com/../shop', false],
    ['example.com/./shop', false],
    ['example.com./shop', false],
    ['example.com/shop?x', false],
  ];

  it.each(modules)('module path %j valid: %s', (module, valid) => {
    const errors = inputErrors(fixture, state({ module }));
    expect(errors).toEqual(valid ? [] : [{ field: 'module', message: MODULE_RULE }]);
  });
});
