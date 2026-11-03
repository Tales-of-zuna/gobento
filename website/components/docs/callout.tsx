import type { ReactNode } from 'react';

/** Callout types accepted from MDX, including Fumadocs' names. */
export type CalloutType =
  'info' | 'note' | 'tip' | 'idea' | 'success' | 'warn' | 'warning' | 'error';

type Variant = 'note' | 'tip' | 'warn';

const VARIANTS: Record<CalloutType, Variant> = {
  info: 'note',
  note: 'note',
  tip: 'tip',
  idea: 'tip',
  success: 'tip',
  warn: 'warn',
  warning: 'warn',
  error: 'warn',
};

const LABELS: Record<Variant, string> = { note: 'Note', tip: 'Tip', warn: 'Warning' };

/**
 * A two-compartment bento callout ("Docs v2"): a colored compartment with a
 * CSS shape (note: sora circle, tip: matcha diamond, warning: tamago
 * triangle) and the message. The shape is decorative; the label names the type.
 */
export function Callout({
  type = 'info',
  title,
  children,
}: {
  type?: CalloutType;
  title?: ReactNode;
  children?: ReactNode;
}) {
  const variant = VARIANTS[type] ?? 'note';
  return (
    <div className={`docs-co docs-co-${variant}`} role="note">
      <span className="docs-co-ic" aria-hidden="true">
        <i className="docs-shp" />
      </span>
      <div className="docs-co-b">
        <span className="docs-co-k">{title ?? LABELS[variant]}</span>
        <div className="docs-co-t">{children}</div>
      </div>
    </div>
  );
}
