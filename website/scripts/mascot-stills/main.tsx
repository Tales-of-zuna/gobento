/**
 * Page the stills generator loads: the R3F mascot in its static pose on a
 * transparent canvas, plus a helper that downsamples the canvas to a still.
 */
import { createRoot } from 'react-dom/client';
import { MascotCanvas, VIEWS, type MascotView } from '@/components/mascot/MascotCanvas';

declare global {
  interface Window {
    /** Set once the first frame is drawn. */
    mascotReady?: boolean;
    /** Returns the canvas scaled to `width`x`height` pixels as a data URL. */
    exportStill?: (width: number, height: number, type: 'image/png' | 'image/webp') => string;
  }
}

/** Halves the image repeatedly, then scales to the target size, for clean antialiasing. */
function downsample(source: HTMLCanvasElement, width: number, height: number): HTMLCanvasElement {
  let current = source;
  while (current.width / 2 >= width) {
    current = drawScaled(current, current.width / 2, current.height / 2);
  }
  return current.width === width ? current : drawScaled(current, width, height);
}

function drawScaled(source: HTMLCanvasElement, width: number, height: number): HTMLCanvasElement {
  const target = document.createElement('canvas');
  target.width = width;
  target.height = height;
  const context = target.getContext('2d');
  if (!context) throw new Error('2D canvas unavailable');
  context.imageSmoothingEnabled = true;
  context.imageSmoothingQuality = 'high';
  context.drawImage(source, 0, 0, width, height);
  return target;
}

window.exportStill = (width, height, type) => {
  const canvas = document.querySelector('canvas');
  if (!canvas) throw new Error('mascot canvas not found');
  return downsample(canvas, width, height).toDataURL(
    type,
    type === 'image/webp' ? 0.92 : undefined,
  );
};

const requested = new URLSearchParams(location.search).get('view') ?? 'hero';
const view: MascotView = requested in VIEWS ? (requested as MascotView) : 'hero';

const root = document.getElementById('root');
if (!root) throw new Error('#root missing');
createRoot(root).render(
  <MascotCanvas
    view={view}
    frameloop="demand"
    animate={false}
    dpr={2}
    preserveDrawingBuffer
    onReady={() => {
      window.mascotReady = true;
    }}
  />,
);
