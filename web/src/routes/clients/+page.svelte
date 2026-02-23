<script lang="ts">
	import { onMount } from 'svelte';
	import { clients, type ClientView, ApiError } from '$lib/api';
	import { cacheClients, getCachedClients } from '$lib/db';

	// ── Client list ──────────────────────────────────────────────────────────
	let allClients = $state<ClientView[]>([]);
	let search = $state('');
	let loading = $state(true);
	let loadError = $state('');

	const filtered = $derived(
		search.length < 1
			? allClients
			: allClients.filter((c) => {
					const q = search.toLowerCase();
					return (
						c.first_name.toLowerCase().includes(q) ||
						c.last_name.toLowerCase().includes(q) ||
						c.email.toLowerCase().includes(q)
					);
				})
	);

	onMount(async () => {
		try {
			const data = (await clients.list({ limit: 500 })) ?? [];
			allClients = data;
			await cacheClients(data);
		} catch {
			allClients = await getCachedClients().catch(() => []);
			if (allClients.length === 0) loadError = 'Failed to load clients.';
		} finally {
			loading = false;
		}
	});

	// ── Create / Edit modal ──────────────────────────────────────────────────
	let showModal = $state(false);
	let editingClient = $state<ClientView | null>(null);
	let firstName = $state('');
	let lastName = $state('');
	let email = $state('');
	let phone = $state('');
	let saving = $state(false);
	let modalError = $state('');

	function openCreate() {
		editingClient = null;
		firstName = '';
		lastName = '';
		email = '';
		phone = '';
		modalError = '';
		showModal = true;
	}

	function openEdit(c: ClientView) {
		editingClient = c;
		firstName = c.first_name;
		lastName = c.last_name;
		email = c.email;
		phone = c.phone;
		modalError = '';
		showModal = true;
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		if (!firstName.trim() || !lastName.trim()) {
			modalError = 'First and last name are required.';
			return;
		}
		modalError = '';
		saving = true;
		try {
			if (editingClient) {
				const updated = await clients.update(editingClient.id, {
					first_name: firstName.trim(),
					last_name: lastName.trim(),
					email: email.trim(),
					phone: phone.trim()
				});
				allClients = allClients.map((c) => (c.id === updated.id ? updated : c));
			} else {
				const created = await clients.create({
					first_name: firstName.trim(),
					last_name: lastName.trim(),
					email: email.trim(),
					phone: phone.trim()
				});
				allClients = [created, ...allClients];
			}
			await cacheClients(allClients);
			showModal = false;
		} catch (err) {
			modalError = err instanceof ApiError ? err.message : 'Failed to save client.';
		} finally {
			saving = false;
		}
	}

	// ── Delete ───────────────────────────────────────────────────────────────
	let deletingId = $state<string | null>(null);

	async function handleDelete(c: ClientView) {
		if (!confirm(`Delete ${c.first_name} ${c.last_name}?`)) return;
		deletingId = c.id;
		try {
			await clients.delete(c.id);
			allClients = allClients.filter((x) => x.id !== c.id);
			await cacheClients(allClients);
		} catch {
			// silently fail — client stays in list
		} finally {
			deletingId = null;
		}
	}
</script>

<svelte:head>
	<title>Clients — Juno</title>
</svelte:head>

