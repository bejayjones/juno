/**
 * Typed API client for the Juno backend.
 * Automatically attaches the JWT bearer token from localStorage.
 */

const BASE = '/api/v1';

// ─── Auth types ──────────────────────────────────────────────────────────────

export interface LicenseView {
	state: string;
	number: string;
}

export interface InspectorView {
	id: string;
	company_id: string;
	first_name: string;
	last_name: string;
	email: string;
	role: string;
	licenses: LicenseView[];
	created_at: number;
	updated_at: number;
}

export interface CompanyView {
	id: string;
	name: string;
	address: {
		street: string;
		city: string;
		state: string;
		zip: string;
		country: string;
	};
	phone: string;
	email: string;
	logo_storage_path?: string;
	created_at: number;
	updated_at: number;
}

export interface RegisterResponse {
	inspector: InspectorView;
	company: CompanyView;
	token: string;
	expires_at: number;
}

export interface LoginResponse {
	inspector: InspectorView;
	token: string;
	expires_at: number;
}

// ─── Appointment types ────────────────────────────────────────────────────────

export interface AppointmentView {
	id: string;
	inspector_id: string;
	client_id: string;
	property: {
		street: string;
		city: string;
		state: string;
		zip: string;
		country: string;
	};
	scheduled_at: number;
	duration_min: number;
	status: 'scheduled' | 'in_progress' | 'completed' | 'cancelled';
	notes: string;
	created_at: number;
	updated_at: number;
}

// ─── HTTP helpers ─────────────────────────────────────────────────────────────

function getToken(): string | null {
	if (typeof localStorage === 'undefined') return null;
	return localStorage.getItem('juno_token');
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const headers: Record<string, string> = {
		'Content-Type': 'application/json'
	};
	const token = getToken();
	if (token) headers['Authorization'] = `Bearer ${token}`;

	const res = await fetch(`${BASE}${path}`, {
		method,
		headers,
		body: body !== undefined ? JSON.stringify(body) : undefined
	});

	if (!res.ok) {
		const err = await res.json().catch(() => ({ error: res.statusText }));
		throw new ApiError(res.status, err.error ?? res.statusText);
	}

	if (res.status === 204) return undefined as T;
	return res.json() as Promise<T>;
}

export class ApiError extends Error {
	constructor(
		public readonly status: number,
		message: string
	) {
		super(message);
	}
}

const get = <T>(path: string) => request<T>('GET', path);
const post = <T>(path: string, body: unknown) => request<T>('POST', path, body);
const put = <T>(path: string, body: unknown) => request<T>('PUT', path, body);
const del = <T>(path: string) => request<T>('DELETE', path);

// ─── Auth API ─────────────────────────────────────────────────────────────────

export const auth = {
	register: (body: {
		first_name: string;
		last_name: string;
		email: string;
		password: string;
		company_name?: string;
		company_id?: string;
	}) => post<RegisterResponse>('/auth/register', body),

	login: (email: string, password: string) =>
		post<LoginResponse>('/auth/login', { email, password })
};

// ─── Inspector API ────────────────────────────────────────────────────────────

export const inspectors = {
	me: () => get<InspectorView>('/me'),
	updateMe: (body: { first_name: string; last_name: string; email: string }) =>
		put<InspectorView>('/me', body),
	setLicense: (state: string, number: string) =>
		put<InspectorView>(`/me/licenses/${state}`, { number })
};

// ─── Company API ──────────────────────────────────────────────────────────────

export const companies = {
	get: (id: string) => get<CompanyView>(`/companies/${id}`),
	update: (
		id: string,
		body: { name: string; street: string; city: string; state: string; zip: string; country: string; phone: string; email: string }
	) => put<CompanyView>(`/companies/${id}`, body)
};

// ─── Appointments API ─────────────────────────────────────────────────────────

export const appointments = {
	list: (params?: { status?: string; from?: string; to?: string; limit?: number }) => {
		const q = new URLSearchParams();
		if (params?.status) q.set('status', params.status);
		if (params?.from) q.set('from', params.from);
		if (params?.to) q.set('to', params.to);
		if (params?.limit) q.set('limit', String(params.limit));
		const qs = q.toString();
		return get<AppointmentView[]>(`/appointments${qs ? '?' + qs : ''}`);
	},
	get: (id: string) => get<AppointmentView>(`/appointments/${id}`),
	create: (body: {
		client_id: string;
		street: string;
		city: string;
		state: string;
		zip: string;
		country?: string;
		scheduled_at: number;
		duration_min?: number;
		notes?: string;
	}) => post<AppointmentView>('/appointments', body),
	update: (
		id: string,
		body: {
			street?: string;
			city?: string;
			state?: string;
			zip?: string;
			country?: string;
			scheduled_at?: number;
			duration_min?: number;
			notes?: string;
		}
	) => put<AppointmentView>(`/appointments/${id}`, body),
	cancel: (id: string) => del<void>(`/appointments/${id}`)
};
