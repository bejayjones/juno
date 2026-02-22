<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		inspections,
		type InspectionView,
		type SystemSectionView,
		type ItemView,
		type FindingView,
		ApiError
	} from '$lib/api';
	import { cacheInspection, getCachedInspection } from '$lib/db';
	import { systemDescriptions, systemShortLabel } from '$lib/catalog';

	const id = $derived($page.params.id);

	// ── Core state ─────────────────────────────────────────────────────────────
	let insp = $state<InspectionView | null>(null);
	let loading = $state(true);
	let error = $state('');
	let offline = $state(false);

	// ── Navigation — initialise from ?system= query param for deep-links ───────
	let activeSystem = $state($page.url.searchParams.get('system') ?? 'roof');
	let expandedItemKey = $state<string | null>(null);

	// ── Description drafts (keyed by systemType → {descKey → value}) ──────────
	let descDrafts = $state<Record<string, Record<string, string>>>({});
	let descSaving = $state<Set<string>>(new Set());

	// ── Status save state ──────────────────────────────────────────────────────
	let statusSaving = $state<Set<string>>(new Set()); // "systemType:itemKey"
	let statusErrors = $state<Map<string, string>>(new Map());

	// ── Finding form ───────────────────────────────────────────────────────────
	type FindingForm = {
		systemType: string;
		itemKey: string;
		findingId: string | null; // null = new
		narrative: string;
		isDeficiency: boolean;
		saving: boolean;
		error: string;
	};
	let findingForm = $state<FindingForm | null>(null);

	// ── Photo upload ───────────────────────────────────────────────────────────
	let photoUploading = $state<Set<string>>(new Set()); // findingId
	let photoError = $state('');

	// ── Complete modal ─────────────────────────────────────────────────────────
	let showCompleteModal = $state(false);
	let completing = $state(false);
	let completeErrors = $state<string[]>([]);

	// ── Derived ────────────────────────────────────────────────────────────────
	const activeSystemData = $derived(
		insp?.systems.find((s) => s.system_type === activeSystem) ?? null
	);

	const totalProgress = $derived(() => {
		if (!insp) return { addressed: 0, total: 0 };
		let a = 0,
			t = 0;
		for (const s of insp.systems) {
			a += s.progress.addressed;
			t += s.progress.total;
		}
		return { addressed: a, total: t };
	});

	const isReadOnly = $derived(insp?.status !== 'in_progress');

	// ── Load ───────────────────────────────────────────────────────────────────
	onMount(async () => {
		try {
			insp = await inspections.get(id);
			await cacheInspection(insp);
		} catch {
			const cached = await getCachedInspection(id).catch(() => undefined);
			if (cached) {
				insp = cached;
				offline = true;
			} else {
				error = 'Inspection not found.';
			}
		} finally {
			loading = false;
		}
	});

	// Initialize description drafts when inspection loads.
	$effect(() => {
		if (insp && Object.keys(descDrafts).length === 0) {
			const init: Record<string, Record<string, string>> = {};
			for (const sys of insp.systems) {
				init[sys.system_type] = { ...sys.descriptions };
			}
			descDrafts = init;
		}
	});

	// ── Helpers ────────────────────────────────────────────────────────────────
	function isAddressed(item: ItemView): boolean {
		if (!item.status) return false;
		if (item.status === 'NI') return !!item.not_inspected_reason;
		return true;
	}

	function findSystem(systemType: string): SystemSectionView | undefined {
		return insp?.systems.find((s) => s.system_type === systemType);
	}

	function replaceSystem(updated: SystemSectionView) {
		if (!insp) return;
		const idx = insp.systems.findIndex((s) => s.system_type === updated.system_type);
		if (idx >= 0) insp.systems[idx] = updated;
	}

	// ── Status ─────────────────────────────────────────────────────────────────
	async function setStatus(systemType: string, itemKey: string, status: string, reason?: string) {
		if (!insp || isReadOnly) return;
		const key = `${systemType}:${itemKey}`;

		// Optimistic update
		const sys = findSystem(systemType);
		if (!sys) return;
		const item = sys.items.find((i) => i.item_key === itemKey);
		if (!item) return;

		const prevStatus = item.status;
		const prevReason = item.not_inspected_reason;
		const wasAddressed = isAddressed(item);

		item.status = status as ItemView['status'];
		item.not_inspected_reason = status === 'NI' ? (reason ?? '') : '';

		const nowAddressed = isAddressed(item);
		if (!wasAddressed && nowAddressed) sys.progress.addressed++;
		if (wasAddressed && !nowAddressed) sys.progress.addressed--;

		statusSaving = new Set([...statusSaving, key]);
		statusErrors.delete(key);
		statusErrors = new Map(statusErrors);

		try {
			const updated = await inspections.setItemStatus(id, systemType, itemKey, {
				status,
				reason: reason ?? ''
			});
			replaceSystem(updated);
			// Re-sync description drafts for this system with server data
			if (!descDrafts[updated.system_type]) {
				descDrafts[updated.system_type] = { ...updated.descriptions };
			}
		} catch (err) {
			// Revert
			item.status = prevStatus;
			item.not_inspected_reason = prevReason;
			if (!wasAddressed && nowAddressed) sys.progress.addressed--;
			if (wasAddressed && !nowAddressed) sys.progress.addressed++;
			statusErrors.set(
				key,
				err instanceof ApiError ? err.message : 'Save failed — tap to retry'
			);
			statusErrors = new Map(statusErrors);
		} finally {
			statusSaving.delete(key);
			statusSaving = new Set(statusSaving);
		}
	}

	// ── NI reason blur ─────────────────────────────────────────────────────────
	function onReasonBlur(systemType: string, itemKey: string, reason: string) {
		const sys = findSystem(systemType);
		const item = sys?.items.find((i) => i.item_key === itemKey);
		if (!item || item.status !== 'NI') return;
		setStatus(systemType, itemKey, 'NI', reason);
	}

	// ── Descriptions ───────────────────────────────────────────────────────────
	async function saveDescriptions(systemType: string) {
		if (!insp || isReadOnly) return;
		const drafts = descDrafts[systemType];
		if (!drafts) return;

		descSaving = new Set([...descSaving, systemType]);
		try {
			const updated = await inspections.setDescriptions(id, systemType, drafts);
			replaceSystem(updated);
		} catch {
			// Silently ignore — user can retry by blurring again
		} finally {
			descSaving.delete(systemType);
			descSaving = new Set(descSaving);
		}
	}

	// ── Findings ───────────────────────────────────────────────────────────────
	function openFindingForm(systemType: string, itemKey: string, existing?: FindingView) {
		findingForm = {
			systemType,
			itemKey,
			findingId: existing?.id ?? null,
			narrative: existing?.narrative ?? '',
			isDeficiency: existing?.is_deficiency ?? false,
			saving: false,
			error: ''
		};
	}

	async function submitFinding() {
		if (!findingForm || !insp) return;
		const { systemType, itemKey, findingId, narrative, isDeficiency } = findingForm;
		findingForm.saving = true;
		findingForm.error = '';

		try {
			if (findingId) {
				const updated = await inspections.updateFinding(id, systemType, itemKey, findingId, {
					narrative,
					is_deficiency: isDeficiency
				});
				const sys = findSystem(systemType);
				const item = sys?.items.find((i) => i.item_key === itemKey);
				if (item) {
					const idx = item.findings.findIndex((f) => f.id === findingId);
					if (idx >= 0) item.findings[idx] = updated;
				}
			} else {
				const created = await inspections.addFinding(id, systemType, itemKey, {
					narrative,
					is_deficiency: isDeficiency
				});
				const sys = findSystem(systemType);
				const item = sys?.items.find((i) => i.item_key === itemKey);
				if (item) item.findings = [...item.findings, created];
			}
			findingForm = null;
		} catch (err) {
			if (findingForm) {
				findingForm.saving = false;
				findingForm.error = err instanceof ApiError ? err.message : 'Save failed.';
			}
		}
	}

	async function deleteFinding(systemType: string, itemKey: string, findingId: string) {
		if (!insp || isReadOnly) return;
		const sys = findSystem(systemType);
		const item = sys?.items.find((i) => i.item_key === itemKey);
		if (!item) return;

		const prev = [...item.findings];
		item.findings = item.findings.filter((f) => f.id !== findingId);
		try {
			await inspections.deleteFinding(id, systemType, itemKey, findingId);
		} catch {
			item.findings = prev;
		}
	}

	// ── Photos ─────────────────────────────────────────────────────────────────
	async function uploadPhoto(
		systemType: string,
		itemKey: string,
		findingId: string,
		e: Event
	) {
		if (isReadOnly) return;
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;

		photoUploading = new Set([...photoUploading, findingId]);
		photoError = '';

		try {
			const ref = await inspections.addPhoto(id, systemType, itemKey, findingId, file);
			const sys = findSystem(systemType);
			const item = sys?.items.find((i) => i.item_key === itemKey);
			const finding = item?.findings.find((f) => f.id === findingId);
			if (finding) finding.photos = [...finding.photos, ref];
		} catch (err) {
			photoError = err instanceof ApiError ? err.message : 'Photo upload failed.';
		} finally {
			photoUploading.delete(findingId);
			photoUploading = new Set(photoUploading);
			// Reset file input
			(e.target as HTMLInputElement).value = '';
		}
	}

	async function removePhoto(systemType: string, itemKey: string, photoId: string) {
		if (!insp || isReadOnly) return;
		const sys = findSystem(systemType);
		const item = sys?.items.find((i) => i.item_key === itemKey);
		if (!item) return;

		for (const finding of item.findings) {
			const idx = finding.photos.findIndex((p) => p.id === photoId);
			if (idx >= 0) {
				const prev = [...finding.photos];
				finding.photos = finding.photos.filter((p) => p.id !== photoId);
				try {
					await inspections.deletePhoto(id, systemType, itemKey, photoId);
				} catch {
					finding.photos = prev;
				}
				return;
			}
		}
	}

	// ── Complete ───────────────────────────────────────────────────────────────
	async function tryComplete() {
		if (!insp) return;
		completing = true;
		completeErrors = [];

		try {
			const updated = await inspections.complete(id);
			insp = updated;
			await cacheInspection(updated);
			showCompleteModal = false;
			goto(`/appointments/${insp.appointment_id}`);
		} catch (err) {
			if (err instanceof ApiError && err.status === 422) {
				completeErrors = err.details ?? [err.message];
			} else {
				completeErrors = [err instanceof ApiError ? err.message : 'Completion failed.'];
			}
		} finally {
			completing = false;
		}
	}

	// ── Status button helpers ──────────────────────────────────────────────────
	const statusConfig = {
		I: { label: 'I', title: 'Inspected', active: 'bg-emerald-600 text-white', ring: 'ring-emerald-500' },
		NI: { label: 'NI', title: 'Not Inspected', active: 'bg-amber-600 text-white', ring: 'ring-amber-500' },
		NP: { label: 'NP', title: 'Not Present', active: 'bg-slate-600 text-white', ring: 'ring-slate-400' },
		D: { label: 'D', title: 'Deficient', active: 'bg-red-600 text-white', ring: 'ring-red-500' }
	} as const;

	type StatusKey = keyof typeof statusConfig;
