<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { appointments, type AppointmentView, ApiError } from '$lib/api';
	import { getCachedAppointment, cacheAppointment } from '$lib/db';

	const id = $derived($page.params.id);

	let appt = $state<AppointmentView | null>(null);
	let loading = $state(true);
	let error = $state('');

	// Form fields (populated once appt loads)
	let street = $state('');
	let city = $state('');
	let state = $state('');
	let zip = $state('');
	let scheduledAt = $state('');
	let durationMin = $state(120);
	let notes = $state('');

	let submitting = $state(false);
	let submitError = $state('');

	onMount(async () => {
		try {
			appt = await appointments.get(id);
		} catch {
			const cached = await getCachedAppointment(id).catch(() => undefined);
			if (cached) appt = cached;
			else error = 'Appointment not found.';
		} finally {
			loading = false;
			if (appt) populateForm(appt);
		}
	});

	function populateForm(a: AppointmentView) {
		street = a.property.street;
		city = a.property.city;
		state = a.property.state;
		zip = a.property.zip;
		durationMin = a.duration_min;
		notes = a.notes ?? '';

		// Convert unix timestamp → datetime-local string (YYYY-MM-DDTHH:mm)
		const d = new Date(a.scheduled_at * 1000);
		const pad = (n: number) => String(n).padStart(2, '0');
		scheduledAt =
			`${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}` +
			`T${pad(d.getHours())}:${pad(d.getMinutes())}`;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		submitError = '';
		submitting = true;
		try {
			const scheduledUnix = Math.floor(new Date(scheduledAt).getTime() / 1000);
			const updated = await appointments.update(id, {
				street,
				city,
				state,
				zip,
				scheduled_at: scheduledUnix,
				duration_min: durationMin,
				notes
			});
			await cacheAppointment(updated);
			goto(`/appointments/${id}`);
		} catch (err) {
			submitError = err instanceof ApiError ? err.message : 'Failed to update appointment.';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Edit Appointment — Juno</title>
</svelte:head>

<div class="px-4 py-6">
	<!-- Header -->
	<div class="mb-6 flex items-center gap-3">
		<a
			href="/appointments/{id}"
			class="flex size-9 items-center justify-center rounded-lg border border-slate-700 bg-slate-800 text-slate-400 hover:text-white"
			aria-label="Back"
		>
			<svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
			</svg>
		</a>
		<h2 class="text-xl font-bold text-white">Edit Appointment</h2>
	</div>

	{#if loading}
		<div class="space-y-4">
			{#each Array(4) as _}
				<div class="h-12 rounded-lg bg-slate-800 animate-pulse"></div>
			{/each}
		</div>
	{:else if error}
		<div class="rounded-xl border border-red-800 bg-red-950 p-6 text-center">
			<p class="text-sm text-red-400">{error}</p>
			<a href="/appointments" class="mt-3 inline-block text-sm text-blue-400 underline">
				Back to appointments
			</a>
		</div>
	{:else if appt}
		<form onsubmit={handleSubmit} class="space-y-5">
			{#if submitError}
				<div class="rounded-lg border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-300">
					{submitError}
				</div>
			{/if}

			<!-- Property address -->
			<fieldset>
				<legend class="mb-2 text-sm font-medium text-slate-300">Property Address</legend>
				<div class="space-y-3">
					<input
						type="text"
						bind:value={street}
						required
						placeholder="Street address"
						class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
							placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					/>
					<div class="grid grid-cols-2 gap-3">
						<input
							type="text"
							bind:value={city}
							required
							placeholder="City"
							class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
								placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
						/>
						<input
							type="text"
							bind:value={state}
							required
							maxlength={2}
							placeholder="State"
							class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
								placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
						/>
					</div>
					<input
						type="text"
						bind:value={zip}
						required
						inputmode="numeric"
						placeholder="ZIP code"
						class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
							placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					/>
				</div>
			</fieldset>

			<!-- Date & time -->
			<div>
				<label for="scheduled-at" class="mb-1 block text-sm font-medium text-slate-300">
					Date & Time <span class="text-red-400">*</span>
				</label>
				<input
					id="scheduled-at"
					type="datetime-local"
					bind:value={scheduledAt}
					required
					class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
						focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500
						[color-scheme:dark]"
				/>
			</div>

			<!-- Duration -->
			<div>
				<label for="duration" class="mb-1 block text-sm font-medium text-slate-300">
					Duration (minutes)
				</label>
				<input
					id="duration"
					type="number"
					bind:value={durationMin}
					min={30}
					max={480}
					step={15}
					class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
						focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
				/>
			</div>

			<!-- Notes -->
			<div>
				<label for="notes" class="mb-1 block text-sm font-medium text-slate-300">Notes</label>
				<textarea
					id="notes"
					bind:value={notes}
					rows={3}
					placeholder="Access instructions, gate codes, etc."
					class="w-full resize-none rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
						placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
				></textarea>
			</div>

			<div class="flex gap-3">
				<a
					href="/appointments/{id}"
					class="flex-1 rounded-xl border border-slate-700 bg-slate-800 py-3 text-center text-sm
						font-medium text-slate-400 tap-target transition-colors hover:bg-slate-700 hover:text-white"
				>
					Cancel
				</a>
				<button
					type="submit"
					disabled={submitting}
					class="flex-1 rounded-xl bg-blue-600 py-3 text-sm font-semibold text-white
						tap-target transition-colors hover:bg-blue-500 active:bg-blue-700
						disabled:cursor-not-allowed disabled:opacity-50"
				>
					{submitting ? 'Saving…' : 'Save Changes'}
				</button>
			</div>
		</form>
	{/if}
</div>
