/**
 * Juno service worker — offline-capable app shell.
 *
 * Strategy:
 *   - Navigation requests (HTML): network-first, falling back to cache.
 *     This ensures deploys with new content-hashed chunks are picked up
 *     immediately; the cached shell is only used when offline.
 *   - Immutable assets (/_app/immutable/**): cache-first (hash in filename).
 *   - Other static assets: stale-while-revalidate.
 *   - API calls (/api/**): network only, no caching.
 */

const SHELL_CACHE = 'juno-shell-v2';
const ASSET_CACHE = 'juno-assets-v1';

// Install: cache the minimal app shell.
self.addEventListener('install', (event) => {
	event.waitUntil(
		caches.open(SHELL_CACHE).then((cache) =>
			cache.addAll(['/', '/index.html', '/manifest.json'])
		)
	);
	self.skipWaiting();
});

// Activate: remove stale caches.
self.addEventListener('activate', (event) => {
	const keep = new Set([SHELL_CACHE, ASSET_CACHE]);
	event.waitUntil(
		caches
			.keys()
			.then((keys) =>
				Promise.all(keys.filter((k) => !keep.has(k)).map((k) => caches.delete(k)))
			)
	);
	self.clients.claim();
});

// Fetch handler.
self.addEventListener('fetch', (event) => {
	const url = new URL(event.request.url);

	// API & health: network only, let the browser handle it.
	if (url.pathname.startsWith('/api/') || url.pathname === '/health') {
		return;
	}

	// Immutable hashed assets: cache-first (filename changes on every build).
	if (url.pathname.startsWith('/_app/immutable/')) {
		event.respondWith(
			caches.match(event.request).then(
				(cached) =>
					cached ||
					fetch(event.request).then((response) => {
						if (response.ok) {
							const clone = response.clone();
							caches.open(ASSET_CACHE).then((c) => c.put(event.request, clone));
						}
						return response;
					})
			)
		);
		return;
	}

	// Navigation requests (HTML pages): network-first so new deploys are
	// picked up immediately; fall back to cached shell when offline.
	if (event.request.mode === 'navigate') {
		event.respondWith(
			fetch(event.request)
				.then((response) => {
					if (response.ok) {
						const clone = response.clone();
						caches.open(SHELL_CACHE).then((c) => c.put(event.request, clone));
					}
					return response;
				})
				.catch(() => caches.match('/index.html').then((c) => c || caches.match('/')))
		);
		return;
	}

	// Other static assets (manifest, sw.js, etc.): stale-while-revalidate.
	event.respondWith(
		caches.match(event.request).then((cached) => {
			const fetchPromise = fetch(event.request).then((response) => {
				if (event.request.method === 'GET' && response.ok) {
					const clone = response.clone();
					caches.open(SHELL_CACHE).then((c) => c.put(event.request, clone));
				}
				return response;
			});
			return cached || fetchPromise;
		})
	);
});