</script>

<svelte:head>
	<title>Inspection — Juno</title>
</svelte:head>

{#if loading}
	<div class="flex h-dvh items-center justify-center">
		<div class="h-8 w-8 animate-spin rounded-full border-2 border-slate-600 border-t-blue-400"></div>
	</div>
{:else if error}
	<div class="px-4 py-12 text-center">
		<p class="text-sm text-red-400">{error}</p>
		<a href="/inspections" class="mt-3 inline-block text-sm text-blue-400 underline">
			Back to inspections
		</a>
	</div>
{:else if insp}
	<!-- ── Layout wrapper ─────────────────────────────────────────────────────── -->
	<div class="flex h-dvh flex-col overflow-hidden">

		<!-- ── Top bar ──────────────────────────────────────────────────────────── -->
		<header class="flex shrink-0 items-center gap-3 border-b border-slate-800 bg-slate-900 px-4 py-3">
			<a
				href="/appointments/{insp.appointment_id}"
				class="flex size-8 shrink-0 items-center justify-center rounded-lg border border-slate-700 bg-slate-800 text-slate-400 hover:text-white"
				aria-label="Back"
			>
				<svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
				</svg>
			</a>

			<div class="min-w-0 flex-1">
				<p class="truncate text-sm font-semibold text-white">Inspection</p>
				{#if !isReadOnly}
					{@const prog = totalProgress()}
					<p class="text-xs text-slate-400">{prog.addressed} / {prog.total} items addressed</p>
				{:else}
					<span class="inline-flex items-center rounded-full border border-emerald-800 bg-emerald-950 px-2 py-0.5 text-xs text-emerald-400">
						Completed
					</span>
				{/if}
			</div>

			{#if !isReadOnly}
				<button
					onclick={() => { completeErrors = []; showCompleteModal = true; }}
					class="shrink-0 rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-semibold text-white
						hover:bg-blue-500 active:bg-blue-700"
				>
					Complete
				</button>
			{/if}
		</header>

		{#if offline}
			<div class="shrink-0 flex items-center gap-2 border-b border-amber-800 bg-amber-950/40 px-4 py-1.5 text-xs text-amber-300">
				<svg class="size-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
				</svg>
				Offline — cached data shown, edits unavailable
			</div>
		{/if}

		<!-- ── System tab bar ────────────────────────────────────────────────────── -->
		<nav class="shrink-0 overflow-x-auto border-b border-slate-800 bg-slate-900">
			<div class="flex min-w-max">
				{#each insp.systems as sys (sys.system_type)}
					{@const allDone = sys.progress.addressed === sys.progress.total}
					<button
						onclick={() => { activeSystem = sys.system_type; expandedItemKey = null; findingForm = null; }}
						class="relative flex flex-col items-center gap-0.5 px-3 py-2.5 text-xs transition-colors
							{activeSystem === sys.system_type
							? 'text-blue-400 after:absolute after:bottom-0 after:inset-x-0 after:h-0.5 after:bg-blue-500'
							: 'text-slate-500 hover:text-slate-300'}"
					>
						<span class="font-medium whitespace-nowrap">{systemShortLabel[sys.system_type] ?? sys.system_label}</span>
						<span class="text-[10px] {allDone ? 'text-emerald-400' : activeSystem === sys.system_type ? 'text-blue-300' : 'text-slate-600'}">
							{sys.progress.addressed}/{sys.progress.total}
						</span>
					</button>
				{/each}
			</div>
		</nav>

		<!-- ── System content (scrollable) ─────────────────────────────────────── -->
		<main class="flex-1 overflow-y-auto">
			{#if activeSystemData}
				{@const descFields = systemDescriptions[activeSystemData.system_type] ?? []}
				<div class="space-y-0">

					<!-- Required descriptions -->
					{#if descFields.length > 0}
						<section class="border-b border-slate-800 bg-slate-900 px-4 py-4">
							<p class="mb-3 text-xs font-semibold uppercase tracking-wider text-slate-500">
								Required Descriptions
								{#if descSaving.has(activeSystemData.system_type)}
									<span class="ml-1 text-blue-400">saving…</span>
								{/if}
							</p>
							<div class="space-y-3">
								{#each descFields as field}
									<div>
										<label for="desc-{field.key}" class="mb-1 block text-xs font-medium text-slate-400">{field.label}</label>
										<input
											id="desc-{field.key}"
											type="text"
											disabled={isReadOnly || offline}
											value={descDrafts[activeSystemData.system_type]?.[field.key] ?? ''}
											oninput={(e) => {
												if (!descDrafts[activeSystemData.system_type]) {
													descDrafts[activeSystemData.system_type] = {};
												}
												descDrafts[activeSystemData.system_type][field.key] =
													(e.target as HTMLInputElement).value;
											}}
											onblur={() => saveDescriptions(activeSystemData.system_type)}
											placeholder="Enter value…"
											class="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white
												placeholder-slate-600 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500
												disabled:opacity-50"
										/>
									</div>
								{/each}
							</div>
						</section>
					{/if}

					<!-- Items -->
					<ul class="divide-y divide-slate-800">
						{#each activeSystemData.items as item (item.item_key)}
							{@const itemKey = `${activeSystemData.system_type}:${item.item_key}`}
							{@const isExpanded = expandedItemKey === itemKey}
							{@const isSaving = statusSaving.has(itemKey)}
							{@const saveErr = statusErrors.get(itemKey)}

							<li class="bg-slate-900">
								<!-- Item row -->
								<button
									onclick={() => (expandedItemKey = isExpanded ? null : itemKey)}
									class="flex w-full items-center gap-3 px-4 py-3 text-left"
								>
									<!-- Expand chevron -->
									<svg
										class="size-4 shrink-0 text-slate-600 transition-transform {isExpanded ? 'rotate-90' : ''}"
										fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
									>
										<path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5" />
									</svg>

									<span class="flex-1 text-sm {isAddressed(item) ? 'text-white' : 'text-slate-400'}">
										{item.label}
									</span>

									{#if isSaving}
										<div class="size-4 animate-spin rounded-full border border-slate-600 border-t-blue-400"></div>
									{:else if item.findings.length > 0}
										<span class="text-xs text-amber-400">{item.findings.length} finding{item.findings.length !== 1 ? 's' : ''}</span>
									{/if}
								</button>

								<!-- Status buttons (always visible) -->
								<div class="flex gap-1.5 px-4 pb-3">
									{#each Object.entries(statusConfig) as [key, cfg]}
										<button
											onclick={(e) => { e.stopPropagation(); if (!isReadOnly && !offline) setStatus(activeSystemData.system_type, item.item_key, key); }}
											disabled={isReadOnly || offline}
											title={cfg.title}
											class="flex-1 rounded-lg border py-2 text-xs font-bold transition-colors
												{item.status === key
												? `${cfg.active} border-transparent`
												: 'border-slate-700 bg-slate-800 text-slate-500 hover:text-slate-300'}
												disabled:cursor-default"
										>
											{cfg.label}
										</button>
									{/each}
								</div>

								{#if saveErr}
									<p class="px-4 pb-2 text-xs text-red-400">{saveErr}</p>
								{/if}

								<!-- Expanded detail -->
								{#if isExpanded}
									<div class="border-t border-slate-800 bg-slate-950 px-4 py-4 space-y-4">

										<!-- NI reason -->
										{#if item.status === 'NI'}
											<div>
												<label for="ni-reason-{item.item_key}" class="mb-1 block text-xs font-medium text-slate-400">
													Reason not inspected <span class="text-red-400">*</span>
												</label>
												<textarea
													id="ni-reason-{item.item_key}"
													rows={2}
													disabled={isReadOnly || offline}
													value={item.not_inspected_reason}
													oninput={(e) => {
														item.not_inspected_reason = (e.target as HTMLTextAreaElement).value;
													}}
													onblur={(e) => onReasonBlur(activeSystemData.system_type, item.item_key, (e.target as HTMLTextAreaElement).value)}
													placeholder="Explain why this item was not inspected…"
													class="w-full resize-none rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm
														text-white placeholder-slate-600 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500
														disabled:opacity-50"
												></textarea>
											</div>
										{/if}

										<!-- Findings -->
										{#if item.findings.length > 0}
											<div class="space-y-3">
												{#each item.findings as finding (finding.id)}
													<div class="rounded-lg border {finding.is_deficiency ? 'border-red-800 bg-red-950/20' : 'border-slate-700 bg-slate-800/50'} p-3">
														<div class="flex items-start justify-between gap-2">
															<div class="flex-1">
																{#if finding.is_deficiency}
																	<span class="mb-1 inline-flex items-center rounded-full border border-red-700 bg-red-950 px-1.5 py-0.5 text-[10px] font-semibold text-red-400">
																		DEFICIENCY
																	</span>
																{/if}
																<p class="text-sm text-slate-200">{finding.narrative}</p>
															</div>

															{#if !isReadOnly && !offline}
																<div class="flex shrink-0 gap-1">
																	<button
																		onclick={() => openFindingForm(activeSystemData.system_type, item.item_key, finding)}
																		class="rounded p-1 text-slate-500 hover:text-slate-300"
																		aria-label="Edit finding"
																	>
																		<svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
																			<path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Z" />
																		</svg>
																	</button>
																	<button
																		onclick={() => deleteFinding(activeSystemData.system_type, item.item_key, finding.id)}
																		class="rounded p-1 text-slate-500 hover:text-red-400"
																		aria-label="Delete finding"
																	>
																		<svg class="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
																			<path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
																		</svg>
																	</button>
																</div>
															{/if}
														</div>

														<!-- Photos -->
														{#if finding.photos.length > 0}
															<div class="mt-2 flex flex-wrap gap-2">
																{#each finding.photos as photo (photo.id)}
																	<div class="relative">
																		<img
																			src="/api/v1/photos/{photo.id}"
																			alt=""
																			class="h-16 w-16 rounded-lg object-cover border border-slate-700"
																			loading="lazy"
																		/>
																		{#if !isReadOnly && !offline}
																			<button
																				onclick={() => removePhoto(activeSystemData.system_type, item.item_key, photo.id)}
																				class="absolute -right-1.5 -top-1.5 flex size-5 items-center justify-center
																					rounded-full bg-red-600 text-white hover:bg-red-500"
																				aria-label="Remove photo"
																			>
																				<svg class="size-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
																					<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
																				</svg>
																			</button>
																		{/if}
																	</div>
																{/each}
															</div>
														{/if}

														<!-- Photo upload for this finding -->
														{#if !isReadOnly && !offline}
															<label class="mt-2 flex cursor-pointer items-center gap-1.5 text-xs text-slate-500 hover:text-slate-300">
																{#if photoUploading.has(finding.id)}
																	<div class="size-3.5 animate-spin rounded-full border border-slate-600 border-t-blue-400"></div>
																	Uploading…
																{:else}
																	<svg class="size-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
																		<path stroke-linecap="round" stroke-linejoin="round" d="M6.827 6.175A2.31 2.31 0 0 1 5.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 0 0-1.134-.175 2.31 2.31 0 0 1-1.64-1.055l-.822-1.316a2.192 2.192 0 0 0-1.736-1.039 48.774 48.774 0 0 0-5.232 0 2.192 2.192 0 0 0-1.736 1.039l-.821 1.316Z" />
																		<path stroke-linecap="round" stroke-linejoin="round" d="M16.5 12.75a4.5 4.5 0 1 1-9 0 4.5 4.5 0 0 1 9 0ZM18.75 10.5h.008v.008h-.008V10.5Z" />
																	</svg>
																	Add photo
																{/if}
																<input
																	type="file"
																	accept="image/jpeg,image/png,image/heic"
																	capture="environment"
																	class="sr-only"
																	onchange={(e) => uploadPhoto(activeSystemData.system_type, item.item_key, finding.id, e)}
																/>
															</label>
														{/if}
													</div>
												{/each}
											</div>
										{/if}

										{#if photoError}
											<p class="text-xs text-red-400">{photoError}</p>
										{/if}

										<!-- Finding form (add/edit) -->
										{#if findingForm?.systemType === activeSystemData.system_type && findingForm?.itemKey === item.item_key}
											<div class="rounded-lg border border-blue-800 bg-blue-950/20 p-3 space-y-3">
												<p class="text-xs font-semibold text-blue-300">
													{findingForm.findingId ? 'Edit Finding' : 'Add Finding'}
												</p>

												{#if findingForm.error}
													<p class="text-xs text-red-400">{findingForm.error}</p>
												{/if}

												<textarea
													bind:value={findingForm.narrative}
													rows={3}
													placeholder="Describe the observation…"
													class="w-full resize-none rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm
														text-white placeholder-slate-600 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
												></textarea>

												<label class="flex cursor-pointer items-center gap-2">
													<input
														type="checkbox"
														bind:checked={findingForm.isDeficiency}
														class="size-4 rounded border-slate-600 bg-slate-800 text-red-500 focus:ring-red-500"
													/>
													<span class="text-sm text-slate-300">Report as deficiency (in need of correction)</span>
												</label>

												<div class="flex gap-2">
													<button
														onclick={() => (findingForm = null)}
														class="flex-1 rounded-lg border border-slate-700 bg-slate-800 py-2 text-xs font-medium
															text-slate-400 hover:bg-slate-700"
													>
														Cancel
													</button>
													<button
														onclick={submitFinding}
														disabled={findingForm.saving || !findingForm.narrative.trim()}
														class="flex-1 rounded-lg bg-blue-600 py-2 text-xs font-semibold text-white
															hover:bg-blue-500 active:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
													>
														{findingForm.saving ? 'Saving…' : 'Save'}
													</button>
												</div>
											</div>
										{:else if !isReadOnly && !offline}
											<button
												onclick={() => openFindingForm(activeSystemData.system_type, item.item_key)}
												class="flex w-full items-center justify-center gap-1.5 rounded-lg border border-slate-700
													bg-slate-800 py-2 text-xs font-medium text-slate-400 hover:bg-slate-700 hover:text-white"
											>
												<svg class="size-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
													<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
												</svg>
												Add Finding
											</button>
										{/if}
									</div>
								{/if}
							</li>
						{/each}
					</ul>
				</div>
			{/if}
		</main>
	</div>

	<!-- ── Complete modal ─────────────────────────────────────────────────────── -->
	{#if showCompleteModal}
		<div
			class="fixed inset-0 z-50 flex items-end justify-center bg-black/60 sm:items-center"
			role="dialog"
			aria-modal="true"
		>
			<div class="w-full max-w-sm rounded-t-2xl border border-slate-700 bg-slate-900 p-6 sm:rounded-2xl">
				<h2 class="text-lg font-bold text-white">Complete Inspection?</h2>

				{#if !isReadOnly}
					{@const prog = totalProgress()}
					<p class="mt-1 text-sm text-slate-400">
						{prog.addressed} of {prog.total} items addressed.
					</p>
				{/if}

				{#if completeErrors.length > 0}
					<div class="mt-4 rounded-lg border border-red-800 bg-red-950 p-3">
						<p class="mb-1 text-xs font-semibold text-red-300">Cannot complete — fix these issues:</p>
						<ul class="space-y-1">
							{#each completeErrors as e}
								<li class="text-xs text-red-400">• {e}</li>
							{/each}
						</ul>
					</div>
				{:else}
					<p class="mt-3 text-sm text-slate-300">
						This will lock the inspection and allow you to generate the PDF report.
					</p>
				{/if}

				<div class="mt-5 flex gap-3">
					<button
						onclick={() => { showCompleteModal = false; completeErrors = []; }}
						class="flex-1 rounded-xl border border-slate-700 bg-slate-800 py-3 text-sm font-medium
							text-slate-400 hover:bg-slate-700"
					>
						{completeErrors.length > 0 ? 'Go Back' : 'Cancel'}
					</button>
					{#if completeErrors.length === 0}
						<button
							onclick={tryComplete}
							disabled={completing}
							class="flex-1 rounded-xl bg-blue-600 py-3 text-sm font-semibold text-white
								hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-50"
						>
							{completing ? 'Completing…' : 'Complete'}
						</button>
					{/if}
				</div>
			</div>
		</div>
	{/if}
{/if}
