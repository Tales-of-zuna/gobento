import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { catalogWith } from '@/lib/test/catalog';
import { Configurator } from './Configurator';

beforeEach(() => {
  window.history.replaceState(null, '', '/configurator/');
  vi.stubGlobal(
    'matchMedia',
    (query: string) =>
      ({
        matches: query.includes('reduce'),
        addEventListener: () => {},
        removeEventListener: () => {},
      }) as unknown as MediaQueryList,
  );
});

const command = () => screen.getAllByText((_, el) => el?.hasAttribute('data-command') ?? false)[0];

describe('Configurator', () => {
  it('shows the default command and updates it live', () => {
    render(<Configurator catalog={catalogWith()} />);
    expect(command()).toHaveTextContent(/^gobento new myapp$/);
    fireEvent.change(screen.getByLabelText('Project name'), { target: { value: 'shop' } });
    fireEvent.change(screen.getByLabelText('Module path'), {
      target: { value: 'example.com/shop' },
    });
    expect(command()).toHaveTextContent(/^gobento new shop --module example\.com\/shop$/);
    expect(window.location.search).toBe('?name=shop&module=example.com%2Fshop');
  });

  it('explains an invalid name next to the field and offers no command', () => {
    render(<Configurator catalog={catalogWith()} />);
    const name = screen.getByLabelText('Project name');
    fireEvent.change(name, { target: { value: 'My App' } });
    expect(name).toHaveAttribute('aria-invalid', 'true');
    expect(name).toHaveAccessibleDescription(/Start with a lowercase letter/);
    expect(screen.queryByRole('button', { name: 'Copy command' })).toBeNull();
    expect(screen.getAllByText(/No command yet. Fix the project name/)[0]).toBeVisible();
  });

  it('restores a shared link and reports unknown values', () => {
    window.history.replaceState(null, '', '/configurator/?name=blog&framework=fiber');
    render(<Configurator catalog={catalogWith()} />);
    expect(command()).toHaveTextContent(/^gobento new blog$/);
    const notice = screen.getByRole('status');
    expect(within(notice).getByText(/framework "fiber"/)).toBeInTheDocument();
    fireEvent.click(within(notice).getByRole('button', { name: 'Dismiss' }));
    expect(screen.queryByText(/framework "fiber"/)).toBeNull();
  });

  // Version too new for Go. The shipped catalog has no module version that
  // needs a newer Go, so this test adds one: sqlite v1.62.0+ needs Go 1.27.
  it('disables a Go version that a module version is too new for', async () => {
    const strict = catalogWith((data) => {
      data.modules[0]!.constraints.push({
        when: '>=v1.62.0',
        requires: { target: 'go', range: '>=1.27' },
      });
    });
    render(<Configurator catalog={strict} />);
    fireEvent.click(screen.getByRole('button', { name: /Advanced/ }));
    await act(async () => {
      fireEvent.change(screen.getByLabelText('modernc.org/sqlite'), {
        target: { value: 'v1.62.0' },
      });
    });
    const go126 = screen.getByRole('radio', { name: 'Go 1.26' });
    expect(go126).toBeDisabled();
    expect(go126).toHaveAccessibleDescription(/needs Go >=1\.27, but Go 1\.26 is selected/);
    expect(command()).toHaveTextContent(/--pin modernc\.org\/sqlite=v1\.62\.0/);
  });

  it('shows the pre-release notice from the release status it is given', () => {
    const { unmount } = render(<Configurator catalog={catalogWith()} released={false} />);
    expect(screen.getByText('Early development · not released yet')).toBeInTheDocument();
    unmount();
    render(<Configurator catalog={catalogWith()} released />);
    expect(screen.queryByText('Early development · not released yet')).toBeNull();
  });
});
