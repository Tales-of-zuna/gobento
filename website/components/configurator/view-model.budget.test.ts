/**
 * A cost budget for the configurator's view model on the real catalog. Every
 * render of the page recomputes it, so it must stay cheap as the catalog
 * grows: it measured about 0.15 ms per selection once warm, 1.2 ms cold
 * (fix-configurator-latency), and the budget leaves room for slow CI runners while still catching a
 * blow-up such as compatibility checks that scale with every combination.
 */
import { describe, expect, it } from 'vitest';
import { selectionProblems, type Selection } from '@/lib/catalog/compat';
import { catalog } from '@/lib/catalog/data';
import { KIND_IDS } from '@/lib/catalog/kinds';
import { buildCommand, type ConfigState } from '@/lib/command/build';
import { defaultState, encodeState } from '@/lib/command/url-state';
import {
  errorsByField,
  fileTree,
  goOptions,
  kindGroups,
  moduleFields,
  trayCells,
} from './view-model';

/** Average budget per selection, in milliseconds. */
const BUDGET_MS = 10;

/** Every selection of one component per kind that the catalog supports. */
function supportedSelections(): ConfigState[] {
  const base = defaultState(catalog);
  let partial: Selection['components'][] = [{}];
  for (const kind of KIND_IDS) {
    const names = catalog.components.filter((c) => c.kind === kind).map((c) => c.name);
    partial = partial.flatMap((components) =>
      names.map((name) => ({ ...components, [kind]: name })),
    );
  }
  return partial
    .map((components) => ({ ...base, name: 'shop', module: 'example.com/shop', components }))
    .filter((state) => selectionProblems(catalog, state).length === 0);
}

/** Everything the configurator derives from the state on one render. */
function viewModel(state: ConfigState) {
  const result = buildCommand(catalog, state);
  const errors = errorsByField(result);
  return [
    kindGroups(catalog, state, errors),
    goOptions(catalog, state),
    moduleFields(catalog, state, errors),
    trayCells(catalog, state),
    fileTree(catalog, state, !errors.has('name')),
    encodeState(catalog, state),
  ];
}

describe('view model cost', () => {
  const selections = supportedSelections();

  it('covers every supported selection of the real catalog', () => {
    expect(selections.length).toBeGreaterThan(1);
    for (const state of selections) expect(buildCommand(catalog, state).ok).toBe(true);
  });

  it(`stays under ${BUDGET_MS} ms per selection`, () => {
    selections.forEach(viewModel); // warm up the JIT
    const rounds = 3;
    const start = performance.now();
    for (let i = 0; i < rounds; i++) selections.forEach(viewModel);
    const perSelection = (performance.now() - start) / (rounds * selections.length);
    expect(perSelection).toBeLessThan(BUDGET_MS);
  });
});
