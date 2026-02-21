<script lang="ts">
	import { onMount } from 'svelte';
	import { appointments, type AppointmentView, ApiError } from '$lib/api';

	let items = $state<AppointmentView[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			items = (await appointments.list()) ?? [];
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Failed to load appointments.';
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
</script>

<svelte:head>
	<title>Appointments — Juno</title>
</svelte:head>

<div class="px-4 py-6">
	<div class="mb-4 flex items-center justify-between">
		<h2 class="text-2xl font-bold text-white">Appointments</h2>
	</div>

	{#if loading}
		<div class="space-y-3">
			{#each Array(3) as _}
				<div class="h-24 rounded-xl bg-slate-800 animate-pulse"></div>
			{/each}
		</div>
	{:else if error}
		<p class="text-sm text-red-400">{error}</p>
	{:else if items.length === 0}
		<div class="rounded-xl border border-slate-700 bg-slate-800 p-8 text-center">
			<p class="text-slate-400 text-sm">No appointments yet.</p>
		</div>
	{:else}
		<ul class="space-y-3">
			{#each items as appt}
				<li>
					<a
						href="/appointments/{appt.id}"
						class="block rounded-xl border border-slate-700 bg-slate-800 p-4 hover:bg-slate-750 transition-colors"
					>
						<div class="flex items-start justify-between gap-2">
							<div class="min-w-0 flex-1">
								<p class="text-sm font-semibold text-white truncate">
									{appt.property.street}, {appt.property.city}
								</p>
								<p class="mt-0.5 text-xs text-slate-400">{formatDate(appt.scheduled_at)}</p>
								<p class="mt-0.5 text-xs text-slate-500">{appt.duration_min} min</p>
							</div>
							<span class="inline-flex shrink-0 items-center rounded-full border px-2 py-0.5 text-xs font-medium {statusColor[appt.status] ?? ''}">
								{appt.status}
							</span>
						</div>
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</div>
