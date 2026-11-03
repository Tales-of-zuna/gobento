import { expect, it } from 'vitest';
import { editUrl } from './shared';

it('builds the GitHub editor URL of a docs source file on main', () => {
  expect(editUrl('concepts/traits.mdx')).toBe(
    'https://github.com/Tales-of-zuna/gobento/edit/main/website/content/docs/concepts/traits.mdx',
  );
});
