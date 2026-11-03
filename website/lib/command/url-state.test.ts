import { describe, expect, it } from 'vitest';
import { buildCommand, type ConfigState } from './build';
import { decodeState, defaultState, encodeState } from './url-state';
import { addPostgres, catalogWith } from '@/lib/test/catalog';

const fixture = catalogWith();

describe('defaultState (spec: Defaults)', () => {
  it('preselects single-option kinds and the default Go version', () => {
    expect(defaultState(fixture)).toEqual({
      name: 'myapp',
      module: '',
      components: { architecture: 'mvc', framework: 'nethttp', database: 'sqlite', data: 'sqlc' },
      go: '1.27',
      pins: {},
    });
  });

  it('leaves a kind with several options unchosen', () => {
    expect(defaultState(catalogWith(addPostgres)).components.database).toBeUndefined();
  });
});

describe('encodeState', () => {
  it('encodes nothing for the defaults', () => {
    expect(encodeState(fixture, defaultState(fixture))).toBe('');
  });

  it('encodes changed values with the CLI flag names', () => {
    const state: ConfigState = {
      ...defaultState(fixture),
      name: 'shop',
      module: 'example.com/shop',
      go: '1.26',
      pins: { 'modernc.org/sqlite': 'v1.62.0' },
    };
    expect(encodeState(fixture, state)).toBe(
      'name=shop&module=example.com%2Fshop&go=1.26&pin=modernc.org%2Fsqlite%3Dv1.62.0',
    );
  });
});

describe('decodeState', () => {
  it('restores the same choices and command from a shared link (spec: Share a link)', () => {
    const catalog = catalogWith(addPostgres);
    const chosen: ConfigState = {
      name: 'shop',
      module: 'example.com/shop',
      components: { ...defaultState(catalog).components, database: 'sqlite' },
      go: '1.26',
      pins: { 'modernc.org/sqlite': 'v1.62.0' },
    };
    const link = `https://tales-of-zuna.github.io/gobento/configurator/?${encodeState(catalog, chosen)}`;

    const { state, notices } = decodeState(catalog, new URL(link).search);

    expect(notices).toEqual([]);
    expect(state).toEqual(chosen);
    expect(buildCommand(catalog, state)).toEqual(buildCommand(catalog, chosen));
  });

  it('accepts a leading "?"', () => {
    expect(decodeState(fixture, '?name=shop').state.name).toBe('shop');
  });

  const unknown: {
    name: string;
    query: string;
    notice: string;
    check: (s: ConfigState) => void;
  }[] = [
    {
      name: 'component',
      query: 'framework=express',
      notice:
        'The link names http framework "express", which is not available; the default is used.',
      check: (s) => expect(s.components.framework).toBe('nethttp'),
    },
    {
      name: 'Go version',
      query: 'go=1.20',
      notice: 'The link names Go 1.20, which is not supported; Go 1.27 is used.',
      check: (s) => expect(s.go).toBe('1.27'),
    },
    {
      name: 'module version',
      query: 'pin=modernc.org/sqlite=v1.50.0',
      notice:
        'The link pins modernc.org/sqlite v1.50.0, which is not supported; the default v1.61.1 is used.',
      check: (s) => expect(s.pins).toEqual({}),
    },
    {
      name: 'module',
      query: 'pin=example.com/gone=v1.0.0',
      notice: 'The link pins "example.com/gone", which cannot be pinned; it is ignored.',
      check: (s) => expect(s.pins).toEqual({}),
    },
    {
      name: 'tool module',
      query: 'pin=github.com/sqlc-dev/sqlc=v1.31.1',
      notice: 'The link pins "github.com/sqlc-dev/sqlc", which cannot be pinned; it is ignored.',
      check: (s) => expect(s.pins).toEqual({}),
    },
    {
      name: 'malformed pin',
      query: 'pin=nonsense',
      notice: 'The link pins "nonsense", which cannot be pinned; it is ignored.',
      check: (s) => expect(s.pins).toEqual({}),
    },
  ];

  it.each(unknown)(
    'falls back to the default for an unknown $name, with a notice (spec: Unknown value in a shared link)',
    ({ query, notice, check }) => {
      const { state, notices } = decodeState(fixture, `name=shop&${query}`);
      expect(notices).toEqual([notice]);
      expect(state.name).toBe('shop');
      check(state);
    },
  );

  it('keeps an invalid name so the form can explain it', () => {
    const { state, notices } = decodeState(fixture, 'name=My+App');
    expect(notices).toEqual([]);
    expect(buildCommand(fixture, state).ok).toBe(false);
  });
});
