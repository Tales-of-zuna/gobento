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
        // The site palette: Catppuccin Latte (light) and Frappé (dark).
        themes: { light: 'catppuccin-latte', dark: 'catppuccin-frappe' },
        // Lets components/docs/code-block name each block after its language.
        addLanguageClass: true,
        // Latte's accents are pastel and most fall below 4.5:1 as text on a
        // light code background, so each is replaced by the same hue, darker
        // (OKLCH lightness lowered, chroma kept in gamut), until it reaches
        // 4.6:1 on Latte's darkest surface (crust, #dce0e8). Frappé passes on
        // its own surfaces except its comment greys, which are lightened.
        colorReplacements: {
          'catppuccin-latte': {
            '#dc8a78': '#974c3c', // rosewater
            '#dd7878': '#a24346', // flamingo
            '#ea76cb': '#a33489', // pink
            '#8839ef': '#802de5', // mauve
            '#d20f39': '#c30232', // red
            '#e64553': '#bf1734', // maroon
            '#fe640b': '#aa4002', // peach
            '#df8e1d': '#8c5602', // yellow
            '#40a02b': '#1d7102', // green
            '#179299': '#026d72', // teal
            '#04a5e5': '#056892', // sky
            '#209fb5': '#026c7c', // sapphire
            '#1e66f5': '#0a55e3', // blue
            '#7287fd': '#4856c7', // lavender
            '#6c6f85': '#5d6076', // subtext0
            '#7c7f93': '#5e6174', // overlay2 (comments)
            '#8c8fa1': '#5f6172', // overlay1
          },
          'catppuccin-frappe': {
            '#949cbb': '#9da5c5', // overlay2 (comments)
            '#838ba7': '#9da5c2', // overlay1
          },
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
