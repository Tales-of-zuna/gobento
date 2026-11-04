import type { KindId } from '@/lib/catalog/schema';
import s from './home.module.css';

/** The CSS class carrying a component kind's colors (--k, --ks, --ki). */
export const KIND_CLASS: Readonly<Record<KindId, string>> = {
  architecture: s.kArch!,
  framework: s.kHttp!,
  database: s.kDb!,
  data: s.kData!,
  frontend: s.kFront!,
  css: s.kCss!,
};
