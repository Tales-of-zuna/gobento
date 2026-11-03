'use client';
/**
 * The configurator (design D4): choices on the left, the packing stage, live
 * command, share control, and file preview on the right. On phones the stage
 * comes first and the command sits in a bar pinned to the bottom.
 *
 * State lives here and is mirrored into the URL query, so a copied link
 * restores the same choices. Every rule comes from the catalog through
 * `buildCommand`, the compat helpers, and the URL-state codec.
 */
import { MotionConfig } from 'motion/react';
import { useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';
import { prefersReducedMotion } from '@/components/mascot/support';
import type { Catalog, KindId } from '@/lib/catalog/schema';
import { buildCommand, type ConfigState } from '@/lib/command/build';
import { decodeState, defaultState, encodeState } from '@/lib/command/url-state';
import { ChoiceGroup } from './ChoiceGroup';
import { CommandBox } from './CommandBox';
import styles from './configurator.module.css';
import { PackingStage } from './PackingStage';
import {
  blockedMessage,
  errorsByField,
  fileTree,
  goOptions,
  kindGroups,
  moduleFields,
  trayCells,
  withPin,
} from './view-model';

/** Props of {@link Configurator}. */
export interface ConfiguratorProps {
  /** The catalog to configure from; the page passes the one the site is built with. */
  readonly catalog: Catalog;
}

interface Toast {
  readonly message: string;
  readonly kind: 'command' | 'link' | 'error';
}

const TOAST_MS = 2000;
const PACK_STEP_MS = 320;

async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}

