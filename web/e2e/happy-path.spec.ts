import { test, expect, type Page } from '@playwright/test';

const PASSWORD = 'TestPass123!';

// ─── Helpers ─────────────────────────────────────────────────────────────────

/** Fill an input by its label text. */
async function fillLabel(page: Page, label: string | RegExp, value: string) {
	await page.getByLabel(label).fill(value);
}

/** Register a new inspector and return their email (for later sign-in). */
async function registerUser(
	page: Page,
	firstName: string,
	lastName: string,
	company: string
): Promise<string> {
	const email = `${firstName.toLowerCase()}+${Date.now()}@juno-test.com`;
	await page.goto('/register');
	await fillLabel(page, /first name/i, firstName);
	await fillLabel(page, /last name/i, lastName);
	await fillLabel(page, /email address/i, email);
	await fillLabel(page, /password/i, PASSWORD);
	await fillLabel(page, /company name/i, company);
	await page.getByRole('button', { name: /create account/i }).click();
	await page.waitForURL('/');
	return email;
}

/** Create a client via API, returns the created client name. */
async function createClientViaAPI(page: Page): Promise<string> {
	// Extract JWT from localStorage to make an authenticated API call
	const token = await page.evaluate(() => {
		try {
			const raw = localStorage.getItem('juno_auth');
			if (!raw) return null;
			return JSON.parse(raw)?.token ?? null;
		} catch {
			return null;
		}
	});
	if (!token) throw new Error('No auth token found in localStorage');

	const resp = await page.request.post('/api/v1/clients', {
		headers: { Authorization: `Bearer ${token}` },
		data: {
			first_name: 'Alice',
			last_name: 'Homeowner',
			email: `alice+${Date.now()}@owner.test`,
			phone: '555-0100'
		}
	});
	expect(resp.status()).toBe(201);
	return 'Alice Homeowner';
}

// ─── Tests ───────────────────────────────────────────────────────────────────

