import { serve } from "https://deno.land/std@0.172.0/http/server.ts";
import {
  createCA,
  dumpData,
  getHostConfig,
  getLighthouseConfig,
  loadData,
  signCert,
} from "./nebula.ts";

const port = +(Deno.env.get("PORT") ?? 0) || 4000;
const distDir = Deno.env.get('DIST_DIR') || 'dist';

const handlers: Record<string, (...args: any) => any> = {
  createCA,
  signCert,
  getLighthouseConfig,
  getHostConfig,
  loadData,
  dumpData,
};

let promiseFallback: Promise<Uint8Array>;

function loadFallback() {
  promiseFallback ||= Deno.readFile(`${distDir}/index.html`);
  return promiseFallback;
}

await serve(async (request: Request) => {
  const { pathname } = new URL(request.url);
  if (
    request.method === "POST" && pathname === "/api"
  ) {
    let body;
    let status = 200;
    try {
      const { command, args } = await request.json();
      const handle = handlers[command];
      const result = await handle(...args || []);
      body = { result };
    } catch (error) {
      status = 500;
      console.error(error);
      body = { error: error || "Unknown error" };
    }
    return new Response(JSON.stringify(body), {
      status,
      headers: {
        "content-type": "application/json",
      },
    });
  }
  try {
    const file = await Deno.readFile(`${distDir}${pathname}`);
    const headers: Record<string, string> = {};
    if (pathname.endsWith('.css')) {
      headers['content-type'] = 'text/css';
    } else if (pathname.endsWith('.js')) {
      headers['content-type'] = 'text/javascript';
    }
    return new Response(file, {
      headers,
    });
  } catch {
    // fallback
  }
  try {
    return new Response(
      await loadFallback(),
      {
        headers: {
          "content-type": "text/html",
        },
      },
    );
  } catch (err) {
    if (err.code === 'ENOENT') {
      return new Response('File not found', {
        status: 404,
      });
    }
    throw err;
  }
}, { port });
