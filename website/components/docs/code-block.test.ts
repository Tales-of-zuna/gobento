import { describe, expect, it } from 'vitest';
import { codeBlockLabel } from './code-block';

describe('codeBlockLabel', () => {
  it('numbers each block and adds its title or language', () => {
    expect(codeBlockLabel(1, 'components/versions.yaml', 'yaml')).toBe(
      'Code example 1: components/versions.yaml',
    );
    expect(codeBlockLabel(2, undefined, 'sh')).toBe('Code example 2 (shell)');
    expect(codeBlockLabel(3, undefined, 'toml')).toBe('Code example 3 (toml)');
    expect(codeBlockLabel(4)).toBe('Code example 4');
  });
});
