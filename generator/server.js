// E7 Gear Generator Static Server using Bun
import { join } from 'path';

const PORT = 8080;
const BASE_DIR = import.meta.dirname || '.';

console.log(`Starting E7 Gear Forge server on http://localhost:${PORT}...`);

Bun.serve({
  port: PORT,
  async fetch(req) {
    const url = new URL(req.url);
    let pathname = url.pathname;

    // Default to index.html
    if (pathname === '/' || pathname === '') {
      pathname = '/index.html';
    }

    const filePath = join(BASE_DIR, pathname);

    try {
      const file = Bun.file(filePath);
      const exists = await file.exists();

      if (!exists) {
        console.log(`[404] ${pathname} -> Not Found`);
        return new Response('404 Not Found', { status: 404 });
      }

      // Automatically sets content-type based on file extension (html, css, js)
      return new Response(file);
    } catch (e) {
      console.error(`[500] Error serving ${pathname}:`, e);
      return new Response('500 Internal Server Error', { status: 500 });
    }
  }
});
