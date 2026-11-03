/**
 * Page the stills generator loads: the R3F mascot in its static pose on a
 * transparent canvas, plus a helper that downsamples the canvas to a still.
 */
import { createRoot } from 'react-dom/client';
import { MascotCanvas } from '@/components/mascot/MascotCanvas';

declare global {
  interface Window {
    /** Set once the first frame is drawn. */
    mascotReady?: boolean;
    /** Returns the canvas as a `size`-pixel square data URL. */
    exportStill?: (size: number, type: 'image/png' | 'image/webp') => string;
  }
}

/** Halves the image repeatedly, then scales to `size`, for clean antialiasing. */
function downsample(source: HTMLCanvasElement, size: number): HTMLCanvasElement {
  let current = source;
  while (current.width / 2 >= size) {
    current = drawScaled(current, current.width / 2);
  }
  return current.width === size ? current : drawScaled(current, size);
}

function drawScaled(source: HTMLCanvasElement, size: number): HTMLCanvasElement {
  const target = document.createElement('canvas');
  target.width = target.height = size;
  const context = target.getContext('2d');
  if (!context) throw new Error('2D canvas unavailable');
  context.imageSmoothingEnabled = true;
  context.imageSmoothingQuality = 'high';
  context.drawImage(source, 0, 0, size, size);
  return target;
}

window.exportStill = (size, type) => {
  const canvas = document.querySelector('canvas');
  if (!canvas) throw new Error('mascot canvas not found');
  return downsample(canvas, size).toDataURL(type, type === 'image/webp' ? 0.92 : undefined);
};

const root = document.getElementById('root');
if (!root) throw new Error('#root missing');
createRoot(root).render(
  <MascotCanvas
    frameloop="demand"
    animate={false}
    dpr={2}
    preserveDrawingBuffer
    onReady={() => {
      window.mascotReady = true;
    }}
  />,
);
