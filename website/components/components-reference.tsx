/**
 * The Components reference (design D5): every component of the catalog with
 * its traits, module versions, vendored assets, the platform binaries it
 * downloads on demand, and npm packages, rendered at build time so that no
 * docs content contains a version number. Optional kinds are marked.
 */
import { catalog as builtCatalog } from '@/lib/catalog/data';
import {
  KIND_IDS,
  type Asset,
  type Binary,
  type Catalog,
  type Component,
  type Module,
  type NpmPackage,
} from '@/lib/catalog/schema';
import { KIND_FLAGS } from '@/lib/command/build';

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
      {catalog.npm.length > 0 && (
        <>
          <h2 id="node">Node.js</h2>
          <p>
            Components with npm packages need Node.js and npm on the machine that runs{' '}
            <code>gobento new</code>; the generated project&apos;s tooling (container image and CI)
            uses the release line below.
          </p>
          <div
            className="docs-table"
            tabIndex={0}
            role="region"
            aria-label="Node.js versions table"
          >
            <table aria-label="Node.js versions">
              <thead>
                <tr>
                  <th scope="col">Needed to generate</th>
                  <th scope="col">Tooling release line</th>
                  <th scope="col">Used by</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>
                    <code>{catalog.node.range}</code>
                  </td>
                  <td>
                    <code>{catalog.node.default}</code>
                  </td>
                  <td>
                    {[...new Set(catalog.npm.flatMap((pkg) => pkg.components))]
                      .sort()
                      .map((id, i) => (
                        <span key={id}>
                          {i > 0 && ', '}
                          <code>{id.slice(id.indexOf('/') + 1)}</code>
                        </span>
                      ))}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </>
      )}
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
              {kind.optional && <span className="docs-optional">optional</span>}
            </h2>
            <p>{kind.description}</p>
            {kind.optional && (
              <p>
                Optional: without <code>--{KIND_FLAGS[kindId]}</code> (or with{' '}
                <code>--{KIND_FLAGS[kindId]} none</code>) the project has none.
              </p>
            )}
            {components.map((component) => (
              <ComponentEntry
                key={component.id}
                component={component}
                kindTitle={kind.title}
                modules={catalog.modules.filter((m) => m.components.includes(component.id))}
                assets={catalog.assets.filter((a) => a.components.includes(component.id))}
                binaries={catalog.binaries.filter((b) => b.components.includes(component.id))}
                packages={catalog.npm.filter((p) => p.components.includes(component.id))}
                node={catalog.node}
              />
            ))}
          </section>
        );
      })}
    </>
  );
}

/** How each npm package role reads in the reference. */
const NPM_ROLES: Readonly<Record<NpmPackage['role'], string>> = {
  scaffolder: 'scaffolder, run at generation time (fixed per release)',
  dependency: 'dependency (fixed per release)',
  'dev-dependency': 'dev dependency (fixed per release)',
};

function ComponentEntry({
  component,
  kindTitle,
  modules,
  assets,
  binaries,
  packages,
  node,
}: {
  component: Component;
  kindTitle: string;
  modules: Module[];
  assets: Asset[];
  binaries: Binary[];
  packages: NpmPackage[];
  node: Catalog['node'];
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
                    {Object.entries(module.traits).map(([trait, values]) => (
                      <span key={trait}>
                        {' '}
                        (only when <code>{trait}</code> is{' '}
                        {values.map((value, i) => (
                          <span key={value}>
                            {i > 0 && ' or '}
                            <code>{value}</code>
                          </span>
                        ))}
                        )
                      </span>
                    ))}
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
      {assets.length > 0 && (
        <div
          className="docs-table"
          tabIndex={0}
          role="region"
          aria-label={`Vendored assets of ${component.name} table`}
        >
          <table aria-label={`Vendored assets of ${component.name}`}>
            <thead>
              <tr>
                <th scope="col">Asset</th>
                <th scope="col">Role</th>
                <th scope="col">Version</th>
              </tr>
            </thead>
            <tbody>
              {assets.map((asset) => (
                <tr key={asset.name}>
                  <td>
                    <code>{asset.name}</code>
                  </td>
                  <td>vendored (fixed per release)</td>
                  <td>
                    <code>{asset.version}</code>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {binaries.length > 0 && (
        <div
          className="docs-table"
          tabIndex={0}
          role="region"
          aria-label={`Binaries of ${component.name} table`}
        >
          <table aria-label={`Binaries of ${component.name}`}>
            <thead>
              <tr>
                <th scope="col">Binary</th>
                <th scope="col">Role</th>
                <th scope="col">Version</th>
                <th scope="col">Platforms</th>
              </tr>
            </thead>
            <tbody>
              {binaries.map((binary) => (
                <tr key={binary.name}>
                  <td>
                    <code>{binary.name}</code>
                  </td>
                  <td>downloaded on demand (fixed per release)</td>
                  <td>
                    <code>{binary.version}</code>
                  </td>
                  <td>
                    {binary.platforms.map((platform, i) => (
                      <span key={platform}>
                        {i > 0 && ', '}
                        <code>{platform}</code>
                      </span>
                    ))}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {packages.length > 0 && (
        <>
          <div
            className="docs-table"
            tabIndex={0}
            role="region"
            aria-label={`npm packages of ${component.name} table`}
          >
            <table aria-label={`npm packages of ${component.name}`}>
              <thead>
                <tr>
                  <th scope="col">Package</th>
                  <th scope="col">Role</th>
                  <th scope="col">Version</th>
                </tr>
              </thead>
              <tbody>
                {packages.map((pkg) => (
                  <tr key={pkg.name}>
                    <td>
                      <code>{pkg.name}</code>
                    </td>
                    <td>{NPM_ROLES[pkg.role]}</td>
                    <td>
                      <code>{pkg.version}</code>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p>
            Generating it needs Node.js <code>{node.range}</code>; the project&apos;s tooling uses
            Node.js <code>{node.default}</code>.
          </p>
        </>
      )}
    </article>
  );
}

/** The reference for the catalog the site is built from, for use in MDX. */
export function ComponentsReference() {
  return <CatalogReference catalog={builtCatalog} />;
}
