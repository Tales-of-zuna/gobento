import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Item, Stagger } from './motion';

/**
 * Progressive enhancement: reveals render their content as plain, visible
 * markup (no inline hidden state), so it shows without JavaScript and before
 * hydration. The animations live in CSS and only run once scripts mark the page.
 */
describe('Stagger and Item', () => {
  it('render visible content with no inline hidden state', () => {
    render(
      <Stagger>
        <Item>
          <h1>Title</h1>
        </Item>
        <Item kind="pop">
          <p>Body</p>
        </Item>
      </Stagger>,
    );
    for (const text of ['Title', 'Body']) {
      const item = screen.getByText(text).closest('[data-reveal-item]') as HTMLElement;
      expect(item).not.toBeNull();
      expect(item.getAttribute('style')).toBeNull();
      expect(item.parentElement).not.toHaveAttribute('data-reveal');
    }
  });

  it('renders list items for trays', () => {
    render(
      <Stagger inView>
        <ul>
          <Item as="li">one</Item>
        </ul>
      </Stagger>,
    );
    expect(screen.getByRole('listitem')).toHaveTextContent('one');
    expect(screen.getByRole('listitem')).toBeVisible();
  });
});
