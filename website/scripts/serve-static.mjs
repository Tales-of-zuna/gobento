// Serves the static export the way GitHub Pages serves a project site:
// requests under /gobento/ map to files in out/, directories resolve to
// index.html (with a redirect when the trailing slash is missing), unknown
// paths get out/404.html with status 404, and nothing outside the prefix
// is served.
//
// Usage: node scripts/serve-static.mjs [port]
import { createReadStream } from 'node:fs';
import { stat } from 'node:fs/promises';
import { createServer } from 'node:http';
import { extname, join, resolve, sep } from 'node:path';

const prefix = '/gobento';
const root = resolve(import.meta.dirname, '..', 'out');
const port = Number(process.argv[2] ?? 4173);

const types = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json',
  '.txt': 'text/plain; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.ico': 'image/x-icon',
  '.woff2': 'font/woff2',
};

/** Returns the stats of path, or undefined when it does not exist. */
async function statOrUndefined(path) {
  try {
    return await stat(path);
  } catch {
    return undefined;
  }
}

async function send(req, res, status, file) {
  const { size } = await stat(file);
  res.writeHead(status, {
    'content-type': types[extname(file)] ?? 'application/octet-stream',
    'content-length': size,
  });
  if (req.method === 'HEAD') return res.end();
  createReadStream(file).pipe(res);
}

createServer(async (req, res) => {
  const { pathname } = new URL(req.url ?? '/', 'http://localhost');
  const notFound = () => send(req, res, 404, join(root, '404.html'));

  if (pathname !== prefix && !pathname.startsWith(`${prefix}/`)) return notFound();
  const file = resolve(root, `.${decodeURIComponent(pathname.slice(prefix.length))}`);
  if (file !== root && !file.startsWith(root + sep)) return notFound();

  const info = await statOrUndefined(file);
  if (info?.isFile()) return send(req, res, 200, file);
  if (info?.isDirectory()) {
    if (!pathname.endsWith('/')) {
      res.writeHead(301, { location: `${pathname}/` });
      return res.end();
    }
    const index = join(file, 'index.html');
    if ((await statOrUndefined(index))?.isFile()) return send(req, res, 200, index);
  }
  return notFound();
}).listen(port, '127.0.0.1', () => {
  console.log(`Serving ${root} at http://127.0.0.1:${port}${prefix}/`);
});
