import { describe, expect, it } from 'vitest';
import { buildCommand, inputErrors, type ConfigState } from './build';
import { defaultState } from './url-state';
import { addPostgres, catalogWith } from '@/lib/test/catalog';

const fixture = catalogWith();
const NAME_RULE = fixture.rules.projectName.message;
const MODULE_RULE = fixture.rules.modulePath.message;
const RESERVED_RULE = fixture.rules.projectName.reservedMessage;

/** The default state with each kind that has several options (architecture, data) chosen. */
const state = (overrides: Partial<ConfigState> = {}): ConfigState => ({
  ...defaultState(fixture),
  components: { ...defaultState(fixture).components, architecture: 'mvc', data: 'sqlc' },
  ...overrides,
});

describe('buildCommand', () => {
  const cases: { name: string; state: ConfigState; command: string }[] = [
    { name: 'defaults', state: state(), command: 'gobento new myapp --arch mvc --data sqlc' },
    {
      name: 'changing a choice (spec: Changing a choice)',
      state: state({ name: 'shop', module: 'example.com/shop' }),
      command: 'gobento new shop --module example.com/shop --arch mvc --data sqlc',
    },
    {
      name: 'module equal to the name is the default',
      state: state({ name: 'shop', module: 'shop' }),
      command: 'gobento new shop --arch mvc --data sqlc',
    },
    {
      name: 'default Go version written another way',
      state: state({ go: '1.27.0' }),
      command: 'gobento new myapp --arch mvc --data sqlc',
    },
    {
      name: 'non-default Go version',
      state: state({ go: '1.26' }),
      command: 'gobento new myapp --arch mvc --data sqlc --go 1.26',
    },
    {
      name: 'pin equal to the default',
      state: state({ pins: { 'modernc.org/sqlite': 'v1.61.1' } }),
      command: 'gobento new myapp --arch mvc --data sqlc',
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
        'gobento new shop --module example.com/shop --arch mvc --data sqlc --go 1.26 --pin modernc.org/sqlite=v1.62.0',
    },
  ];

  it.each([
    {
      name: 'no frontend omits the flag (spec: Choosing a frontend)',
      state: state(),
      command: 'gobento new myapp --arch mvc --data sqlc',
    },
    {
      name: 'a frontend adds its flag after the required kinds, even as the only option',
      state: state({
        components: {
          framework: 'nethttp',
          database: 'sqlite',
          architecture: 'mvc',
          data: 'sqlc',
          frontend: 'htmx',
        },
        go: '1.26',
      }),
      command: 'gobento new myapp --arch mvc --data sqlc --frontend htmx --go 1.26',
    },
    {
      name: 'a CSS framework adds --css after --frontend (spec: CSS framework with a frontend)',
      state: state({
        components: {
          css: 'tailwind',
          architecture: 'mvc',
          data: 'sqlc',
          frontend: 'htmx',
          framework: 'nethttp',
          database: 'sqlite',
        },
      }),
      command: 'gobento new myapp --arch mvc --data sqlc --frontend htmx --css tailwind',
    },
  ])('$name', ({ state: input, command }) => {
    expect(buildCommand(fixture, input)).toEqual({ ok: true, command });
  });

  it('never asks for an optional kind to be chosen', () => {
    const result = buildCommand(fixture, state());
    expect(result.ok).toBe(true);
  });

  it('offers no command for a CSS framework without a frontend (spec: CSS framework needs a frontend)', () => {
    const result = buildCommand(
      fixture,
      state({ components: { ...state().components, css: 'tailwind' } }),
    );
    expect(result).toEqual({
      ok: false,
      errors: [
        {
          field: 'css',
          message:
            'Tailwind CSS requires a frontend that provides frontend-build go or vite, such as htmx + templ.',
          involves: ['css/tailwind'],
        },
      ],
    });
  });

  it.each(cases)('$name', ({ state: input, command }) => {
    expect(buildCommand(fixture, input)).toEqual({ ok: true, command });
  });

  it('includes the flag of each kind with several options, after --module', () => {
    const catalog = catalogWith(addPostgres);
    const input: ConfigState = {
      ...defaultState(catalog),
      name: 'shop',
      module: 'example.com/shop',
      components: {
        ...defaultState(catalog).components,
        architecture: 'mvc',
        database: 'sqlite',
        data: 'sqlc',
      },
      go: '1.26',
    };
    expect(buildCommand(catalog, input)).toEqual({
      ok: true,
      command:
        'gobento new shop --module example.com/shop --arch mvc --db sqlite --data sqlc --go 1.26',
    });
  });

  it('requires a choice for each kind with several options (spec: Defaults)', () => {
    expect(buildCommand(fixture, defaultState(fixture))).toEqual({
      ok: false,
      errors: [
        { field: 'architecture', message: 'Choose one of: mvc, vertical-slice.' },
        { field: 'data', message: 'Choose one of: gorm, sqlc.' },
      ],
    });
    const catalog = catalogWith(addPostgres);
    expect(buildCommand(catalog, defaultState(catalog))).toEqual({
      ok: false,
      errors: [
        { field: 'architecture', message: 'Choose one of: mvc, vertical-slice.' },
        { field: 'database', message: 'Choose one of: sqlite, postgres.' },
        { field: 'data', message: 'Choose one of: gorm, sqlc.' },
      ],
    });
  });

  it('requires a data access, which has several options', () => {
    const archOnly = state({
      components: { ...defaultState(fixture).components, architecture: 'mvc' },
    });
    expect(buildCommand(fixture, archOnly)).toEqual({
      ok: false,
      errors: [{ field: 'data', message: 'Choose one of: gorm, sqlc.' }],
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

  it.each(['con', 'nul', 'com1', 'lpt9'])('reserved project name %j is rejected', (name) => {
    expect(inputErrors(fixture, state({ name }))).toEqual([
      { field: 'name', message: RESERVED_RULE },
    ]);
  });

  it.each(['com', 'com0', 'console'])('project name %j is not reserved', (name) => {
    expect(inputErrors(fixture, state({ name }))).toEqual([]);
  });

  const modules: [string, boolean][] = [
    ['', true], // empty means the project name
    ['shop', true],
    ['example.com/shop', true],
    ['github.com/Tales-of-zuna/gobento', true],
    ['example.com/shop/v2', true],
    ['gopkg.in/yaml.v3', true],
    ['example.com/a+b', true],
    ['-shop', false],
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
