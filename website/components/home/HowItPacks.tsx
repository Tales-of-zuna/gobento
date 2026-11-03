import { Item, Stagger } from './motion';
import s from './home.module.css';

const STEPS = [
  {
    kind: s.kHttp,
    title: 'Choose',
    text: 'Pick an architecture, HTTP framework, database and data access, in the CLI or the configurator.',
  },
  {
    kind: s.kDb,
    title: 'Pack',
    text: 'Versions are resolved from compatibility rules kept as data, then a ready-to-run module is written.',
  },
  {
    kind: s.kData,
    title: 'Prove',
    text: 'Every supported combination is generated, built and tested in CI before it is offered to you.',
  },
];

/** "How it packs": three compartments in a lacquer tray. */
export function HowItPacks() {
  return (
    <section className={`${s.band} ${s.edge}`} aria-labelledby="packs-title">
      <div className={s.in}>
        <div className={s.head}>
          <div>
            <p className={s.eyebrow}>
              <i aria-hidden="true" />
              How it packs
            </p>
            <h2 id="packs-title" className={s.h2}>
              Three steps, one box.
            </h2>
          </div>
        </div>
        <Stagger inView>
          <ol className={s.tray}>
            {STEPS.map((step, i) => (
              <Item as="li" kind="pop" key={step.title}>
                <div className={`${s.cell} ${step.kind}`}>
                  <span className={s.num} aria-hidden="true">
                    {String(i + 1).padStart(2, '0')}
                  </span>
                  <h3>{step.title}</h3>
                  <p>{step.text}</p>
                </div>
              </Item>
            ))}
          </ol>
        </Stagger>
      </div>
    </section>
  );
}
