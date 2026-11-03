import { catalog } from '@/lib/catalog/data';
import { KIND_IDS } from '@/lib/catalog/schema';
import { beltPlates, KIND_SHORT } from '@/lib/home/menu';
import { KIND_CLASS } from './kinds';
import { Item, Stagger } from './motion';
import s from './home.module.css';

/**
 * "On the belt": the dark lacquer band with a looping conveyor of plates.
 * The plate list is rendered twice for a seamless loop; the copy is hidden
 * from assistive technology, and the first copy is a plain list.
 */
export function Conveyor() {
  const plates = beltPlates(catalog);
  const renderPlates = (copy: boolean) =>
    plates.map((plate) => (
      <li
        key={`${copy ? 'b' : 'a'}-${plate.kind}-${plate.name}`}
        className={`${s.plate} ${KIND_CLASS[plate.kind]} ${plate.status === 'roadmap' ? s.road : ''}`}
        aria-hidden={copy || undefined}
      >
        <span className={s.pk}>{KIND_SHORT[plate.kind]}</span>
        <span className={s.pn}>{plate.name}</span>
        <span className={s.pt}>{plate.status}</span>
      </li>
    ));

  return (
    <section className={`${s.band} ${s.bLacq} ${s.edge}`} aria-labelledby="belt-title">
      <Stagger inView className={s.in}>
        <Item>
          <p className={s.eyebrow}>
            <i aria-hidden="true" />
            On the belt
          </p>
        </Item>
        <Item>
          <h2 id="belt-title" className={s.h2}>
            Today&apos;s menu, and what comes next.
          </h2>
        </Item>
        <Item>
          <p className={s.lead}>
            Solid plates are supported today. Dashed plates are on the roadmap and join once they
            pass CI.
          </p>
        </Item>
      </Stagger>
      <div className={s.belt}>
        <ul className={s.track} aria-label="Components">
          {renderPlates(false)}
          {renderPlates(true)}
        </ul>
      </div>
      <div className={s.in}>
        <ul className={s.legend} aria-label="Component kinds">
          {KIND_IDS.map((kind) => (
            <li key={kind} className={KIND_CLASS[kind]}>
              <i aria-hidden="true" />
              {catalog.kinds.find((k) => k.id === kind)?.title ?? kind}
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
