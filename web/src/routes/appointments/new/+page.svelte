<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { appointments, clients, type ClientView, ApiError } from '$lib/api';
	import { cacheClients, getCachedClients, cacheAppointment } from '$lib/db';

	// ── Client search ────────────────────────────────────────────────────────
	let allClients = $state<ClientView[]>([]);
	let clientSearch = $state('');
	let selectedClient = $state<ClientView | null>(null);
	let showClientDropdown = $state(false);

	const clientMatches = $derived(
		clientSearch.length < 1
			? allClients.slice(0, 8)
			: allClients.filter((c) => {
					const q = clientSearch.toLowerCase();
					return (
						c.first_name.toLowerCase().includes(q) ||
						c.last_name.toLowerCase().includes(q) ||
						c.email.toLowerCase().includes(q)
					);
				})
	);

	onMount(async () => {
		try {
			const data = (await clients.list({ limit: 200 })) ?? [];
			allClients = data;
			await cacheClients(data);
		} catch {
			allClients = await getCachedClients().catch(() => []);
		}
	});

	function selectClient(c: ClientView) {
		selectedClient = c;
		clientSearch = `${c.first_name} ${c.last_name}`;
		showClientDropdown = false;
	}

	// ── Form fields ──────────────────────────────────────────────────────────
	let street = $state('');
	let city = $state('');
	let state = $state('');
	let zip = $state('');

	// Default to tomorrow at 9 AM local time
	const tomorrow = new Date();
	tomorrow.setDate(tomorrow.getDate() + 1);
	tomorrow.setHours(9, 0, 0, 0);
	const pad = (n: number) => String(n).padStart(2, '0');
	const defaultDatetime =
		`${tomorrow.getFullYear()}-${pad(tomorrow.getMonth() + 1)}-${pad(tomorrow.getDate())}` +
		`T${pad(tomorrow.getHours())}:${pad(tomorrow.getMinutes())}`;

	let scheduledAt = $state(defaultDatetime);
	let durationMin = $state(120);
	let notes = $state('');

	// ── Submit ───────────────────────────────────────────────────────────────
	let submitting = $state(false);
	let error = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!selectedClient) {
			error = 'Please select a client.';
			return;
		}
		error = '';
		submitting = true;
		try {
			const scheduledUnix = Math.floor(new Date(scheduledAt).getTime() / 1000);
			const appt = await appointments.create({
				client_id: selectedClient.id,
				street,
				city,
				state,
				zip,
				country: 'US',
				scheduled_at: scheduledUnix,
				duration_min: durationMin,
				notes
			});
			await cacheAppointment(appt);
			goto(`/appointments/${appt.id}`);
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Failed to create appointment.';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>New Appointment — Juno</title>
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
		<h2 class="text-xl font-bold text-white">New Appointment</h2>
	</div>

	<form onsubmit={handleSubmit} class="space-y-5">
		{#if error}
			<div class="rounded-lg border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-300">
				{error}
			</div>
		{/if}

		<!-- Client picker -->
		<div class="relative">
			<label for="client-search" class="mb-1 block text-sm font-medium text-slate-300">
				Client <span class="text-red-400">*</span>
			</label>
			<input
				id="client-search"
				type="text"
				bind:value={clientSearch}
				onfocus={() => (showClientDropdown = true)}
				oninput={() => {
					showClientDropdown = true;
					selectedClient = null;
				}}
				placeholder="Search clients…"
				autocomplete="off"
				class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
					placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
			/>
			{#if showClientDropdown && clientMatches.length > 0}
				<ul
					class="absolute z-20 mt-1 max-h-56 w-full overflow-y-auto rounded-lg border border-slate-700 bg-slate-800 shadow-xl"
				>
					{#each clientMatches as c (c.id)}
						<li>
							<button
								type="button"
								onclick={() => selectClient(c)}
								class="w-full px-4 py-3 text-left text-sm text-white hover:bg-slate-700"
							>
								<span class="font-medium">{c.first_name} {c.last_name}</span>
								<span class="ml-2 text-xs text-slate-400">{c.email}</span>
							</button>
						</li>
					{/each}
				</ul>
			{/if}
			{#if showClientDropdown && clientSearch.length > 0 && clientMatches.length === 0}
				<div class="absolute z-20 mt-1 w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-sm text-slate-400">
					No clients found.
				</div>
			{/if}
		</div>

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

		<button
			type="submit"
			disabled={submitting}
			class="w-full rounded-lg bg-blue-600 px-4 py-3 text-sm font-semibold text-white
				tap-target transition-colors hover:bg-blue-500 active:bg-blue-700
				disabled:cursor-not-allowed disabled:opacity-50"
		>
			{submitting ? 'Scheduling…' : 'Schedule Appointment'}
		</button>
	</form>
</div>
