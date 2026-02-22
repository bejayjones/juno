<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth, ApiError } from '$lib/api';
	import { authStore } from '$lib/stores/auth';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleLogin(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const res = await auth.login(email, password);
			authStore.login(res.token, res.expires_at, res.inspector);
			goto('/');
		} catch (err) {
			if (err instanceof ApiError && err.status === 401) {
				error = 'Invalid email or password.';
			} else {
				error = 'Something went wrong. Please try again.';
			}
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Sign In — Juno</title>
</svelte:head>

<div class="flex min-h-dvh flex-col items-center justify-center px-6 py-12">
	<div class="w-full max-w-sm">
		<!-- Logo / wordmark -->
		<div class="mb-8 text-center">
			<h1 class="text-3xl font-bold tracking-tight text-white">Juno</h1>
			<p class="mt-1 text-sm text-slate-400">Home inspection platform</p>
		</div>

		<form onsubmit={handleLogin} class="space-y-4">
			{#if error}
				<div class="rounded-lg bg-red-950 border border-red-800 px-4 py-3 text-sm text-red-300">
					{error}
				</div>
			{/if}

			<div>
				<label for="email" class="block text-sm font-medium text-slate-300 mb-1">
					Email address
				</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					required
					autocomplete="email"
					class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
						placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					placeholder="you@example.com"
				/>
			</div>

			<div>
				<label for="password" class="block text-sm font-medium text-slate-300 mb-1">
					Password
				</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					autocomplete="current-password"
					class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
						placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					placeholder="••••••••"
				/>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="w-full rounded-lg bg-blue-600 px-4 py-3 text-sm font-semibold text-white
					tap-target transition-colors hover:bg-blue-500 active:bg-blue-700
					disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{loading ? 'Signing in…' : 'Sign in'}
			</button>
		</form>

		<p class="mt-6 text-center text-sm text-slate-500">
			Don't have an account?
			<a href="/register" class="font-medium text-blue-400 hover:text-blue-300">Create one</a>
		</p>
	</div>
</div>
