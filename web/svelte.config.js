import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		// Produces a static build in web/build/ for go:embed.
		adapter: adapter({
			fallback: 'index.html' // SPA fallback — served for any path not found
		}),
		// All routes are rendered client-side (offline-capable SPA).
		prerender: { entries: [] }
	}
};

export default config;
