'use client';
/**
 * The live `gobento new` command with its copy control, or the reason no
 * command can be offered yet. The command cross-fades whenever it changes.
 */
import { useReducedMotion } from 'motion/react';
import * as m from 'motion/react-m';
import styles from './configurator.module.css';

/** Props of {@link CommandBox}. */
export interface CommandBoxProps {
  /** The command, or undefined when the state is invalid. */
  readonly command?: string;
  readonly blocked: string;
  readonly copied: boolean;
  readonly onCopy: () => void;
}

/** Splits a command into the program and one word per flag with its value. */
export function commandWords(command: string): { text: string; flag?: string }[] {
  const [head = '', ...flags] = command.split(' --');
  return [{ text: head }, ...flags.map((f) => ({ text: `--${f}`, flag: f.split(' ')[0] }))];
}

export function CommandBox({ command, blocked, copied, onCopy }: CommandBoxProps) {
  if (command === undefined) {
    return <div className={styles.blocked}>{blocked}</div>;
  }
  return <CommandLine command={command} copied={copied} onCopy={onCopy} />;
}

function CommandLine({
  command,
  copied,
  onCopy,
}: Omit<CommandBoxProps, 'blocked'> & { command: string }) {
  // Under reduced motion the new command simply appears.
  const reduce = useReducedMotion();
  const words = commandWords(command);
  return (
    <div className={styles.cmd}>
      <span className={styles.cmdText}>
        <span className={styles.prompt} aria-hidden>
          $
        </span>
        <m.code
          key={command}
          data-command
          initial={reduce ? false : { opacity: 0, y: 4 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.22, ease: 'easeOut' }}
        >
          {words.map((word, i) => (
            <span key={i}>
              {i > 0 && ' '}
              <span className={styles.word} data-flag={word.flag}>
                {word.text}
              </span>
            </span>
          ))}
        </m.code>
      </span>
      <button
        type="button"
        className={styles.copy}
        data-copied={copied}
        onClick={onCopy}
        aria-label={copied ? 'Copied' : 'Copy command'}
      >
        {copied ? 'Copied ✓' : 'Copy'}
      </button>
    </div>
  );
}
