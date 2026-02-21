/**
 * Auth store — persists JWT + inspector profile in localStorage.
 */
import { writable, derived } from 'svelte/store';
import type { InspectorView } from '$lib/api';

interface AuthState {
	token: string | null;
	expiresAt: number | null; // Unix timestamp
	inspector: InspectorView | null;
}

function createAuthStore() {
	// Hydrate from localStorage on first load.
	function load(): AuthState {
		if (typeof localStorage === 'undefined') return { token: null, expiresAt: null, inspector: null };
		try {
			const raw = localStorage.getItem('juno_auth');
			if (!raw) return { token: null, expiresAt: null, inspector: null };
			const parsed = JSON.parse(raw) as AuthState;
			// Treat expired tokens as logged out.
			if (parsed.expiresAt && Date.now() / 1000 > parsed.expiresAt) {
				localStorage.removeItem('juno_auth');
				localStorage.removeItem('juno_token');
				return { token: null, expiresAt: null, inspector: null };
			}
			return parsed;
		} catch {
			return { token: null, expiresAt: null, inspector: null };
		}
	}

	const { subscribe, set, update } = writable<AuthState>(load());

	function persist(state: AuthState) {
		if (typeof localStorage !== 'undefined') {
			localStorage.setItem('juno_auth', JSON.stringify(state));
			if (state.token) {
				localStorage.setItem('juno_token', state.token);
			} else {
				localStorage.removeItem('juno_token');
			}
		}
		set(state);
	}

	return {
		subscribe,
		login(token: string, expiresAt: number, inspector: InspectorView) {
			persist({ token, expiresAt, inspector });
		},
		updateInspector(inspector: InspectorView) {
			update((s) => {
				const next = { ...s, inspector };
				persist(next);
				return next;
			});
		},
		logout() {
			persist({ token: null, expiresAt: null, inspector: null });
		}
	};
}

export const authStore = createAuthStore();

/** True when the user has a valid, non-expired token. */
export const isAuthenticated = derived(
	authStore,
	($auth) => $auth.token !== null
);
