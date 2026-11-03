'use client';
/** The quick-start command with a copy button, a "Copied" state and a toast. */
import { useEffect, useRef, useState } from 'react';
import s from './home.module.css';

export function CopyCommand({ command }: { command: string }) {
  const [copied, setCopied] = useState(false);
  const [toast, setToast] = useState('');
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined);

  useEffect(() => () => clearTimeout(timer.current), []);

  const copy = async () => {
    let message = 'Copied to clipboard';
    try {
      await navigator.clipboard.writeText(command);
      setCopied(true);
    } catch {
      message = 'Copy failed: select the command and copy it manually';
    }
    setToast(message);
    clearTimeout(timer.current);
    timer.current = setTimeout(() => {
      setCopied(false);
      setToast('');
    }, 1800);
  };

  return (
    <>
      <div className={s.cmd}>
        <code>
          <span className={s.prompt} aria-hidden="true">
            ${' '}
          </span>
          {command}
          <span className={s.cur} aria-hidden="true" />
        </code>
        <button type="button" className={s.copy} data-copied={copied} onClick={copy}>
          {copied ? (
            'Copied ✓'
          ) : (
            <>
              Copy<span className="sr-only"> quick-start command</span>
            </>
          )}
        </button>
      </div>
      <div role="status" aria-live="polite" className={s.toast}>
        {toast}
      </div>
    </>
  );
}
