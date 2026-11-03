import { describe, expect, it } from 'vitest';
import { DOCS_SECTIONS, sectionClass, sectionOfSlug, sectionOfUrl } from '@/lib/docs-sections';

describe('docs sections', () => {
  it('gives every section its own color', () => {
    const colors = DOCS_SECTIONS.map((section) => section.color);
    expect(new Set(colors).size).toBe(colors.length);
  });

  it('finds the section of a slug, and none for the introduction', () => {
    expect(sectionOfSlug(['concepts', 'traits'])?.title).toBe('Concepts');
    expect(sectionOfSlug(['components'])?.color).toBe('matcha');
    expect(sectionOfSlug(undefined)).toBeUndefined();
    expect(sectionOfSlug([])).toBeUndefined();
  });

  it('finds the section of a URL with or without the base path', () => {
    expect(sectionOfUrl('/gobento/docs/cli/new/')?.id).toBe('cli');
    expect(sectionOfUrl('/docs/contributing/adding-a-component#steps')?.color).toBe('sakura');
    expect(sectionOfUrl('/docs')).toBeUndefined();
    expect(sectionOfUrl('/gobento/configurator/')).toBeUndefined();
  });

  it('maps a section to its color class', () => {
    expect(sectionClass(sectionOfSlug(['getting-started']))).toBe('g-tamago');
    expect(sectionClass(undefined)).toBe('g-none');
  });
});
