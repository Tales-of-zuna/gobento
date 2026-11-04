/**
 * Component kinds, kept apart from the zod schema (./schema) so client code
 * that only needs the kind list does not ship the validator.
 */

/** Component kinds, in the order their flags appear on `gobento new`. */
export const KIND_IDS = ['architecture', 'framework', 'database', 'data'] as const;

/** A component kind. */
export type KindId = (typeof KIND_IDS)[number];

/**
 * A kind title for use inside a sentence: the first letter is lower-cased
 * unless the first word is an acronym (`Database` → `database`, but
 * `HTTP framework` stays).
 */
export function titleInSentence(title: string): string {
  if (/^[A-Z]{2}/.test(title)) return title;
  return title.charAt(0).toLowerCase() + title.slice(1);
}
