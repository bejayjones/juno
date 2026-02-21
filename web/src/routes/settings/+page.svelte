<script lang="ts">
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';

	const inspector = $derived($authStore.inspector);

	function logout() {
		authStore.logout();
		goto('/login');
	}
</script>

<svelte:head>
	<title>Settings — Juno</title>
</svelte:head>

<div class="px-4 py-6">
	<h2 class="text-2xl font-bold text-white mb-6">Settings</h2>

	{#if inspector}
		<div class="mb-6 rounded-xl border border-slate-700 bg-slate-800 p-4">
			<p class="text-sm font-semibold text-white">{inspector.first_name} {inspector.last_name}</p>
			<p class="text-xs text-slate-400 mt-0.5">{inspector.email}</p>
			<p class="text-xs text-slate-500 mt-0.5 capitalize">{inspector.role}</p>
		</div>
	{/if}

	<button
		onclick={logout}
		class="w-full rounded-xl border border-red-800 bg-red-950 px-4 py-3 text-sm font-semibold
			text-red-400 tap-target hover:bg-red-900 transition-colors"
	>
		Sign out
	</button>
</div>