test.describe('Happy path — deployed app', () => {
	test('register → appointment → inspection → report', async ({ page }) => {
		// ── STEP 1: Register ──────────────────────────────────────────────────
		await test.step('Register new inspector account', async () => {
			await registerUser(page, 'E2E', 'Inspector', 'Happy Path Inspections');
			await expect(page.getByText(/good day/i)).toBeVisible();
		});

		// ── STEP 2: Confirm authenticated nav is visible ──────────────────────
		await test.step('Authenticated bottom nav is visible', async () => {
			const nav = page.locator('nav').last(); // bottom nav
			await expect(nav.getByRole('link', { name: /appointments/i })).toBeVisible();
			await expect(nav.getByRole('link', { name: /reports/i })).toBeVisible();
		});

		// ── STEP 3: Create a client via API ───────────────────────────────────
		let clientName = '';
		await test.step('Create a client', async () => {
			clientName = await createClientViaAPI(page);
		});

		// ── STEP 4: Create an appointment ────────────────────────────────────
		await test.step('Create appointment', async () => {
			await page.goto('/appointments/new');
			await expect(page.getByRole('heading', { name: /new appointment/i })).toBeVisible();

			// Select the client
			const clientInput = page.getByPlaceholder(/search clients/i);
			await clientInput.click();
			await clientInput.fill('Alice');
			await page.waitForTimeout(500); // let dropdown render
			await page.getByRole('button', { name: /Alice Homeowner/i }).click();

			// Property address (inputs have placeholders, not labels)
			await page.getByPlaceholder('Street address').fill('123 Elm Street');
			await page.getByPlaceholder('City').fill('Springfield');
			await page.getByPlaceholder('State').fill('IL');
			await page.getByPlaceholder('ZIP code').fill('62701');

			// Date/time — already defaults to tomorrow 9 AM, just leave it

			await page.getByRole('button', { name: /schedule appointment/i }).click();

			// Should redirect to appointment detail
			await page.waitForURL(/\/appointments\/.+/);
			await expect(page.getByText(/123 Elm/i).first()).toBeVisible();
		});

		// ── STEP 5: Start inspection ───────────────────────────────────────────
		await test.step('Start inspection from appointment', async () => {
			const startBtn = page.getByRole('button', { name: /start inspection/i });
			await expect(startBtn).toBeVisible();

			// Intercept the POST /inspections response so we know the inspection ID
			const inspectionCreatedPromise = page.waitForResponse(
				(r) => r.url().includes('/api/v1/inspections') && r.request().method() === 'POST'
			);

			await startBtn.click();
			const inspResp = await inspectionCreatedPromise;
			const inspData = await inspResp.json();
			console.log('Inspection created:', JSON.stringify(inspData).slice(0, 200));

			await page.waitForURL(/\/inspections\/.+/);
			console.log('Current URL:', page.url());

			// Intercept the GET /inspections/{id} response
			const inspFetchPromise = page.waitForResponse(
				(r) => r.url().includes('/api/v1/inspections/') && r.request().method() === 'GET',
				{ timeout: 15_000 }
			);

			const inspFetchResp = await inspFetchPromise;
			console.log('GET inspection status:', inspFetchResp.status());
			if (!inspFetchResp.ok()) {
				const body = await inspFetchResp.text();
				throw new Error(`GET inspection failed (${inspFetchResp.status()}): ${body}`);
			}

			// Log any errors captured in the page
			const pageError = await page.evaluate(() => {
				// Try to get the Svelte error state from the DOM text
				return document.body.innerText.includes('not found') ? document.body.innerText : 'ok';
			});
			console.log('Page content check:', pageError.slice(0, 100));

			// Wait for the system tab bar to appear (it's a <nav> with system buttons in the inspection page)
			// The system tab buttons are in the inspection header nav, not the bottom nav
			await expect(page.locator('button[title="Inspected"]').first()).toBeVisible({ timeout: 15_000 });
		});

		// ── STEP 6: Mark all visible items Inspected ──────────────────────────
		await test.step('Mark all items on first system as Inspected', async () => {
			// Status buttons have title="Inspected" — wait for them to render
			const inspectedBtns = page.locator('button[title="Inspected"]');
			await expect(inspectedBtns.first()).toBeVisible({ timeout: 15_000 });
			const count = await inspectedBtns.count();
			expect(count).toBeGreaterThan(0);
			for (let i = 0; i < count; i++) {
				await inspectedBtns.nth(i).click();
				// Small delay to avoid overwhelming the server
				await page.waitForTimeout(200);
			}
		});

		// ── STEP 7: Complete the inspection ───────────────────────────────────
		await test.step('Complete the inspection', async () => {
			// "Complete" button is in the header
			const completeBtn = page.getByRole('button', { name: /^Complete$/i });
			await expect(completeBtn).toBeVisible();
			await completeBtn.click();

			// Modal appears — need to confirm completion even if not all items done
			// The modal may have a force-complete option or just list errors
			await page.waitForTimeout(500);

			// If there's a confirm button in the modal, click it
			const confirmBtn = page.getByRole('button', { name: /^complete$/i }).last();
			if (await confirmBtn.isVisible({ timeout: 2000 })) {
				await confirmBtn.click();
			}

			// After completion, the header should show "Completed" badge
			// (If validation errors appear, we may need to mark more items)
			await page.waitForTimeout(1000);
		});

		// ── STEP 8: Navigate to Reports ───────────────────────────────────────
		await test.step('Reports list shows the new report', async () => {
			await page.goto('/reports');
			// Wait for at least one report row to appear
			await expect(page.locator('a[href^="/reports/"]').first()).toBeVisible({ timeout: 20_000 });
		});

		// ── STEP 9: Open report and wait for PDF ──────────────────────────────
		await test.step('Open report detail and wait for PDF generation', async () => {
			await page.locator('a[href^="/reports/"]').first().click();
			await page.waitForURL(/\/reports\/.+/);

			// PDF generation can take up to 30s
			await expect(page.getByRole('link', { name: /download pdf/i })).toBeVisible({
				timeout: 30_000
			});
			await expect(page.getByText(/draft|finalized/i).first()).toBeVisible();
		});

		// ── STEP 10: Finalize the report ──────────────────────────────────────
		await test.step('Finalize the report', async () => {
			const finalizeBtn = page.getByRole('button', { name: /finalize report/i });
			if (await finalizeBtn.isVisible({ timeout: 3_000 })) {
				await finalizeBtn.click();
				await expect(page.getByRole('heading', { name: /finalize report/i })).toBeVisible();
				await page.getByRole('button', { name: /^finalize$/i }).click();
				await expect(page.getByText('Finalized')).toBeVisible({ timeout: 10_000 });
			}
		});

		// ── STEP 11: Queue a delivery ─────────────────────────────────────────
		await test.step('Queue report delivery to a recipient', async () => {
			const emailInput = page.getByPlaceholder(/client@example/i);
			if (await emailInput.isVisible({ timeout: 3_000 })) {
				await emailInput.fill('client@example.com');
				await page.getByRole('button', { name: /^send$/i }).click();
				await expect(page.getByText(/queued for delivery/i)).toBeVisible({ timeout: 10_000 });
			}
		});
	});

	// ─────────────────────────────────────────────────────────────────────────
	test('sign out and sign back in', async ({ page }) => {
		// Register a fresh user — self-contained, no dependency on test 1
		const email = await registerUser(page, 'Return', 'User', 'Return Inspections');
		await expect(page.getByText(/good day/i)).toBeVisible();

		// Sign out from Settings
		await page.goto('/settings');
		await page.waitForLoadState('networkidle');
		const signOutBtn = page.getByRole('button', { name: /sign out/i });
		await expect(signOutBtn).toBeVisible();
		await signOutBtn.click();

		// Should redirect to login
		await expect(page).toHaveURL(/\/login/);

		// Sign back in with the SAME credentials just registered
		await fillLabel(page, /email address/i, email);
		await fillLabel(page, /password/i, PASSWORD);
		await page.getByRole('button', { name: /sign in/i }).click();
		await expect(page).toHaveURL('/');
		await expect(page.getByText(/good day/i)).toBeVisible();
	});
});
