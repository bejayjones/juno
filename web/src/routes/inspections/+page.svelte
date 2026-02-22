<script lang="ts">
	import { onMount } from 'svelte';
	import { inspections, type InspectionView, ApiError } from '$lib/api';
	import { getCachedInspections } from '$lib/db';

	let items = $state<InspectionView[]>([]);
	let loading = $state(true);
	let error = $state('');
	let offline = $state(false);

	onMount(async () => {
		try {
			items = (await inspections.list()) ?? [];
		} catch {
			try {
				items = await getCachedInspections();
				offline = true;
			} catch {
				error = 'Failed to load inspections.';
			}
		} finally {
			loading = false;
		}
	});

	function formatDate(unixSec: number): string {
		return new Date(unixSec * 1000).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			hour: 'numeric',
			minute: '2-digit'
		});
	}

	const inProgress = $derived(items.filter((i) => i.status === 'in_progress'));
	const completed = $derived(items.filter((i) => i.status === 'completed'));
</script>

<svelte:head>
	<title>Inspections — Juno</title>
</svelte:head>

<div class="px-4 py-6">
	<div class="mb-4">
		<h2 class="text-2xl font-bold text-white">Inspections</h2>
	</div>

	{#if offline}
		<div class="mb-3 flex items-center gap-2 rounded-lg border border-amber-800 bg-amber-950 px-3 py-2 text-xs text-amber-300">
			<svg class="size-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
			</svg>
			Showing cached data — you appear to be offline.
		</div>
	{/if}

	{#if loading}
		<div class="space-y-3">
			{#each Array(2) as _}
				<div class="h-20 rounded-xl bg-slate-800 animate-pulse"></div>
			{/each}
		</div>
	{:else if error}
		<p class="text-sm text-red-400">{error}</p>
	{:else}
		{#if inProgress.length > 0}
			<section class="mb-6">
				<h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-slate-500">In Progress</h3>
				<ul class="space-y-3">
					{#each inProgress as insp (insp.id)}
						<li>
							<a
								href="/inspections/{insp.id}"
								class="flex items-center justify-between rounded-xl border border-amber-800 bg-amber-950/30 p-4
									transition-colors hover:bg-amber-950/50 active:bg-amber-950/70"
							>
								<div>
									<p class="text-sm font-semibold text-white">In-Progress Inspection</p>
									<p class="mt-0.5 text-xs text-amber-400">Started {formatDate(insp.started_at)}</p>
								</div>
								<svg class="size-5 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
									<path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5" />
								</svg>
							</a>
						</li>
					{/each}
				</ul>
			</section>
		{/if}

		{#if completed.length > 0}
			<section>
				<h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-slate-500">Completed</h3>
				<ul class="space-y-3">
					{#each completed as insp (insp.id)}
						<li>
							<a
								href="/inspections/{insp.id}"
								class="flex items-center justify-between rounded-xl border border-slate-700 bg-slate-800 p-4
									transition-colors hover:bg-slate-700"
							>
								<div>
									<p class="text-sm font-semibold text-white">Completed Inspection</p>
									<p class="mt-0.5 text-xs text-slate-400">
										{insp.completed_at ? formatDate(insp.completed_at) : formatDate(insp.started_at)}
									</p>
								</div>
								<svg class="size-5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
									<path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5" />
								</svg>
							</a>
						</li>
					{/each}
				</ul>
			</section>
		{/if}

		{#if items.length === 0}
			<div class="rounded-xl border border-slate-700 bg-slate-800 p-8 text-center">
				<p class="text-sm text-slate-400">No inspections yet.</p>
				<p class="mt-1 text-xs text-slate-500">Start an inspection from an appointment.</p>
				<a
					href="/appointments"
					class="mt-3 inline-block rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white hover:bg-blue-500"
				>
					Go to Appointments
				</a>
			</div>
		{/if}
	{/if}
</div>