<div class="px-4 py-6">
	<!-- Header -->
	<div class="mb-4 flex items-center justify-between">
		<h2 class="text-xl font-bold text-white">Clients</h2>
		<button
			onclick={openCreate}
			class="flex items-center gap-1.5 rounded-lg bg-blue-600 px-3 py-2 text-sm font-semibold text-white
				tap-target transition-colors hover:bg-blue-500 active:bg-blue-700"
		>
			<svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
			</svg>
			Add Client
		</button>
	</div>

	<!-- Search -->
	<div class="mb-4">
		<input
			type="text"
			bind:value={search}
			placeholder="Search by name or email…"
			class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
				placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
		/>
	</div>

	<!-- List -->
	{#if loading}
		<div class="space-y-3">
			{#each Array(5) as _}
				<div class="h-16 rounded-xl bg-slate-800 animate-pulse"></div>
			{/each}
		</div>
	{:else if loadError}
		<div class="rounded-xl border border-red-800 bg-red-950 p-6 text-center">
			<p class="text-sm text-red-400">{loadError}</p>
		</div>
	{:else if filtered.length === 0}
		<div class="rounded-xl border border-slate-700 bg-slate-800 p-8 text-center">
			{#if allClients.length === 0}
				<p class="text-sm text-slate-400">No clients yet.</p>
				<button
					onclick={openCreate}
					class="mt-3 text-sm font-medium text-blue-400 hover:text-blue-300"
				>
					Add your first client
				</button>
			{:else}
				<p class="text-sm text-slate-400">No clients match "{search}"</p>
			{/if}
		</div>
	{:else}
		<ul class="space-y-2">
			{#each filtered as c (c.id)}
				<li class="rounded-xl border border-slate-700 bg-slate-800 px-4 py-3">
					<div class="flex items-center justify-between">
						<div class="min-w-0 flex-1">
							<p class="truncate text-sm font-semibold text-white">
								{c.first_name} {c.last_name}
							</p>
							{#if c.email}
								<p class="truncate text-xs text-slate-400">{c.email}</p>
							{/if}
							{#if c.phone}
								<p class="truncate text-xs text-slate-500">{c.phone}</p>
							{/if}
						</div>
						<div class="ml-3 flex items-center gap-1">
							<button
								type="button"
								onclick={() => openEdit(c)}
								class="flex size-8 items-center justify-center rounded-lg text-slate-400 hover:bg-slate-700 hover:text-white"
								aria-label="Edit {c.first_name} {c.last_name}"
							>
								<svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
									<path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125" />
								</svg>
							</button>
							<button
								type="button"
								onclick={() => handleDelete(c)}
								disabled={deletingId === c.id}
								class="flex size-8 items-center justify-center rounded-lg text-slate-400 hover:bg-red-950 hover:text-red-400
									disabled:opacity-50"
								aria-label="Delete {c.first_name} {c.last_name}"
							>
								<svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
									<path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
								</svg>
							</button>
						</div>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<!-- Create / Edit modal -->
{#if showModal}
	<div
		class="fixed inset-0 z-50 flex items-end justify-center bg-black/60 sm:items-center"
		role="dialog"
		aria-modal="true"
		aria-label={editingClient ? 'Edit client' : 'Add new client'}
	>
		<div class="w-full max-w-lg rounded-t-2xl bg-slate-900 p-6 sm:rounded-2xl">
			<div class="mb-4 flex items-center justify-between">
				<h3 class="text-lg font-bold text-white">
					{editingClient ? 'Edit Client' : 'New Client'}
				</h3>
				<button
					type="button"
					onclick={() => (showModal = false)}
					class="flex size-8 items-center justify-center rounded-lg text-slate-400 hover:bg-slate-800 hover:text-white"
					aria-label="Close"
				>
					<svg class="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			{#if modalError}
				<div class="mb-4 rounded-lg border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-300">
					{modalError}
				</div>
			{/if}

			<form onsubmit={handleSave} class="space-y-4">
				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="modal-first-name" class="mb-1 block text-sm font-medium text-slate-300">
							First Name <span class="text-red-400">*</span>
						</label>
						<input
							id="modal-first-name"
							type="text"
							bind:value={firstName}
							required
							placeholder="First name"
							class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
								placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
						/>
					</div>
					<div>
						<label for="modal-last-name" class="mb-1 block text-sm font-medium text-slate-300">
							Last Name <span class="text-red-400">*</span>
						</label>
						<input
							id="modal-last-name"
							type="text"
							bind:value={lastName}
							required
							placeholder="Last name"
							class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
								placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
						/>
					</div>
				</div>

				<div>
					<label for="modal-email" class="mb-1 block text-sm font-medium text-slate-300">Email</label>
					<input
						id="modal-email"
						type="email"
						bind:value={email}
						placeholder="client@example.com"
						class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
							placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					/>
				</div>

				<div>
					<label for="modal-phone" class="mb-1 block text-sm font-medium text-slate-300">Phone</label>
					<input
						id="modal-phone"
						type="tel"
						bind:value={phone}
						placeholder="(555) 123-4567"
						class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
							placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					/>
				</div>

				<div class="flex gap-3 pt-2">
					<button
						type="button"
						onclick={() => (showModal = false)}
						class="flex-1 rounded-lg border border-slate-700 bg-slate-800 py-3 text-sm font-medium
							text-slate-400 tap-target transition-colors hover:bg-slate-700 hover:text-white"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={saving}
						class="flex-1 rounded-lg bg-blue-600 py-3 text-sm font-semibold text-white
							tap-target transition-colors hover:bg-blue-500 active:bg-blue-700
							disabled:cursor-not-allowed disabled:opacity-50"
					>
						{saving ? 'Saving…' : editingClient ? 'Save Changes' : 'Add Client'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