export function Configurator({ catalog }: ConfiguratorProps) {
  const id = useId();
  const [state, setState] = useState<ConfigState>(() => defaultState(catalog));
  const [notices, setNotices] = useState<readonly string[]>([]);
  const [loaded, setLoaded] = useState(false);
  const [advanced, setAdvanced] = useState(false);
  const [toast, setToast] = useState<Toast | undefined>();
  const [packCount, setPackCount] = useState(4);
  const toastTimer = useRef<number | undefined>(undefined);
  const packTimers = useRef<number[]>([]);

  // Restore the state from a shared link once, after hydration.
  useEffect(() => {
    const decoded = decodeState(catalog, window.location.search);
    /* eslint-disable react-hooks/set-state-in-effect -- the URL is only readable after hydration */
    setState(decoded.state);
    setNotices(decoded.notices);
    if (Object.keys(decoded.state.pins).length > 0) setAdvanced(true);
    setLoaded(true);
    /* eslint-enable react-hooks/set-state-in-effect */
  }, [catalog]);

  // Keep the URL in step with the state, without adding history entries.
  useEffect(() => {
    if (!loaded) return;
    const query = encodeState(catalog, state);
    const url = `${window.location.pathname}${query ? `?${query}` : ''}${window.location.hash}`;
    window.history.replaceState(window.history.state, '', url);
  }, [catalog, state, loaded]);

  const repack = useCallback(() => {
    packTimers.current.forEach((timer) => window.clearTimeout(timer));
    packTimers.current = [];
    if (prefersReducedMotion(window.matchMedia?.bind(window))) {
      setPackCount(4);
      return;
    }
    setPackCount(0);
    for (let n = 1; n <= 4; n++) {
      packTimers.current.push(window.setTimeout(() => setPackCount(n), 250 + n * PACK_STEP_MS));
    }
  }, []);

  // Pack the bento once on arrival.
  useEffect(() => {
    const start = window.setTimeout(repack, 0);
    const timers = packTimers;
    return () => {
      window.clearTimeout(start);
      timers.current.forEach((timer) => window.clearTimeout(timer));
    };
  }, [repack]);

  useEffect(() => () => window.clearTimeout(toastTimer.current), []);

  const flash = (next: Toast) => {
    window.clearTimeout(toastTimer.current);
    setToast(next);
    toastTimer.current = window.setTimeout(() => setToast(undefined), TOAST_MS);
  };

  const result = useMemo(() => buildCommand(catalog, state), [catalog, state]);
  const errors = useMemo(() => errorsByField(result), [result]);
  const groups = useMemo(() => kindGroups(catalog, state, errors), [catalog, state, errors]);
  const goCards = useMemo(() => goOptions(catalog, state), [catalog, state]);
  const modules = useMemo(() => moduleFields(catalog, state, errors), [catalog, state, errors]);
  const cells = useMemo(() => trayCells(catalog, state), [catalog, state]);
  const nameError = errors.get('name');
  const moduleError = errors.get('module');
  const tree = fileTree(catalog, state, nameError === undefined);
  const command = result.ok ? result.command : undefined;
  const blocked = blockedMessage(catalog, result);

  const packed = [0, 1, 2, 3].map((i) => i < packCount && cells[i]?.value !== undefined) as [
    boolean,
    boolean,
    boolean,
    boolean,
  ];

  const update = (patch: Partial<ConfigState>) => setState((s) => ({ ...s, ...patch }));
  const choose = (kind: KindId, name: string) =>
    setState((s) => ({ ...s, components: { ...s.components, [kind]: name } }));

  const copyCommand = async () => {
    if (command === undefined) return;
    flash(
      (await copyText(command))
        ? { message: 'Command copied to clipboard', kind: 'command' }
        : { message: 'Copying failed; select the command to copy it.', kind: 'error' },
    );
  };

  const share = async () => {
    if (!result.ok) return;
    flash(
      (await copyText(window.location.href))
        ? { message: 'Link to this configuration copied', kind: 'link' }
        : { message: 'Copying failed; copy the address bar instead.', kind: 'error' },
    );
  };

  const copied = toast?.kind === 'command';
  const toastNode = <div className={styles.toast}>{toast?.message}</div>;

  return (
    <MotionConfig reducedMotion="user">
      <div className={styles.root} data-hydrated={loaded}>
        <header className={styles.head}>
          <div className={`${styles.inner} ${styles.headInner}`}>
            <span className={styles.pill}>Early development · not released yet</span>
            <h1 className={styles.h1}>Pack your project</h1>
            <p className={styles.sub}>
              Choose each part of your Go project. The command updates as you go; copy it, or share
              this exact setup with a link.
            </p>
            {notices.length > 0 && (
              <div className={styles.notice} role="status">
                <div>
                  <b>Part of this shared link was reset.</b>
                  <ul>
                    {notices.map((notice) => (
                      <li key={notice}>{notice}</li>
                    ))}
                  </ul>
                </div>
                <button type="button" className={styles.ghost} onClick={() => setNotices([])}>
                  Dismiss
                </button>
              </div>
            )}
          </div>
        </header>

        <div className={`${styles.inner} ${styles.cols}`}>
          <div className={styles.choices}>
            <section className={styles.group} aria-labelledby={`${id}-project`}>
              <h2 className={styles.gt} id={`${id}-project`}>
                Project
              </h2>
              <div className={styles.fields}>
                <div className={styles.field}>
                  <label className={styles.label} htmlFor={`${id}-name`}>
                    Project name
                  </label>
                  <input
                    id={`${id}-name`}
                    className={styles.input}
                    value={state.name}
                    onChange={(e) => update({ name: e.target.value })}
                    aria-invalid={nameError !== undefined}
                    aria-describedby={`${id}-name-hint`}
                    autoComplete="off"
                    spellCheck={false}
                  />
                  <p
                    id={`${id}-name-hint`}
                    className={nameError ? styles.hintError : styles.hint}
                    aria-live="polite"
                  >
                    {nameError ?? 'Lowercase letters, digits, - and _'}
                  </p>
                </div>
                <div className={styles.field}>
                  <label className={styles.label} htmlFor={`${id}-module`}>
                    Module path
                  </label>
                  <input
                    id={`${id}-module`}
                    className={styles.input}
                    value={state.module}
                    placeholder={nameError ? 'Defaults to the project name' : state.name}
                    onChange={(e) => update({ module: e.target.value })}
                    aria-invalid={moduleError !== undefined}
                    aria-describedby={`${id}-module-hint`}
                    autoComplete="off"
                    spellCheck={false}
                  />
                  <p
                    id={`${id}-module-hint`}
                    className={moduleError ? styles.hintError : styles.hint}
                    aria-live="polite"
                  >
                    {moduleError ??
                      (state.module
                        ? 'Leave empty to use the project name'
                        : 'Empty: uses the project name')}
                  </p>
                </div>
              </div>
            </section>

            {groups.map((group) => (
              <ChoiceGroup
                key={group.kind}
                name={group.kind}
                title={group.title}
                description={group.description}
                options={group.options}
                error={group.error}
                onChange={(value) => choose(group.kind, value)}
              />
            ))}

            <ChoiceGroup
              name="go"
              title="Go version"
              description="Sets the go directive in go.mod."
              options={goCards}
              error={errors.get('go')}
              onChange={(go) => update({ go })}
            />

            <section className={styles.adv}>
              <button
                type="button"
                className={styles.advButton}
                aria-expanded={advanced}
                aria-controls={`${id}-advanced`}
                onClick={() => setAdvanced((open) => !open)}
              >
                <span>
                  Advanced <small>module versions</small>
                </span>
                <span className={styles.chev} aria-hidden>
                  {advanced ? '−' : '+'}
                </span>
              </button>
              <div className={styles.mods} id={`${id}-advanced`} hidden={!advanced}>
                {modules.map((module, i) =>
                  module.fixed ? (
                    <div className={styles.field} key={module.path}>
                      <span className={styles.label} id={`${id}-mod-${i}`}>
                        {module.path} ({module.usedBy})
                      </span>
                      <div
                        className={styles.fixed}
                        role="textbox"
                        aria-readonly="true"
                        aria-labelledby={`${id}-mod-${i}`}
                        aria-describedby={`${id}-mod-${i}-hint`}
                      >
                        <span>{module.value}</span>
                        <span className={styles.tag}>Fixed</span>
                      </div>
                      <p id={`${id}-mod-${i}-hint`} className={styles.hint}>
                        {module.hint}
                      </p>
                    </div>
                  ) : (
                    <div className={styles.field} key={module.path}>
                      <label className={styles.label} htmlFor={`${id}-mod-${i}`}>
                        {module.path}
                      </label>
                      <input
                        id={`${id}-mod-${i}`}
                        className={styles.input}
                        value={module.value}
                        onChange={(e) =>
                          update({
                            pins: withPin(catalog, state.pins, module.path, e.target.value),
                          })
                        }
                        aria-invalid={module.error !== undefined}
                        aria-describedby={`${id}-mod-${i}-hint`}
                        autoComplete="off"
                        spellCheck={false}
                      />
                      <p
                        id={`${id}-mod-${i}-hint`}
                        className={module.error ? styles.hintError : styles.hint}
                        aria-live="polite"
                      >
                        {module.error ?? module.hint}
                      </p>
                    </div>
                  ),
                )}
              </div>
            </section>
          </div>

          <aside className={styles.panel} aria-label="Your project">
            <PackingStage cells={cells} packed={packed} onRepack={repack} />

            <div className={`${styles.field} ${styles.deskOnly}`}>
              <h2 className={styles.eyebrow}>Your command</h2>
              <CommandBox
                command={command}
                blocked={blocked}
                copied={copied}
                onCopy={copyCommand}
              />
            </div>

            <div className={styles.shareRow}>
              <button
                type="button"
                className={styles.btn}
                aria-disabled={!result.ok}
                aria-describedby={result.ok ? undefined : `${id}-share-hint`}
                onClick={share}
              >
                Share this configuration
              </button>
              {!result.ok && (
                <span id={`${id}-share-hint`} className={styles.hint}>
                  Available once every field is valid.
                </span>
              )}
            </div>

            <div className={styles.tree}>
              <h2 className={styles.eyebrow}>What you&apos;ll get</h2>
              <ul aria-label="Generated files">
                {tree.map((line, i) => (
                  <li key={i} data-depth={line.depth} data-dir={line.dir}>
                    {line.label}
                  </li>
                ))}
              </ul>
            </div>

            <div className={`${styles.toastSlot} ${styles.deskOnly}`} aria-live="polite">
              {toast && toastNode}
            </div>
          </aside>
        </div>

        <div className={styles.mbar}>
          <div aria-live="polite">{toast && toastNode}</div>
          <CommandBox command={command} blocked={blocked} copied={copied} onCopy={copyCommand} />
        </div>
      </div>
    </MotionConfig>
  );
}
