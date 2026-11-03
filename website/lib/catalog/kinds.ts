/**
 * Component kinds, kept apart from the zod schema (./schema) so client code
 * that only needs the kind list does not ship the validator.
 */

/** Component kinds, in the order their flags appear on `gobento new`. */
export const KIND_IDS = ['architecture', 'framework', 'database', 'data'] as const;

/** A component kind. */
export type KindId = (typeof KIND_IDS)[number];
