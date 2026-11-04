import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { catalogWith, moduleOf } from '@/lib/test/catalog';
import { Configurator } from './Configurator';

// The mascot stands in for the 3D scene: counting its renders shows whether
// an edit reaches the scene.
const mascotRenders = vi.hoisted(() => ({ count: 0 }));
vi.mock('@/components/mascot/MascotHero', () => ({
  MascotHero: () => {
    mascotRenders.count++;
    return null;
  },
}));

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
  it('asks for an architecture and data access, then shows the command and updates it live', () => {
    render(<Configurator catalog={catalogWith()} />);
    expect(screen.getAllByText(/No command yet. Fix the architecture choice/)[0]).toBeVisible();
    fireEvent.click(screen.getByRole('radio', { name: 'Model-View-Controller' }));
    expect(screen.getAllByText(/No command yet. Fix the data access choice/)[0]).toBeVisible();
    fireEvent.click(screen.getByRole('radio', { name: 'sqlc' }));
    expect(command()).toHaveTextContent(/^gobento new myapp --arch mvc --data sqlc$/);
    fireEvent.change(screen.getByLabelText('Project name'), { target: { value: 'shop' } });
    fireEvent.change(screen.getByLabelText('Module path'), {
      target: { value: 'example.com/shop' },
    });
    expect(command()).toHaveTextContent(
      /^gobento new shop --module example\.com\/shop --arch mvc --data sqlc$/,
    );
    expect(window.location.search).toBe('?name=shop&module=example.com%2Fshop&arch=mvc&data=sqlc');
  });

  it('adds --frontend htmx for htmx and removes it again for None (spec: Choosing a frontend)', () => {
    window.history.replaceState(null, '', '/configurator/?arch=mvc&data=sqlc');
    render(<Configurator catalog={catalogWith()} />);
    const none = screen.getByRole('radio', { name: 'None (API only)' });
    expect(none).toBeChecked();
    fireEvent.click(screen.getByRole('radio', { name: 'htmx + templ' }));
    expect(command()).toHaveTextContent(
      /^gobento new myapp --arch mvc --data sqlc --frontend htmx$/,
    );
    expect(window.location.search).toBe('?arch=mvc&data=sqlc&frontend=htmx');
    fireEvent.click(screen.getByRole('button', { name: /Advanced/ }));
    expect(screen.getByLabelText('github.com/a-h/templ')).toHaveValue('v0.3.1100');
    expect(screen.getByText(/htmx\.org \(htmx \+ templ\)/)).toBeInTheDocument();
    expect(screen.getByText('Vendored')).toBeInTheDocument();
    fireEvent.click(none);
    expect(command()).toHaveTextContent(/^gobento new myapp --arch mvc --data sqlc$/);
    expect(window.location.search).toBe('?arch=mvc&data=sqlc');
    expect(screen.queryByText('Vendored')).toBeNull();
  });

  it('offers Tailwind CSS only with a frontend and adds --css tailwind (spec: CSS framework with a frontend)', () => {
    window.history.replaceState(null, '', '/configurator/?arch=mvc&data=sqlc');
    render(<Configurator catalog={catalogWith()} />);
    const plain = screen.getByRole('radio', { name: 'None (plain CSS)' });
    const tailwind = screen.getByRole('radio', { name: 'Tailwind CSS' });
    expect(plain).toBeChecked();
    expect(tailwind).toBeDisabled();
    expect(tailwind).toHaveAccessibleDescription(
      /requires a frontend that provides frontend-build go or vite, such as htmx \+ templ/,
    );
    fireEvent.click(screen.getByRole('radio', { name: 'htmx + templ' }));
    expect(tailwind).toBeEnabled();
    fireEvent.click(tailwind);
    expect(command()).toHaveTextContent(
      /^gobento new myapp --arch mvc --data sqlc --frontend htmx --css tailwind$/,
    );
    expect(window.location.search).toBe('?arch=mvc&data=sqlc&frontend=htmx&css=tailwind');
    const tray = screen.getByRole('list', { name: 'Bento: chosen parts' });
    expect(within(tray).getAllByRole('listitem')).toHaveLength(6);
    expect(within(tray).getByText('Tailwind CSS')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /Advanced/ }));
    expect(screen.getByText(/tailwindcss \(Tailwind CSS\)/)).toBeInTheDocument();
    expect(screen.getByText('On demand')).toBeInTheDocument();

    // Removing the frontend keeps the CSS choice, which is now incompatible.
    fireEvent.click(screen.getByRole('radio', { name: 'None (API only)' }));
    expect(tailwind).toBeChecked();
    expect(screen.getAllByText(/No command yet. Fix the CSS framework choice/)[0]).toBeVisible();
    expect(window.location.search).toBe('?arch=mvc&data=sqlc&css=tailwind');
    fireEvent.click(plain);
    expect(command()).toHaveTextContent(/^gobento new myapp --arch mvc --data sqlc$/);
    expect(window.location.search).toBe('?arch=mvc&data=sqlc');
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
    window.history.replaceState(
      null,
      '',
      '/configurator/?name=blog&arch=mvc&data=sqlc&framework=fiber',
    );
    render(<Configurator catalog={catalogWith()} />);
    expect(command()).toHaveTextContent(/^gobento new blog --arch mvc --data sqlc$/);
    const notice = screen.getByRole('status');
    expect(within(notice).getByText(/framework "fiber"/)).toBeInTheDocument();
    fireEvent.click(within(notice).getByRole('button', { name: 'Dismiss' }));
    expect(screen.queryByText(/framework "fiber"/)).toBeNull();
  });

  // Version too new for Go. The shipped catalog has no module version that
  // needs a newer Go, so this test adds one: sqlite v1.62.0+ needs Go 1.27.
  it('disables a Go version that a module version is too new for', async () => {
    const strict = catalogWith((data) => {
      moduleOf(data, 'modernc.org/sqlite').constraints.push({
        when: '>=v1.62.0',
        requires: { target: 'go', range: '>=1.27' },
      });
    });
    window.history.replaceState(null, '', '/configurator/?arch=mvc');
    render(<Configurator catalog={strict} />);
    fireEvent.click(screen.getByRole('radio', { name: 'sqlc' }));
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

  it('leaves the packing stage and its 3D scene alone while the visitor types', () => {
    render(<Configurator catalog={catalogWith()} />);
    fireEvent.click(screen.getByRole('radio', { name: 'Model-View-Controller' }));
    const before = mascotRenders.count;
    fireEvent.change(screen.getByLabelText('Project name'), { target: { value: 'shop' } });
    fireEvent.change(screen.getByLabelText('Module path'), {
      target: { value: 'example.com/shop' },
    });
    expect(mascotRenders.count).toBe(before);
    fireEvent.click(screen.getByRole('radio', { name: 'sqlc' }));
    expect(mascotRenders.count).toBeGreaterThan(before);
    expect(command()).toHaveTextContent(/^gobento new shop --module example\.com\/shop/);
  });
});
