/**
 * The Components reference (design D5): every component of the catalog with
 * its traits and module versions, rendered at build time so that no docs
 * content contains a version number.
 */
import { catalog as builtCatalog } from '@/lib/catalog/data';
import { KIND_IDS, type Catalog, type Component, type Module } from '@/lib/catalog/schema';

/** Renders the reference for `catalog`. */
export function CatalogReference({ catalog }: { catalog: Catalog }) {
  return (
    <>
      <h2 id="go">Go</h2>
      <p>
        The Go versions a generated project may use in its <code>go</code> directive.
      </p>
      <div className="docs-table" tabIndex={0} role="region" aria-label="Go versions table">
        <table aria-label="Go versions">
          <thead>
            <tr>
              <th scope="col">Supported range</th>
              <th scope="col">Default</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>
                <code>{catalog.go.range}</code>
              </td>
              <td>
                <code>{catalog.go.default}</code>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      {KIND_IDS.map((kindId) => {
        const kind = catalog.kinds.find((k) => k.id === kindId);
        const components = catalog.components.filter((c) => c.kind === kindId);
        if (!kind) return null;
        return (
          <section
            key={kindId}
            aria-labelledby={`kind-${kindId}`}
            className="docs-kind"
            data-kind={kindId}
          >
            <h2 id={`kind-${kindId}`}>
              <i className="docs-mk" aria-hidden="true" />
              {kind.title}
            </h2>
            <p>{kind.description}</p>
            {components.map((component) => (
              <ComponentEntry
                key={component.id}
                component={component}
                kindTitle={kind.title}
                modules={catalog.modules.filter((m) => m.components.includes(component.id))}
              />
            ))}
          </section>
        );
      })}
    </>
  );
}

function ComponentEntry({
  component,
  kindTitle,
  modules,
}: {
  component: Component;
  kindTitle: string;
  modules: Module[];
}) {
  const provides = Object.entries(component.provides);
  const requires = Object.entries(component.requires);
  const headingId = `component-${component.kind}-${component.name}`;
  return (
    <article aria-labelledby={headingId} className="docs-entry">
      <h3 id={headingId}>
        {component.title} (<code>{component.name}</code>)
      </h3>
      <p>{component.description}</p>
      <dl>
        <dt>Kind</dt>
        <dd>{kindTitle}</dd>
        <dt>Provides</dt>
        <dd>
          {provides.length === 0
            ? 'No traits'
            : provides.map(([trait, value], i) => (
                <span key={trait}>
                  {i > 0 && ', '}
                  <code>
                    {trait}: {value}
                  </code>
                </span>
              ))}
        </dd>
        <dt>Requires</dt>
        <dd>
          {requires.length === 0
            ? 'No traits'
            : requires.map(([trait, values], i) => (
                <span key={trait}>
                  {i > 0 && ', '}
                  <code>{trait}</code> in <code>[{values.join(', ')}]</code>
                </span>
              ))}
        </dd>
      </dl>
      {modules.length === 0 ? (
        <p>Uses no Go modules beyond the standard library.</p>
      ) : (
        <div
          className="docs-table"
          tabIndex={0}
          role="region"
          aria-label={`Modules of ${component.name} table`}
        >
          <table aria-label={`Modules of ${component.name}`}>
            <thead>
              <tr>
                <th scope="col">Module</th>
                <th scope="col">Role</th>
                <th scope="col">Supported range</th>
                <th scope="col">Default</th>
                <th scope="col">Constraints</th>
              </tr>
            </thead>
            <tbody>
              {modules.map((module) => (
                <tr key={module.path}>
                  <td>
                    <code>{module.path}</code>
                  </td>
                  <td>{module.role === 'tool' ? 'tool (fixed per release)' : 'library'}</td>
                  <td>
                    <code>{module.range}</code>
                  </td>
                  <td>
                    <code>{module.default}</code>
                  </td>
                  <td>
                    {module.constraints.length === 0
                      ? 'None'
                      : module.constraints.map((constraint, i) => (
                          <span key={i}>
                            {i > 0 && '; '}
                            when <code>{constraint.when}</code>, requires{' '}
                            {constraint.requires.target === 'go' ? (
                              'Go'
                            ) : (
                              <code>{constraint.requires.target}</code>
                            )}{' '}
                            <code>{constraint.requires.range}</code>
                          </span>
                        ))}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </article>
  );
}

/** The reference for the catalog the site is built from, for use in MDX. */
export function ComponentsReference() {
  return <CatalogReference catalog={builtCatalog} />;
}
