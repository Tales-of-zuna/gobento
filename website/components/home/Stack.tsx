import { catalog } from '@/lib/catalog/data';
import { stackCells } from '@/lib/home/menu';
import { configuratorRoute } from './links';
import { KIND_CLASS } from './kinds';
import { Item, PressLink, Stagger } from './motion';
import s from './home.module.css';

/** "Pick your stack": one compartment per kind, colored by kind. */
export function Stack() {
  const cells = stackCells(catalog);
  return (
    <section className={`${s.band} ${s.bTamago} ${s.edge}`} aria-labelledby="stack-title">
      <div className={`${s.pat} ${s.mizu}`} aria-hidden="true" />
      <div className={s.in}>
        <div className={s.head}>
          <div>
            <p className={`${s.eyebrow} ${s.kDb}`}>
              <i aria-hidden="true" />
              Pick your stack
            </p>
            <h2 id="stack-title" className={s.h2}>
              One choice per compartment.
            </h2>
            <p className={s.lead}>
              Every option below is proven in CI. Others join once they pass the same matrix.
            </p>
          </div>
        </div>
        <Stagger inView>
          <ul className={`${s.tray} ${s.five}`}>
            {cells.map((cell) => (
              <Item as="li" kind="pop" key={cell.kind}>
                <div className={`${s.cell} ${KIND_CLASS[cell.kind]}`}>
                  <h3 className={s.kind}>
                    <i aria-hidden="true" />
                    {cell.title}
                    {cell.optional && <small className={s.optl}>optional</small>}
                  </h3>
                  {cell.options.map((option) => (
                    <div key={option.name} className={`${s.opt} ${s.sel}`}>
                      <b>{option.name}</b>
                      <span>{option.description}</span>
                    </div>
                  ))}
                  <p className={s.soonl}>roadmap</p>
                  <ul className={s.chips} aria-label={`${cell.title} roadmap`}>
                    {(cell.roadmap.length > 0 ? cell.roadmap : ['more later']).map((name) => (
                      <li key={name} className={s.soon}>
                        {name}
                      </li>
                    ))}
                  </ul>
                </div>
              </Item>
            ))}
          </ul>
        </Stagger>
        <div className={s.more}>
          <PressLink className={s.btn} href={configuratorRoute} prefetch={false}>
            Build yours in the configurator
          </PressLink>
        </div>
      </div>
    </section>
  );
}
