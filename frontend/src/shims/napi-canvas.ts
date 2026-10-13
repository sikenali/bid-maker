/**
 * Browser stub for `@napi-rs/canvas`.
 *
 * pdfjs-dist references it inside `NodeCanvasFactory`, which only ever runs
 * in Node.js. The browser build must not bundle the native `.node` binary,
 * so module resolution is aliased here (see vite.config.ts).
 */
function unavailable(): never {
  throw new Error('@napi-rs/canvas is not available in the browser');
}

export const createCanvas = unavailable as unknown as (w: number, h: number) => never;
export const loadImage = unavailable as unknown as (...args: unknown[]) => never;
export default { createCanvas, loadImage };
