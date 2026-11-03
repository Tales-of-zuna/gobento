'use client';
/**
 * One group of choice cards (a component kind or the Go version) as a native
 * radio group: arrow keys move between enabled options, and disabled cards
 * explain themselves through aria-describedby.
 */
import * as m from 'motion/react-m';
import { useId } from 'react';
import styles from './configurator.module.css';
import type { OptionView } from './view-model';

/** Props of {@link ChoiceGroup}. */
export interface ChoiceGroupProps {
  /** Radio group name and the kind's colour key (`data-kind`). */
  readonly name: string;
  readonly title: string;
  readonly description: string;
  readonly options: readonly OptionView[];
  readonly error?: string;
  readonly onChange: (value: string) => void;
}

/** The selected card springs from its pressed position up to its raised toy shadow. */
const cardVariants = {
  idle: { x: 0, y: 0 },
  selected: { x: [3, 0], y: [3, 0] },
};
const spring = { type: 'spring', stiffness: 520, damping: 16 } as const;

function tagOf(option: OptionView): { text: string; bad: boolean } | undefined {
  if (option.status === 'soon') return { text: 'Coming soon', bad: false };
  if (option.status === 'incompatible') return { text: 'Incompatible', bad: true };
  return undefined;
}

export function ChoiceGroup({
  name,
  title,
  description,
  options,
  error,
  onChange,
}: ChoiceGroupProps) {
  const id = useId();
  return (
    <fieldset className={styles.group} data-kind={name} aria-describedby={error && `${id}-err`}>
      <legend className={styles.legend}>
        <span className={styles.dot} aria-hidden />
        <h2 className={styles.gt}>{title}</h2>
        <span className={styles.gd}>{description}</span>
      </legend>
      <div className={styles.opts}>
        {options.map((option) => {
          const enabled = option.status === 'selected' || option.status === 'available';
          const tag = tagOf(option);
          const optionId = `${id}-${option.value}`;
          const detail =
            option.status === 'incompatible' ? option.reasons.join(' ') : option.description;
          return (
            <m.label
              key={option.value}
              className={styles.opt}
              data-status={option.status}
              initial={false}
              animate={option.status === 'selected' ? 'selected' : 'idle'}
              variants={cardVariants}
              transition={spring}
            >
              <input
                type="radio"
                className={styles.srOnly}
                name={name}
                value={option.value}
                checked={option.status === 'selected'}
                disabled={!enabled}
                onChange={() => onChange(option.value)}
                aria-labelledby={`${optionId}-t`}
                aria-describedby={`${optionId}-d`}
              />
              <span className={styles.ot}>
                <b id={`${optionId}-t`} className={styles.optTitle}>
                  {option.title}
                </b>
                {tag && <span className={tag.bad ? styles.tagBad : styles.tag}>{tag.text}</span>}
              </span>
              <span
                id={`${optionId}-d`}
                className={option.status === 'incompatible' ? styles.why : styles.od}
              >
                {tag && <span className={styles.srOnly}>{tag.text}. </span>}
                {detail}
              </span>
            </m.label>
          );
        })}
      </div>
      {error && (
        <p id={`${id}-err`} className={styles.groupError}>
          {error}
        </p>
      )}
    </fieldset>
  );
}
