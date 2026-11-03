'use client';
/**
 * Motion helpers for the home page. Reveals animate opacity and position
 * once; under prefers-reduced-motion they jump straight to the end state,
 * which is also what the server renders after hydration. Presses spring.
 */
import Link from 'next/link';
import {
  motion,
  MotionConfig,
  useReducedMotion,
  type HTMLMotionProps,
  type Variants,
} from 'motion/react';
import type { ComponentProps, ReactNode } from 'react';

/** Wraps the page so Motion honours the visitor's reduced-motion setting. */
export function HomeMotion({ children }: { children: ReactNode }) {
  return <MotionConfig reducedMotion="user">{children}</MotionConfig>;
}

const hiddenRise = { opacity: 0, y: 18 };
const shownRise = { opacity: 1, y: 0 };
const hiddenPop = { opacity: 0, scale: 0.9, y: 10 };
const shownPop = { opacity: 1, scale: 1, y: 0 };

/** Builds an item's variants; under reduced motion the end state is reached at once. */
function itemVariants(kind: 'rise' | 'pop', reduce: boolean | null): Variants {
  const instant = { duration: 0 };
  if (kind === 'pop') {
    return {
      hidden: hiddenPop,
      shown: {
        ...shownPop,
        transition: reduce ? instant : { type: 'spring', stiffness: 320, damping: 18 },
      },
    };
  }
  return {
    hidden: hiddenRise,
    shown: {
      ...shownRise,
      transition: reduce ? instant : { duration: 0.7, ease: [0.2, 0.7, 0.2, 1] },
    },
  };
}

/** Plays its children's reveals in sequence when the page loads. */
export function Stagger({
  children,
  className,
  inView = false,
  ...rest
}: { children: ReactNode; className?: string; inView?: boolean } & Omit<
  HTMLMotionProps<'div'>,
  'children'
>) {
  const reduce = useReducedMotion();
  const play = reduce || !inView ? { animate: 'shown' } : { whileInView: 'shown' };
  return (
    <motion.div
      className={className}
      initial="hidden"
      {...play}
      viewport={{ once: true, amount: 0.25 }}
      variants={{
        hidden: {},
        shown: {
          transition: reduce ? { duration: 0 } : { staggerChildren: 0.1, delayChildren: 0.05 },
        },
      }}
      {...rest}
    >
      {children}
    </motion.div>
  );
}

/** One revealed child of {@link Stagger}: rises (default) or pops in. */
export function Item({
  children,
  className,
  kind = 'rise',
  as = 'div',
}: {
  children: ReactNode;
  className?: string;
  kind?: 'rise' | 'pop';
  as?: 'div' | 'li';
}) {
  const variants = itemVariants(kind, useReducedMotion());
  const Tag = as === 'li' ? motion.li : motion.div;
  return (
    <Tag className={className} variants={variants}>
      {children}
    </Tag>
  );
}

const MotionLink = motion.create(Link);

/** A link that lifts on hover and springs in on press. */
export function PressLink(props: ComponentProps<typeof MotionLink>) {
  return (
    <MotionLink
      whileHover={{ x: -2, y: -2 }}
      whileTap={{ x: 3, y: 3, scale: 0.98 }}
      transition={{ type: 'spring', stiffness: 500, damping: 22 }}
      {...props}
    />
  );
}

/** The external-link variant of {@link PressLink}. */
export function PressAnchor(props: HTMLMotionProps<'a'>) {
  return (
    <motion.a
      whileHover={{ x: -2, y: -2 }}
      whileTap={{ x: 3, y: 3, scale: 0.98 }}
      transition={{ type: 'spring', stiffness: 500, damping: 22 }}
      {...props}
    />
  );
}
