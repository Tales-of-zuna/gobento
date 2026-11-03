import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { PRERELEASE_TEXT, PrereleaseNotice } from './PrereleaseNotice';
import { SiteFooter } from './SiteFooter';
import { SiteNav } from './SiteNav';

describe('PrereleaseNotice', () => {
  it('shows the notice before the first release', () => {
    render(<PrereleaseNotice released={false} />);
    expect(screen.getByText(PRERELEASE_TEXT)).toHaveAttribute('data-prerelease-notice');
  });

  it('renders nothing once a version is released', () => {
    const { container } = render(<PrereleaseNotice released />);
    expect(container).toBeEmptyDOMElement();
  });
});

describe('SiteNav', () => {
  it('marks the current section in the bar and the mobile menu', () => {
    const { container } = render(<SiteNav active="configurator" />);
    for (const name of ['Main', 'Mobile']) {
      const nav = within(container.querySelector<HTMLElement>(`nav[aria-label="${name}"]`)!);
      expect(nav.getByRole('link', { name: 'Configurator', hidden: true })).toHaveAttribute(
        'aria-current',
        'page',
      );
      expect(nav.getByRole('link', { name: 'Docs', hidden: true })).not.toHaveAttribute(
        'aria-current',
      );
    }
  });

  it('marks nothing on the home page', () => {
    render(<SiteNav variant="overlay" />);
    const main = within(screen.getByRole('navigation', { name: 'Main' }));
    expect(main.queryAllByRole('link', { current: 'page' })).toHaveLength(0);
  });

  it('opens the mobile menu and closes it with Escape', () => {
    render(<SiteNav active="docs" />);
    const button = screen.getByRole('button', { name: 'Menu' });
    fireEvent.click(button);
    expect(button).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByRole('navigation', { name: 'Mobile' })).toBeVisible();
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(button).toHaveAttribute('aria-expanded', 'false');
    expect(button).toHaveFocus();
  });
});

describe('SiteFooter', () => {
  it('carries the mascot attribution and non-affiliation', () => {
    render(<SiteFooter flush />);
    const footer = screen.getByRole('contentinfo');
    expect(footer).toHaveTextContent(
      'The gobento mascot is an original character inspired by the Go gopher, designed by Renée French and licensed under CC BY 4.0. gobento is not affiliated with or endorsed by the Go project.',
    );
  });
});
