import { catalog } from '@/lib/catalog/data';
import { versionShowcase } from '@/lib/home/menu';
import s from './home.module.css';

/** Spaces a range's operator from its version for reading (`>=1.26` → `>= 1.26`). */
function spaced(range: string): string {
  return range.replace(/^(>=|<=|>|<|=)\s*/, '$1 ');
}

/**
 * "Versions you can trust": a rule from the catalog and the Go choice it
 * rules out, shown the way the configurator shows a disabled option.
 */
export function Versions() {
  const show = versionShowcase(catalog);
  return (
    <section className={`${s.band} ${s.bSakura} ${s.edge}`} aria-labelledby="versions-title">
      <div className={`${s.in} ${s.split}`}>
        <div>
          <p className={`${s.eyebrow} ${s.kHttp}`}>
            <i aria-hidden="true" />
            Versions you can trust
          </p>
          <h2 id="versions-title" className={s.h2}>
            Compatibility is data, not guesswork.
          </h2>
          <p className={s.lead}>
            Every rule lives in a versions file you can read and review. When a choice cannot work,
            the configurator keeps it visible and tells you why.
          </p>
        </div>
        {show && (
          <div className={s.pick}>
            <pre className={s.code} tabIndex={0} aria-label="Example version rule">
              <code>
                <span className={s.c}># versions.yaml</span>
                {`\n${show.module}:\n  version: `}
                <span className={s.v}>{show.version}</span>
                {'\n  '}
                <span className={s.hl}>requires: go {spaced(show.goRange)}</span>
                {'\n  role: '}
                <span className={s.v}>{show.role}</span>
              </code>
            </pre>
            <p className={s.label} id="go-version-label">
              Go version
            </p>
            <ul className={`${s.opts} ${s.kData}`} aria-labelledby="go-version-label">
              <li className={`${s.opt} ${s.sel}`}>
                <b>Go {show.goDefault}</b>
                <span>default · meets every rule</span>
              </li>
              <li className={`${s.opt} ${s.off}`}>
                <b>
                  Go {show.goDisabled}
                  <span className="sr-only"> (unavailable)</span>
                </b>
                <span className={s.why}>{show.reason}</span>
              </li>
            </ul>
          </div>
        )}
      </div>
    </section>
  );
}
