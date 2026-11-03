import defaultMdxComponents from 'fumadocs-ui/mdx';
import type { MDXComponents } from 'mdx/types';
import { Button } from '@/components/ui/button';
import { ComponentsReference } from '@/components/components-reference';
import { Callout } from '@/components/docs/callout';
import { DocsPre } from '@/components/docs/code-block';

/** Components available in every MDX page, optionally extended or overridden. */
export function getMDXComponents(components?: MDXComponents) {
  return {
    ...defaultMdxComponents,
    Button,
    Callout,
    ComponentsReference,
    pre: DocsPre,
    ...components,
  } satisfies MDXComponents;
}

export const useMDXComponents = getMDXComponents;

declare global {
  type MDXProvidedComponents = ReturnType<typeof getMDXComponents>;
}
