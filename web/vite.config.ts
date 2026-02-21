import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		// In development, proxy API and health calls to the Go backend.
		proxy: {
			'/api': 'http://localhost:8080',
			'/health': 'http://localhost:8080'
		}
	}
});
