import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Callout } from '@/components/docs/callout';

describe('Callout', () => {
  it('is a note labelled by its title', () => {
    render(<Callout title="Planned">Not released yet.</Callout>);
    const note = screen.getByRole('note');
    expect(note).toHaveTextContent('Planned');
    expect(note).toHaveTextContent('Not released yet.');
  });

  it.each([
    ['info', 'Note', 'docs-co-note'],
    ['tip', 'Tip', 'docs-co-tip'],
    ['idea', 'Tip', 'docs-co-tip'],
    ['warn', 'Warning', 'docs-co-warn'],
    ['error', 'Warning', 'docs-co-warn'],
  ] as const)('maps type %s to the %s compartment', (type, label, className) => {
    render(<Callout type={type}>Body</Callout>);
    const note = screen.getByRole('note');
    expect(note).toHaveClass(className);
    expect(note).toHaveTextContent(label);
  });
});
