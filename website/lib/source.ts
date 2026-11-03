import { loader } from 'fumadocs-core/source';
import { applyMdxPreset } from 'fumadocs-mdx/config';
import { defineDocs } from 'fumadocs-mdx/macro';
import { metaSchema, pageSchema } from 'fumadocs-core/source/schema';
import { docsRoute } from './shared';

const docs = defineDocs({
  dir: 'content/docs',
  docs: {
    schema: pageSchema,
    mdxOptions: applyMdxPreset({
      rehypeCodeOptions: {
        themes: { light: 'github-light', dark: 'github-dark' },
        // Lets components/docs/code-block name each block after its language.
        addLanguageClass: true,
        // Shiki's GitHub themes, with the token colors that fall below 4.5:1
        // on the code block's --paper background darkened (light) or
        // lightened (dark) just enough to pass WCAG AA.
        colorReplacements: {
          'github-light': { '#d73a49': '#c4293a', '#e36209': '#b35000' },
          'github-dark': { '#6a737d': '#9aa3ad' },
        },
      },
    }),
  },
  meta: { schema: metaSchema },
});

/** Fumadocs source for `content/docs`; see https://fumadocs.dev/docs/headless/source-api */
export const source = loader({
  baseUrl: docsRoute,
  source: docs.toFumadocsSource(),
});
