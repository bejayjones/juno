<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { reports, inspections, type ReportView, type DeliveryView, type DeficiencyView, ApiError } from '$lib/api';
	import { systemShortLabel } from '$lib/catalog';

	const id = $derived($page.params.id);

	// ── State ──────────────────────────────────────────────────────────────────
	let report = $state<ReportView | null>(null);
	let deficiencies = $state<DeficiencyView[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Finalize
	let finalizing = $state(false);
	let finalizeError = $state('');
	let showFinalizeModal = $state(false);

	// Delivery form
	let recipientEmail = $state('');
	let delivering = $state(false);
	let deliverError = $state('');
	let deliverSuccess = $state('');

	// Retry
	let retrying = $state(false);

	// ── Load ───────────────────────────────────────────────────────────────────
	onMount(async () => {
		try {
			const [r, d] = await Promise.all([
				reports.get(id),
				Promise.resolve([] as DeficiencyView[]) // loaded after we have inspection_id
			]);
			report = r;
			// Fetch deficiency summary for this inspection
			if (r.inspection_id) {
				try {
					deficiencies = (await inspections.summary(r.inspection_id)) ?? [];
				} catch {
					// Non-fatal: deficiency list may be unavailable
				}
			}
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Failed to load report.';
		} finally {
			loading = false;
		}
	});

	// ── Helpers ────────────────────────────────────────────────────────────────
	function formatDate(unixSec: number): string {
		return new Date(unixSec * 1000).toLocaleDateString('en-US', {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: 'numeric',
			minute: '2-digit'
		});
	}

	// Group deficiencies by system
	const deficienciesBySystem = $derived(() => {
		const map = new Map<string, DeficiencyView[]>();
		for (const d of deficiencies) {
			const existing = map.get(d.system_type) ?? [];
			map.set(d.system_type, [...existing, d]);
		}
		return map;
	});

	// ── Finalize ───────────────────────────────────────────────────────────────
	async function finalize() {
		if (!report) return;
		finalizing = true;
		finalizeError = '';
		try {
			report = await reports.finalize(id);
			showFinalizeModal = false;
		} catch (err) {
			finalizeError = err instanceof ApiError ? err.message : 'Failed to finalize report.';
		} finally {
			finalizing = false;
		}
	}

	// ── Deliver ────────────────────────────────────────────────────────────────
	async function sendDelivery(e: Event) {
		e.preventDefault();
		if (!recipientEmail.trim()) return;
		delivering = true;
		deliverError = '';
		deliverSuccess = '';
		try {
			const delivery = await reports.deliver(id, recipientEmail.trim());
			if (report) report.deliveries = [...report.deliveries, delivery];
			recipientEmail = '';
			deliverSuccess = 'Report queued for delivery.';
			setTimeout(() => (deliverSuccess = ''), 3000);
		} catch (err) {
			deliverError = err instanceof ApiError ? err.message : 'Failed to queue delivery.';
		} finally {
			delivering = false;
		}
	}

	// ── Retry ──────────────────────────────────────────────────────────────────
	async function retryFailed() {
		if (!report) return;
		retrying = true;
		try {
			report = await reports.retryDeliveries(id);
		} catch {
			// silently ignore
		} finally {
			retrying = false;
		}
	}

	const hasFailedDeliveries = $derived(
		report?.deliveries.some((d) => d.status === 'failed') ?? false
	);

	const deliveryStatusColor: Record<string, string> = {
		pending: 'text-amber-400 bg-amber-950 border-amber-800',
		succeeded: 'text-emerald-400 bg-emerald-950 border-emerald-800',
		failed: 'text-red-400 bg-red-950 border-red-800'
	};
</script>

<svelte:head>
	<title>Report — Juno</title>
</svelte:head>

<div class="px-4 py-6">
	<!-- Header -->
	<div class="mb-6 flex items-center gap-3">
		<a
			href="/reports"
			class="flex size-9 items-center justify-center rounded-lg border border-slate-700 bg-slate-800 text-slate-400 hover:text-white"
			aria-label="Back"
		>
			<svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
			</svg>
		</a>
		<h2 class="flex-1 text-xl font-bold text-white">Inspection Report</h2>
	</div>

	{#if loading}
		<div class="space-y-4">
			{#each Array(3) as _}
				<div class="h-24 rounded-xl bg-slate-800 animate-pulse"></div>
			{/each}
		</div>
	{:else if error}
		<div class="rounded-xl border border-red-800 bg-red-950 p-6 text-center">
			<p class="text-sm text-red-400">{error}</p>
			<a href="/reports" class="mt-3 inline-block text-sm text-blue-400 underline">Back to reports</a>
		</div>
	{:else if report}
		<!-- Status + Finalize -->
		<div class="mb-4 flex items-center justify-between gap-3">
			<span class="inline-flex items-center rounded-full border px-3 py-1 text-sm font-medium
				{report.status === 'finalized'
				? 'text-emerald-400 bg-emerald-950 border-emerald-800'
				: 'text-amber-400 bg-amber-950 border-amber-800'}">
				{report.status === 'finalized' ? 'Finalized' : 'Draft'}
			</span>

			{#if report.status === 'draft' && report.generated_at}
				<button
					onclick={() => { finalizeError = ''; showFinalizeModal = true; }}
					class="rounded-lg bg-emerald-700 px-3 py-1.5 text-xs font-semibold text-white
						hover:bg-emerald-600 active:bg-emerald-800"
				>
					Finalize Report
				</button>
			{/if}
		</div>

		<!-- PDF section -->
		<div class="mb-4 rounded-xl border border-slate-700 bg-slate-800 p-4">
			<p class="mb-2 text-xs font-semibold uppercase tracking-wider text-slate-500">PDF Report</p>
			{#if report.generated_at}
				<p class="mb-3 text-xs text-slate-400">Generated {formatDate(report.generated_at)}</p>
				<a
					href="/api/v1/reports/{report.id}/pdf"
					target="_blank"
					rel="noopener noreferrer"
					class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-semibold
						text-white hover:bg-blue-500 active:bg-blue-700"
				>
					<svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
					</svg>
					Download PDF
				</a>
			{:else}
				<div class="flex items-center gap-2 text-sm text-amber-400">
					<div class="size-4 animate-spin rounded-full border border-amber-700 border-t-amber-300"></div>
					Generating PDF…
				</div>
			{/if}
		</div>

		<!-- Deficiency summary -->
		{#if deficiencies.length > 0}
			<div class="mb-4 rounded-xl border border-slate-700 bg-slate-800">
				<div class="border-b border-slate-700 px-4 py-3">
					<p class="text-xs font-semibold uppercase tracking-wider text-slate-500">
						Deficiencies
						<span class="ml-1 inline-flex size-5 items-center justify-center rounded-full bg-red-900 text-[10px] font-bold text-red-300">
							{deficiencies.length}
						</span>
					</p>
				</div>

				<!-- Per-system counts -->
				<div class="flex flex-wrap gap-2 border-b border-slate-700 px-4 py-3">
					{@const grouped = deficienciesBySystem()}
					{#each [...grouped.entries()] as [systemType, items]}
						<a
							href="/inspections/{report.inspection_id}?system={systemType}"
							class="inline-flex items-center gap-1 rounded-full border border-red-800 bg-red-950 px-2 py-0.5 text-xs text-red-300 hover:bg-red-900"
						>
							{systemShortLabel[systemType] ?? systemType}
							<span class="font-bold">{items.length}</span>
						</a>
					{/each}
				</div>

				<!-- Full deficiency list -->
				<ul class="divide-y divide-slate-700">
					{#each deficiencies as d (d.finding_id)}
						<li>
							<a
								href="/inspections/{report.inspection_id}?system={d.system_type}"
								class="block px-4 py-3 transition-colors hover:bg-slate-700"
							>
								<p class="text-xs font-medium text-slate-400">
									{systemShortLabel[d.system_type] ?? d.system_type} › {d.item_label}
								</p>
								<p class="mt-0.5 text-sm text-slate-200">{d.narrative}</p>
							</a>
						</li>
					{/each}
				</ul>
			</div>
		{:else if deficiencies.length === 0 && !loading}
			<div class="mb-4 rounded-xl border border-slate-700 bg-slate-800 px-4 py-3 text-center">
				<p class="text-xs text-slate-500">No deficiencies recorded for this inspection.</p>
			</div>
		{/if}

		<!-- Delivery section -->
		<div class="rounded-xl border border-slate-700 bg-slate-800">
			<div class="border-b border-slate-700 px-4 py-3">
				<p class="text-xs font-semibold uppercase tracking-wider text-slate-500">Delivery</p>
			</div>

			<!-- Send form -->
			{#if report.generated_at}
				<form onsubmit={sendDelivery} class="border-b border-slate-700 px-4 py-4">
					<label class="mb-1 block text-xs font-medium text-slate-400">
						Send report to recipient
					</label>
					<div class="flex gap-2">
						<input
							type="email"
							bind:value={recipientEmail}
							required
							placeholder="client@example.com"
							class="min-w-0 flex-1 rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-white
								placeholder-slate-600 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
						/>
						<button
							type="submit"
							disabled={delivering || !recipientEmail.trim()}
							class="shrink-0 rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white
								hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-50"
						>
							{delivering ? '…' : 'Send'}
						</button>
					</div>
					{#if deliverError}
						<p class="mt-1 text-xs text-red-400">{deliverError}</p>
					{/if}
					{#if deliverSuccess}
						<p class="mt-1 text-xs text-emerald-400">{deliverSuccess}</p>
					{/if}
				</form>
			{:else}
				<p class="border-b border-slate-700 px-4 py-3 text-xs text-slate-500">
					Delivery available once PDF is generated.
				</p>
			{/if}

			<!-- Delivery history -->
			{#if report.deliveries.length > 0}
				<ul class="divide-y divide-slate-700">
					{#each report.deliveries as delivery (delivery.id)}
						<li class="px-4 py-3">
							<div class="flex items-start justify-between gap-2">
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm text-white">{delivery.recipient_email}</p>
									{#if delivery.sent_at}
										<p class="text-xs text-slate-400">Sent {formatDate(delivery.sent_at)}</p>
									{:else}
										<p class="text-xs text-slate-500">{delivery.attempts} attempt{delivery.attempts !== 1 ? 's' : ''}</p>
									{/if}
									{#if delivery.failure_reason}
										<p class="text-xs text-red-400">{delivery.failure_reason}</p>
									{/if}
								</div>
								<span class="inline-flex shrink-0 items-center rounded-full border px-2 py-0.5 text-xs font-medium {deliveryStatusColor[delivery.status] ?? ''}">
									{delivery.status}
								</span>
							</div>
						</li>
					{/each}
				</ul>

				{#if hasFailedDeliveries}
					<div class="border-t border-slate-700 px-4 py-3">
						<button
							onclick={retryFailed}
							disabled={retrying}
							class="text-xs font-medium text-amber-400 hover:text-amber-300 disabled:opacity-50"
						>
							{retrying ? 'Retrying…' : 'Retry failed deliveries'}
						</button>
					</div>
				{/if}
			{:else}
				<div class="px-4 py-4 text-center">
					<p class="text-xs text-slate-500">No deliveries yet.</p>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Finalize modal -->
{#if showFinalizeModal}
	<div
		class="fixed inset-0 z-50 flex items-end justify-center bg-black/60 sm:items-center"
		role="dialog"
		aria-modal="true"
	>
		<div class="w-full max-w-sm rounded-t-2xl border border-slate-700 bg-slate-900 p-6 sm:rounded-2xl">
			<h2 class="text-lg font-bold text-white">Finalize Report?</h2>
			<p class="mt-2 text-sm text-slate-400">
				Finalizing locks the report. The inspection can no longer be edited and the PDF will represent
				the final record.
			</p>

			{#if finalizeError}
				<div class="mt-3 rounded-lg border border-red-800 bg-red-950 px-3 py-2 text-sm text-red-400">
					{finalizeError}
				</div>
			{/if}

			<div class="mt-5 flex gap-3">
				<button
					onclick={() => (showFinalizeModal = false)}
					class="flex-1 rounded-xl border border-slate-700 bg-slate-800 py-3 text-sm font-medium
						text-slate-400 hover:bg-slate-700"
				>
					Cancel
				</button>
				<button
					onclick={finalize}
					disabled={finalizing}
					class="flex-1 rounded-xl bg-emerald-700 py-3 text-sm font-semibold text-white
						hover:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-50"
				>
					{finalizing ? 'Finalizing…' : 'Finalize'}
				</button>
			</div>
		</div>
	</div>
{/if}
