# Design

## Code font display

Options considered for JetBrains Mono:

- **Preload it.** It removes the late swap but adds about 40 KB to every
  page's critical path, competing with the body font and the LCP text, and
  the layout comment notes code is rarely the first thing on screen.
- **`font-display: optional` (chosen).** The browser uses the font only if
  it is ready within the short block period, so there is never a late
  reflow. A first visit on a slow connection may show code in the fallback
  monospace font; later visits use the cached font. Display and body fonts
  keep `swap`, because their next/font fallbacks are metric-adjusted and
  caused no shift.

## Layout-shift assertion

`categories:performance` hides which metric regressed. An explicit
`cumulative-layout-shift` assertion (`maxNumericValue: 0.1`, median run, the
"good" threshold of Core Web Vitals) names it.

## Integration timeout

`-timeout 60m` leaves room for the slower macOS and Windows runners (a local
run takes about 11 minutes). The documented commands carry the same flag so
contributors do not hit the 10-minute default either. `make test` (short
tests) keeps the default.
