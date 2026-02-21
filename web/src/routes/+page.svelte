<script lang="ts">
	import { authStore } from '$lib/stores/auth';

	const inspector = $derived($authStore.inspector);
</script>

<svelte:head>
	<title>Dashboard — Juno</title>
</svelte:head>

<div class="px-4 py-6">
	<div class="mb-6">
		<h2 class="text-2xl font-bold text-white">
			{#if inspector}
				Good day, {inspector.first_name}
			{:else}
				Dashboard
			{/if}
		</h2>
		<p class="mt-1 text-sm text-slate-400">
			{inspector?.role === 'owner' ? 'Company owner' : 'Inspector'}
		</p>
	</div>

	<!-- Quick action cards -->
	<div class="grid grid-cols-2 gap-3">
		<a
			href="/appointments"
			class="rounded-xl bg-slate-800 border border-slate-700 p-4 hover:bg-slate-750 transition-colors"
		>
			<div class="mb-2">
				<svg class="size-7 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" />
				</svg>
			</div>
			<p class="text-sm font-semibold text-white">Appointments</p>
			<p class="text-xs text-slate-400 mt-0.5">View schedule</p>
		</a>

		<a
			href="/reports"
			class="rounded-xl bg-slate-800 border border-slate-700 p-4 hover:bg-slate-750 transition-colors"
		>
			<div class="mb-2">
				<svg class="size-7 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
				</svg>
			</div>
			<p class="text-sm font-semibold text-white">Reports</p>
			<p class="text-xs text-slate-400 mt-0.5">Manage & send</p>
		</a>
	</div>
</div>
