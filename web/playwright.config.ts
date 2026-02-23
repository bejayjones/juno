import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: './e2e',
	timeout: 60_000,
	expect: { timeout: 10_000 },
	fullyParallel: false,
	retries: 0,
	reporter: 'list',
	use: {
		baseURL: 'https://juno-production-d439.up.railway.app',
		// Run headless in CI; set HEADED=1 to see the browser locally
		headless: process.env.HEADED !== '1',
		screenshot: 'only-on-failure',
		video: 'retain-on-failure',
		trace: 'retain-on-failure'
	},
	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] }
		}
	]
});
