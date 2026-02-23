<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth, ApiError } from '$lib/api';
	import { authStore } from '$lib/stores/auth';

	let firstName = $state('');
	let lastName = $state('');
	let email = $state('');
	let password = $state('');
	let companyName = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleRegister(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const res = await auth.register({
				first_name: firstName,
				last_name: lastName,
				email,
				password,
				company_name: companyName.trim()
			});
			authStore.login(res.token, res.expires_at, res.inspector);
			goto('/');
		} catch (err) {
			if (err instanceof ApiError && err.status === 409) {
				error = 'An account with that email already exists.';
			} else if (err instanceof ApiError && err.status === 400) {
				error = err.message || 'Please check your details and try again.';
			} else {
				error = 'Something went wrong. Please try again.';
			}
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Create Account — Juno</title>
</svelte:head>

<div class="flex min-h-dvh flex-col items-center justify-center px-6 py-12">
	<div class="w-full max-w-sm">
		<!-- Logo / wordmark -->
		<div class="mb-8 text-center">
			<h1 class="text-3xl font-bold tracking-tight text-white">Juno</h1>
			<p class="mt-1 text-sm text-slate-400">Home inspection platform</p>
		</div>

		<form onsubmit={handleRegister} class="space-y-4">
			{#if error}
				<div class="rounded-lg bg-red-950 border border-red-800 px-4 py-3 text-sm text-red-300">
					{error}
				</div>
			{/if}

			<div class="flex gap-3">
				<div class="flex-1">
					<label for="first-name" class="block text-sm font-medium text-slate-300 mb-1">
						First name
					</label>
					<input
						id="first-name"
						type="text"
						bind:value={firstName}
						required
						autocomplete="given-name"
						class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
							placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
						placeholder="Jane"
					/>
				</div>
				<div class="flex-1">
					<label for="last-name" class="block text-sm font-medium text-slate-300 mb-1">
						Last name
					</label>
					<input
						id="last-name"
						type="text"
						bind:value={lastName}
						required
						autocomplete="family-name"
						class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
							placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
						placeholder="Smith"
					/>
				</div>
			</div>

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
					autocomplete="new-password"
					minlength={8}
					class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
						placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					placeholder="••••••••"
				/>
			</div>

			<div>
				<label for="company-name" class="block text-sm font-medium text-slate-300 mb-1">
					Company name
				</label>
				<input
					id="company-name"
					type="text"
					bind:value={companyName}
					required
					autocomplete="organization"
					class="w-full rounded-lg border border-slate-700 bg-slate-800 px-4 py-3 text-white
						placeholder-slate-500 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					placeholder="Acme Inspections LLC"
				/>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="w-full rounded-lg bg-blue-600 px-4 py-3 text-sm font-semibold text-white
					tap-target transition-colors hover:bg-blue-500 active:bg-blue-700
					disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{loading ? 'Creating account…' : 'Create account'}
			</button>
		</form>

		<p class="mt-6 text-center text-sm text-slate-500">
			Already have an account?
			<a href="/login" class="font-medium text-blue-400 hover:text-blue-300">Sign in</a>
		</p>
	</div>
</div>
