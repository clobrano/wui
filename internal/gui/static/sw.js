'use strict';

/* wui service worker: makes the GUI installable and keeps the app shell
   available when the server is briefly unreachable. Task data always comes
   from the network; cached pages are only a fallback. */

const CACHE = 'wui-v1';
const SHELL = [
  '/static/styles.css',
  '/static/app.js',
  '/static/favicon.svg',
  '/static/icon.svg',
  '/static/icon-192.png',
  '/manifest.webmanifest',
];

const OFFLINE_HTML = `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0"><title>wui — offline</title>
<style>body{margin:0;height:100vh;display:flex;align-items:center;justify-content:center;
background:#1e1e2e;color:#cdd6f4;font-family:system-ui,sans-serif;text-align:center}
button{margin-top:1rem;padding:.4rem 1rem;border:none;border-radius:4px;background:#2a2a3d;color:#cdd6f4;cursor:pointer}</style>
</head><body><div><p>Can't reach the wui server.</p><button onclick="location.reload()">Retry</button></div></body></html>`;

self.addEventListener('install', event => {
  event.waitUntil(caches.open(CACHE).then(c => c.addAll(SHELL)).then(() => self.skipWaiting()));
});

self.addEventListener('activate', event => {
  event.waitUntil(
    caches.keys()
      .then(keys => Promise.all(keys.filter(k => k !== CACHE).map(k => caches.delete(k))))
      .then(() => self.clients.claim())
  );
});

self.addEventListener('fetch', event => {
  const req = event.request;
  const url = new URL(req.url);
  if (req.method !== 'GET' || url.origin !== self.location.origin) return;
  // Never serve task data from cache.
  if (url.pathname.startsWith('/api/')) return;

  if (req.mode === 'navigate') {
    // Network first; fall back to the last copy of the page, then an offline notice.
    event.respondWith(
      fetch(req)
        .then(res => {
          if (res.ok) {
            const copy = res.clone();
            caches.open(CACHE).then(c => c.put(req, copy));
          }
          return res;
        })
        .catch(() => caches.match(req).then(hit => hit ||
          new Response(OFFLINE_HTML, { headers: { 'Content-Type': 'text/html; charset=utf-8' } })))
    );
    return;
  }

  if (url.pathname.startsWith('/static/')) {
    // Stale-while-revalidate: instant from cache, refreshed in the background.
    event.respondWith(
      caches.open(CACHE).then(c => c.match(req).then(hit => {
        const fresh = fetch(req).then(res => {
          if (res.ok) c.put(req, res.clone());
          return res;
        });
        if (hit) {
          event.waitUntil(fresh.catch(() => {}));
          return hit;
        }
        return fresh;
      }))
    );
  }
});
