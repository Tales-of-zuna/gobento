import { Fragment } from 'react';
import { Item, Stagger } from './motion';
import s from './home.module.css';

const OSES = ['Linux', 'macOS', 'Windows'] as const;
const GO = ['Go stable', 'Go oldstable'] as const;
const STEPS = ['generate', 'build', 'test'] as const;

/** "Proven in CI": the OS × Go matrix every combination passes, with a stamp. */
export function ProvenInCI() {
  return (
    <section className={`${s.band} ${s.bMatcha} ${s.edge}`} aria-labelledby="ci-title">
      <div className={`${s.pat} ${s.ichi}`} aria-hidden="true" />
      <div className={s.in}>
        <div className={s.head}>
          <div>
            <p className={`${s.eyebrow} ${s.kData}`}>
              <i aria-hidden="true" />
              Proven in CI
            </p>
            <h2 id="ci-title" className={s.h2}>
              Built on all three, before you ever run it.
            </h2>
            <p className={s.lead}>
              Every supported combination is generated, built and tested on Linux, macOS and
              Windows, against the current and previous Go release.
            </p>
          </div>
          <Stagger inView>
            <Item kind="pop">
              <p
                className={s.hanko}
                role="img"
                aria-label="Stamp: passed on 3 OS and 2 Go versions"
              >
                <b aria-hidden="true">PASSED</b>
                <span aria-hidden="true">3 OS · 2 Go</span>
              </p>
            </Item>
          </Stagger>
        </div>
        <Stagger inView className={s.ci}>
          <span className={s.colh} aria-hidden="true" />
          {GO.map((go) => (
            <span key={go} className={s.colh} aria-hidden="true">
              {go}
            </span>
          ))}
          {OSES.map((os) => (
            <Fragment key={os}>
              <span className={s.os}>{os}</span>
              {GO.map((go) => (
                <Item key={go}>
                  <ul className={s.run} aria-label={`${os}, ${go}`}>
                    <li className={s.runGo} aria-hidden="true">
                      {go}
                    </li>
                    {STEPS.map((step) => (
                      <li key={step} className={s.step}>
                        {step}
                      </li>
                    ))}
                  </ul>
                </Item>
              ))}
            </Fragment>
          ))}
        </Stagger>
      </div>
    </section>
  );
}
