/**
 * Lightweight native IndexedDB wrapper for offline caching.
 * Stores appointments and clients locally so the app works without a network.
 */

import type { AppointmentView, ClientView } from './api';

const DB_NAME = 'juno';
const DB_VERSION = 1;

function openDB(): Promise<IDBDatabase> {
	return new Promise((resolve, reject) => {
		const req = indexedDB.open(DB_NAME, DB_VERSION);

		req.onupgradeneeded = (e) => {
			const db = (e.target as IDBOpenDBRequest).result;
			if (!db.objectStoreNames.contains('appointments')) {
				db.createObjectStore('appointments', { keyPath: 'id' });
			}
			if (!db.objectStoreNames.contains('clients')) {
				db.createObjectStore('clients', { keyPath: 'id' });
			}
		};

		req.onsuccess = () => resolve(req.result);
		req.onerror = () => reject(req.error);
	});
}

function putAll<T>(db: IDBDatabase, store: string, items: T[]): Promise<void> {
	return new Promise((resolve, reject) => {
		const tx = db.transaction(store, 'readwrite');
		const s = tx.objectStore(store);
		for (const item of items) s.put(item);
		tx.oncomplete = () => resolve();
		tx.onerror = () => reject(tx.error);
	});
}

function getAll<T>(db: IDBDatabase, store: string): Promise<T[]> {
	return new Promise((resolve, reject) => {
		const tx = db.transaction(store, 'readonly');
		const req = tx.objectStore(store).getAll();
		req.onsuccess = () => resolve(req.result as T[]);
		req.onerror = () => reject(req.error);
	});
}

function getOne<T>(db: IDBDatabase, store: string, key: string): Promise<T | undefined> {
	return new Promise((resolve, reject) => {
		const tx = db.transaction(store, 'readonly');
		const req = tx.objectStore(store).get(key);
		req.onsuccess = () => resolve(req.result as T | undefined);
		req.onerror = () => reject(req.error);
	});
}

// ── Appointments ──────────────────────────────────────────────────────────────

export async function cacheAppointments(items: AppointmentView[]): Promise<void> {
	const db = await openDB();
	await putAll(db, 'appointments', items);
}

export async function getCachedAppointments(): Promise<AppointmentView[]> {
	const db = await openDB();
	return getAll<AppointmentView>(db, 'appointments');
}

export async function getCachedAppointment(id: string): Promise<AppointmentView | undefined> {
	const db = await openDB();
	return getOne<AppointmentView>(db, 'appointments', id);
}

export async function cacheAppointment(item: AppointmentView): Promise<void> {
	const db = await openDB();
	await putAll(db, 'appointments', [item]);
}

// ── Clients ───────────────────────────────────────────────────────────────────

export async function cacheClients(items: ClientView[]): Promise<void> {
	const db = await openDB();
	await putAll(db, 'clients', items);
}

export async function getCachedClients(): Promise<ClientView[]> {
	const db = await openDB();
	return getAll<ClientView>(db, 'clients');
}
