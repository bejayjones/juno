<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { appointments, inspections, reports, type AppointmentView, ApiError } from '$lib/api';
	import { getCachedAppointment, cacheAppointment, getCachedInspections } from '$lib/db';

	const id = $derived($page.params.id);

	let appt = $state<AppointmentView | null>(null);
	let loading = $state(true);
	let error = $state('');
	let offline = $state(false);

	// Action state
	let starting = $state(false);
	let cancelling = $state(false);
	let generatingReport = $state(false);
	let actionError = $state('');

	// Inspection ID (found from cache for completed appointments)
	let inspectionId = $state<string | null>(null);

	onMount(async () => {
		try {
			const data = await appointments.get(id);
			appt = data;
			await cacheAppointment(data);
		} catch {
			const cached = await getCachedAppointment(id).catch(() => undefined);
			if (cached) {
				appt = cached;
				offline = true;
			} else {
				error = 'Appointment not found.';
			}
		} finally {
			loading = false;
			// For completed appointments, find the linked inspection from cache
			if (appt?.status === 'completed') {
				try {
					const cached = await getCachedInspections();
					const linked = cached.find((i) => i.appointment_id === appt!.id);
					if (linked) inspectionId = linked.id;
				} catch {
					// non-fatal
				}
			}
		}
	});

	async function startInspection() {
		if (!appt) return;
		starting = true;
		actionError = '';
		try {
			const inspection = await inspections.start(appt.id);
			goto(`/inspections/${inspection.id}`);
		} catch (err) {
			actionError = err instanceof ApiError ? err.message : 'Failed to start inspection.';
			starting = false;
		}
	}

	async function generateReport() {
		if (!inspectionId) return;
		generatingReport = true;
		actionError = '';
		try {
			const report = await reports.generate(inspectionId);
			goto(`/reports/${report.id}`);
		} catch (err) {
			if (err instanceof ApiError && err.status === 409) {
				// Report already exists — go to reports list
				goto('/reports');
			} else {
				actionError = err instanceof ApiError ? err.message : 'Failed to generate report.';
				generatingReport = false;
			}
		}
	}

	async function cancelAppointment() {
		if (!appt) return;
		cancelling = true;
		actionError = '';
		try {
			await appointments.cancel(appt.id);
			goto('/appointments');
		} catch (err) {
			actionError = err instanceof ApiError ? err.message : 'Failed to cancel appointment.';
			cancelling = false;
		}
	}

	function formatDate(unixSec: number): string {
		return new Date(unixSec * 1000).toLocaleDateString('en-US', {
			weekday: 'long',
			year: 'numeric',
			month: 'long',
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
</script>

<svelte:head>
	<title>Appointment — Juno</title>
</svelte:head>

<div class="px-4 py-6">
	<!-- Header -->
	<div class="mb-6 flex items-center gap-3">
		<a
			href="/appointments"
			class="flex size-9 items-center justify-center rounded-lg border border-slate-700 bg-slate-800 text-slate-400 hover:text-white"
			aria-label="Back"
		>
			<svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
			</svg>
		</a>
		<h2 class="flex-1 truncate text-xl font-bold text-white">Appointment</h2>
		{#if appt && appt.status === 'scheduled' && !offline}
			<a
				href="/appointments/{appt.id}/edit"
				class="flex size-9 items-center justify-center rounded-lg border border-slate-700 bg-slate-800 text-slate-400 hover:text-white"
				aria-label="Edit"
			>
				<svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125" />
				</svg>
			</a>
		{/if}
	</div>

	{#if offline}
		<div class="mb-4 flex items-center gap-2 rounded-lg border border-amber-800 bg-amber-950 px-3 py-2 text-xs text-amber-300">
			<svg class="size-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
			</svg>
			Showing cached data — you appear to be offline.
		</div>
	{/if}

	{#if loading}
		<div class="space-y-4">
			<div class="h-32 rounded-xl bg-slate-800 animate-pulse"></div>
			<div class="h-16 rounded-xl bg-slate-800 animate-pulse"></div>
		</div>
	{:else if error}
		<div class="rounded-xl border border-red-800 bg-red-950 p-6 text-center">
			<p class="text-sm text-red-400">{error}</p>
			<a href="/appointments" class="mt-3 inline-block text-sm text-blue-400 underline">
				Back to appointments
			</a>
		</div>
	{:else if appt}
		<!-- Status badge -->
		<div class="mb-4">
			<span class="inline-flex items-center rounded-full border px-3 py-1 text-sm font-medium {statusColor[appt.status] ?? ''}">
				{statusLabel[appt.status] ?? appt.status}
			</span>
		</div>

		<!-- Detail card -->
		<div class="mb-4 rounded-xl border border-slate-700 bg-slate-800 divide-y divide-slate-700">
			<div class="px-4 py-3">
				<p class="text-xs font-medium uppercase tracking-wider text-slate-500">Property</p>
				<p class="mt-1 text-sm font-semibold text-white">{appt.property.street}</p>
				<p class="text-sm text-slate-400">
					{appt.property.city}, {appt.property.state} {appt.property.zip}
				</p>
			</div>
			<div class="px-4 py-3">
				<p class="text-xs font-medium uppercase tracking-wider text-slate-500">Date & Time</p>
				<p class="mt-1 text-sm text-white">{formatDate(appt.scheduled_at)}</p>
			</div>
			<div class="px-4 py-3">
				<p class="text-xs font-medium uppercase tracking-wider text-slate-500">Duration</p>
				<p class="mt-1 text-sm text-white">{appt.duration_min} minutes</p>
			</div>
			{#if appt.notes}
				<div class="px-4 py-3">
					<p class="text-xs font-medium uppercase tracking-wider text-slate-500">Notes</p>
					<p class="mt-1 text-sm text-slate-300 whitespace-pre-wrap">{appt.notes}</p>
				</div>
			{/if}
		</div>

		<!-- Action error -->
		{#if actionError}
			<div class="mb-4 rounded-lg border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-300">
				{actionError}
			</div>
		{/if}

		<!-- Primary CTA: Start Inspection -->
		{#if appt.status === 'scheduled' && !offline}
			<button
				onclick={startInspection}
				disabled={starting}
				class="mb-3 w-full rounded-xl bg-blue-600 py-4 text-base font-bold text-white
					tap-target transition-colors hover:bg-blue-500 active:bg-blue-700
					disabled:cursor-not-allowed disabled:opacity-50"
			>
				{starting ? 'Starting…' : 'Start Inspection'}
			</button>
		{/if}

		{#if appt.status === 'in_progress'}
			<a
				href="/inspections"
				class="mb-3 flex w-full items-center justify-center rounded-xl bg-amber-600 py-4 text-base font-bold text-white
					tap-target transition-colors hover:bg-amber-500"
			>
				Continue Inspection
			</a>
		{/if}

		<!-- Generate Report (completed appointments) -->
		{#if appt.status === 'completed' && !offline}
			{#if inspectionId}
				<button
					onclick={generateReport}
					disabled={generatingReport}
					class="mb-3 w-full rounded-xl bg-blue-600 py-4 text-base font-bold text-white
						tap-target transition-colors hover:bg-blue-500 active:bg-blue-700
						disabled:cursor-not-allowed disabled:opacity-50"
				>
					{generatingReport ? 'Generating…' : 'Generate Report'}
				</button>
			{/if}
			<a
				href="/reports"
				class="mb-3 flex w-full items-center justify-center rounded-xl border border-slate-700
					bg-slate-800 py-3 text-sm font-medium text-slate-400 tap-target hover:bg-slate-700 hover:text-white"
			>
				View All Reports
			</a>
		{/if}

		<!-- Cancel -->
		{#if (appt.status === 'scheduled' || appt.status === 'in_progress') && !offline}
			<button
				onclick={cancelAppointment}
				disabled={cancelling}
				class="w-full rounded-xl border border-slate-700 bg-slate-800 py-3 text-sm font-medium
					text-slate-400 tap-target transition-colors hover:bg-slate-700 hover:text-white
					disabled:cursor-not-allowed disabled:opacity-50"
			>
				{cancelling ? 'Cancelling…' : 'Cancel Appointment'}
			</button>
		{/if}
	{/if}
</div>
