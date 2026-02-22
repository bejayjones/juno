<script lang="ts">
	import { onMount } from 'svelte';
	import { appointments, type AppointmentView, ApiError } from '$lib/api';
	import { cacheAppointments, getCachedAppointments } from '$lib/db';

	type StatusFilter = 'all' | 'scheduled' | 'in_progress' | 'completed' | 'cancelled';

	let items = $state<AppointmentView[]>([]);
	let loading = $state(true);
	let error = $state('');
	let offline = $state(false);
	let filter = $state<StatusFilter>('all');

	const filtered = $derived(
		filter === 'all' ? items : items.filter((a) => a.status === filter)
	);

	onMount(async () => {
		try {
			const data = (await appointments.list()) ?? [];
			items = data;
			offline = false;
			await cacheAppointments(data);
		} catch {
			// Fall back to IndexedDB cache.
			try {
				items = await getCachedAppointments();
				offline = items.length > 0;
				if (!offline) error = 'No connection and no cached data.';
			} catch {
				error = 'Failed to load appointments.';
			}
		} finally {
			loading = false;
		}
	});

	function formatDate(unixSec: number): string {
		return new Date(unixSec * 1000).toLocaleDateString('en-US', {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
			hour: 'numeric',
			minute: '2-digit'
		});
	}

	const statusColor: Record<string, string> = {
		scheduled: 'text-blue-400 bg-blue-950 border-blue-800',
		in_progress: 'text-amber-400 bg-amber-950 border-amber-800',
		completed: 'text-emerald-400 bg-emerald-950 border-emerald-800',
		cancelled: 'text-slate-500 bg-slate-800 border-slate-700'
	};

	const statusLabel: Record<string, string> = {
		scheduled: 'Scheduled',
		in_progress: 'In Progress',
		completed: 'Completed',
		cancelled: 'Cancelled'
	};

	const filterTabs: { key: StatusFilter; label: string }[] = [
		{ key: 'all', label: 'All' },
		{ key: 'scheduled', label: 'Scheduled' },
		{ key: 'in_progress', label: 'In Progress' },
		{ key: 'completed', label: 'Done' }
	];
</script>

<svelte:head>
	<title>Appointments — Juno</title>
</svelte:head>

<div class="px-4 py-6">
	<!-- Header -->
	<div class="mb-4 flex items-center justify-between">
		<h2 class="text-2xl font-bold text-white">Appointments</h2>
		<a
			href="/appointments/new"
			class="inline-flex items-center gap-1.5 rounded-lg bg-blue-600 px-3 py-2 text-sm font-semibold
				text-white tap-target transition-colors hover:bg-blue-500 active:bg-blue-700"
		>
			<svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
			</svg>
			New
		</a>
	</div>

	<!-- Offline banner -->
	{#if offline}
		<div class="mb-3 flex items-center gap-2 rounded-lg border border-amber-800 bg-amber-950 px-3 py-2 text-xs text-amber-300">
			<svg class="size-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
			</svg>
			Showing cached data — you appear to be offline.
		</div>
	{/if}

	<!-- Filter tabs -->
	<div class="mb-4 flex gap-1 overflow-x-auto pb-1">
		{#each filterTabs as tab}
			<button
				onclick={() => (filter = tab.key)}
				class="shrink-0 rounded-full px-3 py-1.5 text-xs font-medium transition-colors
					{filter === tab.key
					? 'bg-blue-600 text-white'
					: 'bg-slate-800 text-slate-400 hover:text-slate-200'}"
			>
				{tab.label}
			</button>
		{/each}
	</div>

	{#if loading}
		<div class="space-y-3">
			{#each Array(3) as _}
				<div class="h-24 rounded-xl bg-slate-800 animate-pulse"></div>
			{/each}
		</div>
	{:else if error}
		<div class="rounded-xl border border-red-800 bg-red-950 p-6 text-center">
			<p class="text-sm text-red-400">{error}</p>
		</div>
	{:else if filtered.length === 0}
		<div class="rounded-xl border border-slate-700 bg-slate-800 p-8 text-center">
			<p class="text-slate-400 text-sm">
				{filter === 'all' ? 'No appointments yet.' : `No ${statusLabel[filter] ?? filter} appointments.`}
			</p>
			{#if filter === 'all'}
				<a
					href="/appointments/new"
					class="mt-3 inline-block rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white hover:bg-blue-500"
				>
					Schedule your first appointment
				</a>
			{/if}
		</div>
	{:else}
		<ul class="space-y-3">
			{#each filtered as appt (appt.id)}
				<li>
					<a
						href="/appointments/{appt.id}"
						class="block rounded-xl border border-slate-700 bg-slate-800 p-4 transition-colors hover:bg-slate-750 active:bg-slate-700"
					>
						<div class="flex items-start justify-between gap-2">
							<div class="min-w-0 flex-1">
								<p class="truncate text-sm font-semibold text-white">
									{appt.property.street}, {appt.property.city}
								</p>
								<p class="mt-0.5 text-xs text-slate-400">{formatDate(appt.scheduled_at)}</p>
								<p class="mt-0.5 text-xs text-slate-500">{appt.duration_min} min</p>
							</div>
							<span
								class="inline-flex shrink-0 items-center rounded-full border px-2 py-0.5 text-xs font-medium
									{statusColor[appt.status] ?? ''}"
							>
								{statusLabel[appt.status] ?? appt.status}
							</span>
						</div>
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</div>
