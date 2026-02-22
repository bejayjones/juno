<script lang="ts">
	import { onMount } from 'svelte';
	import { reports, type ReportView, ApiError } from '$lib/api';

	let items = $state<ReportView[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			items = (await reports.list({ limit: 50 })) ?? [];
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Failed to load reports.';
		} finally {
			loading = false;
		}
	});

	function formatDate(unixSec: number): string {
		return new Date(unixSec * 1000).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: 'numeric',
			minute: '2-digit'
		});
	}

	const statusColor: Record<string, string> = {
		draft: 'text-amber-400 bg-amber-950 border-amber-800',
		finalized: 'text-emerald-400 bg-emerald-950 border-emerald-800'
	};

	const statusLabel: Record<string, string> = {
		draft: 'Draft',
		finalized: 'Finalized'
	};
</script>

<svelte:head>
	<title>Reports — Juno</title>
</svelte:head>

<div class="px-4 py-6">
	<div class="mb-4">
		<h2 class="text-2xl font-bold text-white">Reports</h2>
		<p class="mt-0.5 text-xs text-slate-500">Generated from completed inspections.</p>
	</div>

	{#if loading}
		<div class="space-y-3">
			{#each Array(3) as _}
				<div class="h-20 rounded-xl bg-slate-800 animate-pulse"></div>
			{/each}
		</div>
	{:else if error}
		<div class="rounded-xl border border-red-800 bg-red-950 p-6 text-center">
			<p class="text-sm text-red-400">{error}</p>
		</div>
	{:else if items.length === 0}
		<div class="rounded-xl border border-slate-700 bg-slate-800 p-8 text-center">
			<svg class="mx-auto mb-3 size-10 text-slate-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
				<path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
			</svg>
			<p class="text-slate-400 text-sm">No reports yet.</p>
			<p class="mt-1 text-xs text-slate-500">Complete an inspection, then use the appointment detail to generate a report.</p>
			<a
				href="/appointments"
				class="mt-3 inline-block rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white hover:bg-blue-500"
			>
				Go to Appointments
			</a>
		</div>
	{:else}
		<ul class="space-y-3">
			{#each items as report (report.id)}
				<li>
					<a
						href="/reports/{report.id}"
						class="block rounded-xl border border-slate-700 bg-slate-800 p-4
							transition-colors hover:bg-slate-750 active:bg-slate-700"
					>
						<div class="flex items-start justify-between gap-2">
							<div class="min-w-0 flex-1">
								<p class="text-sm font-semibold text-white">
									Inspection Report
								</p>
								<p class="mt-0.5 text-xs text-slate-400">{formatDate(report.created_at)}</p>
								{#if report.generated_at}
									<p class="mt-0.5 text-xs text-slate-500">PDF generated {formatDate(report.generated_at)}</p>
								{:else}
									<p class="mt-0.5 text-xs text-amber-500">PDF not yet generated</p>
								{/if}
							</div>
							<span class="inline-flex shrink-0 items-center rounded-full border px-2 py-0.5 text-xs font-medium {statusColor[report.status] ?? ''}">
								{statusLabel[report.status] ?? report.status}
							</span>
						</div>
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</div>
