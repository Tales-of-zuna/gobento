import { stillUrl } from '@/components/mascot/stills';
import { configuratorRoute, githubUrl } from './links';
import { Item, PressAnchor, PressLink, Stagger } from './motion';
import s from './home.module.css';

/** "Fill your box": the closing call to action, with the mascot peeking in. */
export function ClosingCta() {
  return (
    <section className={`${s.band} ${s.edge} ${s.bClose}`} aria-labelledby="close-title">
      <div className={s.in}>
        <Stagger inView>
          <Item kind="pop">
            <div className={s.lid}>
              <div>
                <h2 id="close-title" className={s.h2}>
                  Fill your box.
                </h2>
                <p className={s.lead}>
                  Try the configurator now; the CLI follows when the first release is ready.
                </p>
              </div>
              <div className={s.ctas}>
                <PressLink
                  className={`${s.btn} ${s.pri}`}
                  href={configuratorRoute}
                  prefetch={false}
                >
                  Open the configurator
                </PressLink>
                <PressAnchor className={s.btn} href={githubUrl}>
                  Follow on GitHub
                </PressAnchor>
              </div>
              {/* eslint-disable-next-line @next/next/no-img-element -- static export, pre-sized still */}
              <img
                className={s.peek}
                src={stillUrl(480, 'webp')}
                width={480}
                height={480}
                alt=""
                loading="lazy"
                decoding="async"
              />
            </div>
          </Item>
        </Stagger>
      </div>
    </section>
  );
}
