import { CodeBlock, Pre } from 'fumadocs-ui/components/codeblock';
import { cache, isValidElement, type ComponentProps, type ReactNode } from 'react';

/** Readable names of the languages the docs use; others show their Shiki id. */
const LANGUAGE_NAMES: Record<string, string> = {
  sh: 'shell',
  bash: 'shell',
  shell: 'shell',
  powershell: 'PowerShell',
  go: 'Go',
  yaml: 'YAML',
  json: 'JSON',
  sql: 'SQL',
  md: 'Markdown',
  text: 'text',
  plaintext: 'text',
};

/** A counter per page render (React's `cache` is scoped to one server request). */
const pageCounter = cache(() => ({ value: 0 }));

/** The Shiki language of a highlighted block, from its `<code class="language-…">`. */
function languageOf(children: ReactNode): string | undefined {
  if (!isValidElement<{ className?: string }>(children)) return undefined;
  return /\blanguage-(\S+)/.exec(children.props.className ?? '')?.[1];
}

/** The accessible name of a code block, unique on its page. */
export function codeBlockLabel(index: number, title?: string, language?: string): string {
  const base = `Code example ${index}`;
  if (title) return `${base}: ${title}`;
  if (language) return `${base} (${LANGUAGE_NAMES[language] ?? language})`;
  return base;
}

/**
 * Fumadocs' code block, with a name for its scrollable region. Fumadocs makes
 * each block's viewport a focusable `role="region"` (so keyboard users can
 * scroll it), but leaves it unnamed, and a page with several blocks then has
 * several identical, unnamed landmarks (axe `landmark-unique`). Each block is
 * named after its position on the page, its title, or its language.
 */
export function DocsPre({ title, ...props }: ComponentProps<'pre'>) {
  const counter = pageCounter();
  counter.value += 1;
  const label = codeBlockLabel(counter.value, title, languageOf(props.children));
  return (
    <CodeBlock
      {...(props as Omit<ComponentProps<typeof CodeBlock>, 'title'>)}
      title={title}
      viewportProps={{ 'aria-label': label }}
    >
      <Pre>{props.children}</Pre>
    </CodeBlock>
  );
}
