import s from './home.module.css';

/** The small four-compartment bento logo next to the wordmark (decorative). */
export function BentoMark() {
  return <span className={s.logo} aria-hidden="true" />;
}
