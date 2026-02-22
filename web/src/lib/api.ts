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

// ─── Client types ─────────────────────────────────────────────────────────────

export interface ClientView {
	id: string;
	company_id: string;
	first_name: string;
	last_name: string;
	email: string;
	phone: string;
	created_at: number;
	updated_at: number;
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

// ─── Inspection types ─────────────────────────────────────────────────────────

export interface HeaderView {
	weather: string;
	temperature_f: number;
	attendees: string[];
	year_built: number;
	structure_type: string;
}

export interface ProgressView {
	addressed: number;
	total: number;
}

export interface PhotoRefView {
	id: string;
	storage_path: string;
	mime_type: string;
	captured_at: number;
}

export interface FindingView {
	id: string;
	narrative: string;
	is_deficiency: boolean;
	photos: PhotoRefView[];
	created_at: number;
	updated_at: number;
}

export interface ItemView {
	id: string;
	item_key: string;
	label: string;
	status: '' | 'I' | 'NI' | 'NP' | 'D';
	not_inspected_reason: string;
	findings: FindingView[];
	updated_at: number;
}

export interface SystemSectionView {
	id: string;
	system_type: string;
	system_label: string;
	descriptions: Record<string, string>;
	items: ItemView[];
	inspector_notes: string;
	progress: ProgressView;
	updated_at: number;
}

export interface InspectionView {
	id: string;
	appointment_id: string;
	inspector_id: string;
	status: 'in_progress' | 'completed' | 'voided';
	header: HeaderView;
	systems: SystemSectionView[];
	started_at: number;
	completed_at?: number;
}

export interface DeficiencyView {
	system_type: string;
	system_label: string;
	item_key: string;
	item_label: string;
	finding_id: string;
	narrative: string;
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
		throw new ApiError(res.status, err.error ?? res.statusText, err.fields);
	}

	if (res.status === 204) return undefined as T;
	return res.json() as Promise<T>;
}

async function postMultipart<T>(path: string, formData: FormData): Promise<T> {
	const headers: Record<string, string> = {};
	const token = getToken();
	if (token) headers['Authorization'] = `Bearer ${token}`;
	// Do NOT set Content-Type — browser sets it with the multipart boundary.

	const res = await fetch(`${BASE}${path}`, {
		method: 'POST',
		headers,
		body: formData
	});

	if (!res.ok) {
		const err = await res.json().catch(() => ({ error: res.statusText }));
		throw new ApiError(res.status, err.error ?? res.statusText, err.fields);
	}

	return res.json() as Promise<T>;
}

export class ApiError extends Error {
	constructor(
		public readonly status: number,
		message: string,
		public readonly details?: string[]
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

// ─── Clients API ──────────────────────────────────────────────────────────────

export const clients = {
	list: (params?: { search?: string; limit?: number }) => {
		const q = new URLSearchParams();
		if (params?.search) q.set('search', params.search);
		if (params?.limit) q.set('limit', String(params.limit));
		const qs = q.toString();
		return get<ClientView[]>(`/clients${qs ? '?' + qs : ''}`);
	},
	get: (id: string) => get<ClientView>(`/clients/${id}`),
	create: (body: { first_name: string; last_name: string; email: string; phone?: string }) =>
		post<ClientView>('/clients', body),
	update: (id: string, body: { first_name?: string; last_name?: string; email?: string; phone?: string }) =>
		put<ClientView>(`/clients/${id}`, body),
	delete: (id: string) => del<void>(`/clients/${id}`)
};

// ─── Report types ─────────────────────────────────────────────────────────────

export interface DeliveryView {
	id: string;
	recipient_email: string;
	status: 'pending' | 'succeeded' | 'failed';
	attempts: number;
	sent_at?: number;
	failure_reason?: string;
	created_at: number;
	updated_at: number;
}

export interface ReportView {
	id: string;
	inspection_id: string;
	inspector_id: string;
	status: 'draft' | 'finalized';
	pdf_storage_path?: string;
	generated_at?: number;
	deliveries: DeliveryView[];
	created_at: number;
	updated_at: number;
}

// ─── Reports API ───────────────────────────────────────────────────────────────

export const reports = {
	generate: (inspectionId: string) =>
		post<ReportView>('/reports', { inspection_id: inspectionId }),

	list: (params?: { limit?: number; offset?: number }) => {
		const q = new URLSearchParams();
		if (params?.limit) q.set('limit', String(params.limit));
		if (params?.offset) q.set('offset', String(params.offset));
		const qs = q.toString();
		return get<ReportView[]>(`/reports${qs ? '?' + qs : ''}`);
	},

	get: (id: string) => get<ReportView>(`/reports/${id}`),
	finalize: (id: string) => put<ReportView>(`/reports/${id}/finalize`, {}),

	deliver: (id: string, recipientEmail: string) =>
		post<DeliveryView>(`/reports/${id}/deliver`, { recipient_email: recipientEmail }),

	retryDeliveries: (id: string) =>
		post<ReportView>(`/reports/${id}/deliveries/retry`, {})
};

// ─── Inspections API ──────────────────────────────────────────────────────────

export const inspections = {
	start: (appointmentId: string) =>
		post<InspectionView>('/inspections', { appointment_id: appointmentId }),
	get: (id: string) => get<InspectionView>(`/inspections/${id}`),
	list: () => get<InspectionView[]>('/inspections'),

	getSystem: (id: string, systemType: string) =>
		get<SystemSectionView>(`/inspections/${id}/systems/${systemType}`),

	setItemStatus: (
		id: string,
		systemType: string,
		itemKey: string,
		body: { status: string; reason?: string }
	) => put<SystemSectionView>(`/inspections/${id}/systems/${systemType}/items/${itemKey}/status`, body),

	setDescriptions: (id: string, systemType: string, descriptions: Record<string, string>) =>
		put<SystemSectionView>(`/inspections/${id}/systems/${systemType}/descriptions`, descriptions),

	addFinding: (
		id: string,
		systemType: string,
		itemKey: string,
		body: { narrative: string; is_deficiency: boolean }
	) => post<FindingView>(`/inspections/${id}/systems/${systemType}/items/${itemKey}/findings`, body),

	updateFinding: (
		id: string,
		systemType: string,
		itemKey: string,
		findingId: string,
		body: { narrative: string; is_deficiency: boolean }
	) =>
		put<FindingView>(
			`/inspections/${id}/systems/${systemType}/items/${itemKey}/findings/${findingId}`,
			body
		),

	deleteFinding: (id: string, systemType: string, itemKey: string, findingId: string) =>
		del<void>(`/inspections/${id}/systems/${systemType}/items/${itemKey}/findings/${findingId}`),

	addPhoto: (id: string, systemType: string, itemKey: string, findingId: string, file: File) => {
		const fd = new FormData();
		fd.append('finding_id', findingId);
		fd.append('photo', file);
		return postMultipart<PhotoRefView>(
			`/inspections/${id}/systems/${systemType}/items/${itemKey}/photos`,
			fd
		);
	},

	deletePhoto: (id: string, systemType: string, itemKey: string, photoId: string) =>
		del<void>(`/inspections/${id}/systems/${systemType}/items/${itemKey}/photos/${photoId}`),

	complete: (id: string) => post<InspectionView>(`/inspections/${id}/complete`, {}),

	summary: (id: string) => get<DeficiencyView[]>(`/inspections/${id}/summary`)
};
