import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Button } from './button';

describe('Button', () => {
  it('renders an accessible button with the default variant classes', () => {
    render(<Button>Copy</Button>);
    const button = screen.getByRole('button', { name: 'Copy' });
    expect(button).toHaveAttribute('data-slot', 'button');
    expect(button).toHaveClass('bg-primary');
  });
});
