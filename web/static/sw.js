/**
 * Juno service worker — offline-capable app shell.
 *
 * Strategy:
 *   - App shell (HTML/CSS/JS): cache-first, falling back to network.
 *   - API calls (/api/**): network-first, no caching (data must be fresh).
 *   - Health endpoint: network-first, ignored.
 */

const CACHE = 'juno-shell-v1';

// Install: cache the app shell.
self.addEventListener('install', (event) => {
	event.waitUntil(
		caches.open(CACHE).then((cache) =>
			cache.addAll(['/', '/index.html', '/manifest.json'])
		)
	);
	self.skipWaiting();
});

// Activate: remove stale caches.
self.addEventListener('activate', (event) => {
	event.waitUntil(
		caches
			.keys()
			.then((keys) =>
				Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k)))
			)
	);
	self.clients.claim();
});

// Fetch: app shell = cache-first, API = network-first.
self.addEventListener('fetch', (event) => {
	const url = new URL(event.request.url);

	if (url.pathname.startsWith('/api/') || url.pathname === '/health') {
		// API: network only, no cache.
		return;
	}

	event.respondWith(
		caches.match(event.request).then(
			(cached) =>
				cached ||
				fetch(event.request).then((response) => {
					// Cache successful GET responses for the app shell.
					if (event.request.method === 'GET' && response.ok) {
						const clone = response.clone();
						caches.open(CACHE).then((cache) => cache.put(event.request, clone));
					}
					return response;
				})
		)
	);
});
