import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { parseReleaseManifest } from '@/lib/release';
import {
  closingLead,
  HeroInstallNote,
  PLANNED_CLOSING,
  PLANNED_NOTE,
  RELEASED_CLOSING,
} from './ReleaseCopy';

const unreleased = parseReleaseManifest('{".": "0.0.0"}');
const released = parseReleaseManifest('{".": "0.1.0"}');

describe('HeroInstallNote', () => {
  it('No release yet: says there is nothing to install', () => {
    const { container } = render(<HeroInstallNote release={unreleased} />);
    expect(screen.getByText(PLANNED_NOTE)).toBeInTheDocument();
    expect(container.querySelector('[data-install-note]')).toBeNull();
    expect(screen.queryByRole('link')).toBeNull();
  });

  it('Released: names the version and links to the install guide and the binaries', () => {
    render(<HeroInstallNote release={released} />);
    expect(screen.queryByText(PLANNED_NOTE)).toBeNull();
    const note = screen.getByText(/gobento v0\.1\.0 is out\./);
    expect(note).toHaveAttribute('data-install-note');
    expect(screen.getByRole('link', { name: 'Install it' })).toHaveAttribute(
      'href',
      expect.stringMatching(/^\/docs\/getting-started\/installation\/?$/),
    );
    expect(screen.getByRole('link', { name: 'prebuilt binary' })).toHaveAttribute(
      'href',
      'https://github.com/Tales-of-zuna/gobento/releases',
    );
  });
});

describe('closingLead', () => {
  it('promises the CLI before the first release', () => {
    expect(closingLead(unreleased)).toBe(PLANNED_CLOSING);
  });

  it('points at the configurator and the CLI once released', () => {
    expect(closingLead(released)).toBe(RELEASED_CLOSING);
  });
});
