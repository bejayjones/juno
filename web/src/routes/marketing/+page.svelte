<svelte:head>
	<title>Juno — Home Inspection Software for Professionals</title>
	<meta
		name="description"
		content="Mobile-first inspection software for licensed home inspectors. InterNACHI-compliant reports, offline-ready, built for the field."
	/>
</svelte:head>

<div class="min-h-dvh bg-slate-950 text-white">
	<!-- Navigation -->
	<nav class="fixed top-0 inset-x-0 z-50 border-b border-slate-800/50 bg-slate-950/80 backdrop-blur-lg">
		<div class="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
			<a href="/marketing" class="text-xl font-bold tracking-tight">Juno</a>
			<div class="flex items-center gap-3">
				<a
					href="/login"
					class="rounded-lg px-4 py-2 text-sm font-medium text-slate-300 transition-colors hover:text-white"
				>
					Sign in
				</a>
				<a
					href="/register"
					class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-blue-500"
				>
					Get started
				</a>
			</div>
		</div>
	</nav>

	<!-- Hero -->
	<section class="relative overflow-hidden pt-32 pb-20 px-6">
		<!-- Gradient background -->
		<div
			class="pointer-events-none absolute inset-0 -top-40"
			aria-hidden="true"
		>
			<div
				class="absolute left-1/2 top-0 -translate-x-1/2 h-[600px] w-[900px] rounded-full opacity-15"
				style="background: radial-gradient(ellipse, oklch(0.6 0.2 250), transparent 70%);"
			></div>
		</div>

		<div class="relative mx-auto max-w-3xl text-center">
			<div class="mb-6 inline-flex items-center gap-2 rounded-full border border-slate-700 bg-slate-900 px-4 py-1.5 text-xs font-medium text-slate-300">
				<span class="inline-block size-1.5 rounded-full bg-emerald-400"></span>
				InterNACHI SOP compliant
			</div>
			<h1 class="text-4xl font-extrabold tracking-tight sm:text-5xl lg:text-6xl">
				Inspect smarter.
				<br />
				<span class="bg-gradient-to-r from-blue-400 to-cyan-300 bg-clip-text text-transparent">
					Report faster.
				</span>
			</h1>
			<p class="mx-auto mt-6 max-w-xl text-lg leading-relaxed text-slate-400">
				Juno is the mobile-first platform built for licensed home inspectors. Schedule, inspect, and deliver professional PDF reports — even without internet.
			</p>
			<div class="mt-10 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
				<a
					href="/register"
					class="w-full rounded-xl bg-blue-600 px-8 py-3.5 text-center text-sm font-semibold transition-colors hover:bg-blue-500 active:bg-blue-700 sm:w-auto"
				>
					Start free trial
				</a>
				<a
					href="#features"
					class="w-full rounded-xl border border-slate-700 bg-slate-900 px-8 py-3.5 text-center text-sm font-semibold text-slate-300 transition-colors hover:border-slate-600 hover:text-white sm:w-auto"
				>
					See how it works
				</a>
			</div>
		</div>

		<!-- Phone mockup -->
		<div class="relative mx-auto mt-16 max-w-sm">
			<div class="rounded-[2.5rem] border border-slate-700 bg-slate-900 p-3 shadow-2xl shadow-blue-950/30">
				<div class="rounded-[2rem] bg-slate-800 overflow-hidden">
					<!-- Status bar -->
					<div class="flex items-center justify-between px-6 py-2 text-xs text-slate-400">
						<span>9:41</span>
						<div class="flex items-center gap-1.5">
							<svg class="size-3.5" fill="currentColor" viewBox="0 0 24 24"><path d="M1 9l2 2c4.97-4.97 13.03-4.97 18 0l2-2C16.93 2.93 7.08 2.93 1 9zm8 8l3 3 3-3a4.237 4.237 0 00-6 0zm-4-4l2 2a7.074 7.074 0 0110 0l2-2C15.14 9.14 8.87 9.14 5 13z"/></svg>
							<svg class="size-3.5" fill="currentColor" viewBox="0 0 24 24"><path d="M15.67 4H14V2h-4v2H8.33C7.6 4 7 4.6 7 5.33v15.33C7 21.4 7.6 22 8.33 22h7.33c.74 0 1.34-.6 1.34-1.33V5.33C17 4.6 16.4 4 15.67 4z"/></svg>
						</div>
					</div>
					<!-- App content mockup -->
					<div class="px-5 py-4">
						<p class="text-lg font-bold text-white">Good morning, Sarah</p>
						<p class="text-xs text-slate-400 mt-0.5">3 inspections today</p>
						<div class="mt-4 space-y-2.5">
							<div class="rounded-xl bg-slate-700/50 border border-slate-700 p-3">
								<div class="flex items-center justify-between">
									<div>
										<p class="text-sm font-semibold text-white">142 Oak Street</p>
										<p class="text-xs text-slate-400">9:00 AM — Residential</p>
									</div>
									<span class="rounded-full bg-blue-600/20 px-2 py-0.5 text-xs font-medium text-blue-400">Scheduled</span>
								</div>
							</div>
							<div class="rounded-xl bg-slate-700/50 border border-slate-700 p-3">
								<div class="flex items-center justify-between">
									<div>
										<p class="text-sm font-semibold text-white">88 Maple Drive</p>
										<p class="text-xs text-slate-400">1:00 PM — Condo</p>
									</div>
									<span class="rounded-full bg-amber-600/20 px-2 py-0.5 text-xs font-medium text-amber-400">In Progress</span>
								</div>
							</div>
							<div class="rounded-xl bg-slate-700/50 border border-slate-700 p-3">
								<div class="flex items-center justify-between">
									<div>
										<p class="text-sm font-semibold text-white">7 Pine Court</p>
										<p class="text-xs text-slate-400">4:00 PM — Townhouse</p>
									</div>
									<span class="rounded-full bg-emerald-600/20 px-2 py-0.5 text-xs font-medium text-emerald-400">Complete</span>
								</div>
							</div>
						</div>
					</div>
					<!-- Bottom nav mockup -->
					<div class="flex border-t border-slate-700 px-2 py-2">
						<div class="flex flex-1 flex-col items-center gap-0.5 py-1">
							<svg class="size-5 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="m2.25 12 8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25" /></svg>
							<span class="text-[10px] text-blue-400">Home</span>
						</div>
						<div class="flex flex-1 flex-col items-center gap-0.5 py-1">
							<svg class="size-5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" /></svg>
							<span class="text-[10px] text-slate-500">Schedule</span>
						</div>
						<div class="flex flex-1 flex-col items-center gap-0.5 py-1">
							<svg class="size-5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" /></svg>
							<span class="text-[10px] text-slate-500">Clients</span>
						</div>
						<div class="flex flex-1 flex-col items-center gap-0.5 py-1">
							<svg class="size-5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" /></svg>
							<span class="text-[10px] text-slate-500">Reports</span>
						</div>
					</div>
				</div>
			</div>
		</div>
	</section>

	<!-- Trusted by / social proof bar -->
	<section class="border-y border-slate-800 bg-slate-900/50 py-10 px-6">
		<p class="text-center text-xs font-medium uppercase tracking-widest text-slate-500 mb-6">
			Trusted by inspection professionals
		</p>
		<div class="mx-auto flex max-w-3xl flex-wrap items-center justify-center gap-x-10 gap-y-4 text-slate-600">
			<span class="text-sm font-semibold tracking-wide">InterNACHI</span>
			<span class="text-slate-800">|</span>
			<span class="text-sm font-semibold tracking-wide">ASHI</span>
			<span class="text-slate-800">|</span>
			<span class="text-sm font-semibold tracking-wide">CREIA</span>
			<span class="text-slate-800">|</span>
			<span class="text-sm font-semibold tracking-wide">CAHPI</span>
			<span class="text-slate-800">|</span>
			<span class="text-sm font-semibold tracking-wide">NAHI</span>
		</div>
	</section>

	<!-- Features grid -->
	<section id="features" class="py-20 px-6 scroll-mt-20">
		<div class="mx-auto max-w-5xl">
			<div class="text-center mb-14">
				<h2 class="text-3xl font-bold tracking-tight sm:text-4xl">
					Everything you need in the field
				</h2>
				<p class="mt-3 text-base text-slate-400 max-w-xl mx-auto">
					From scheduling to delivery, Juno covers the entire inspection workflow on your phone or tablet.
				</p>
			</div>

			<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
				<!-- Feature 1 -->
				<div class="rounded-2xl border border-slate-800 bg-slate-900 p-6">
					<div class="mb-4 flex size-11 items-center justify-center rounded-xl bg-blue-600/10">
						<svg class="size-6 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M3 3v1.5M3 21v-6m0 0 2.77-.693a9 9 0 0 1 6.208.682l.108.054a9 9 0 0 0 6.086.71l3.114-.732a48.524 48.524 0 0 1-.005-10.499l-3.11.732a9 9 0 0 1-6.085-.711l-.108-.054a9 9 0 0 0-6.208-.682L3 4.5M3 15V4.5" />
						</svg>
					</div>
					<h3 class="text-base font-semibold text-white">InterNACHI Compliant</h3>
					<p class="mt-2 text-sm leading-relaxed text-slate-400">
						All 10 mandatory systems (3.1–3.10) built in. Every "shall inspect," "shall describe," and "shall report" field is covered.
					</p>
				</div>

				<!-- Feature 2 -->
				<div class="rounded-2xl border border-slate-800 bg-slate-900 p-6">
					<div class="mb-4 flex size-11 items-center justify-center rounded-xl bg-emerald-600/10">
						<svg class="size-6 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
						</svg>
					</div>
					<h3 class="text-base font-semibold text-white">Works Offline</h3>
					<p class="mt-2 text-sm leading-relaxed text-slate-400">
						No Wi-Fi in the basement? No problem. Every feature works offline with local storage. Data syncs automatically when you reconnect.
					</p>
				</div>

				<!-- Feature 3 -->
				<div class="rounded-2xl border border-slate-800 bg-slate-900 p-6">
					<div class="mb-4 flex size-11 items-center justify-center rounded-xl bg-amber-600/10">
						<svg class="size-6 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
						</svg>
					</div>
					<h3 class="text-base font-semibold text-white">PDF Reports</h3>
					<p class="mt-2 text-sm leading-relaxed text-slate-400">
						Generate polished, client-ready PDF reports with photos, findings, and deficiencies. Brand them with your logo and colors.
					</p>
				</div>

				<!-- Feature 4 -->
				<div class="rounded-2xl border border-slate-800 bg-slate-900 p-6">
					<div class="mb-4 flex size-11 items-center justify-center rounded-xl bg-violet-600/10">
						<svg class="size-6 text-violet-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" />
						</svg>
					</div>
					<h3 class="text-base font-semibold text-white">Scheduling</h3>
					<p class="mt-2 text-sm leading-relaxed text-slate-400">
						Book inspections, track appointments by status, and manage your day from a single dashboard. Never double-book again.
					</p>
				</div>

				<!-- Feature 5 -->
				<div class="rounded-2xl border border-slate-800 bg-slate-900 p-6">
					<div class="mb-4 flex size-11 items-center justify-center rounded-xl bg-cyan-600/10">
						<svg class="size-6 text-cyan-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M6.827 6.175A2.31 2.31 0 0 1 5.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 0 0-1.134-.175 2.31 2.31 0 0 1-1.64-1.055l-.822-1.316a2.192 2.192 0 0 0-1.736-1.039 48.774 48.774 0 0 0-5.232 0 2.192 2.192 0 0 0-1.736 1.039l-.821 1.316Z" />
							<path stroke-linecap="round" stroke-linejoin="round" d="M16.5 12.75a4.5 4.5 0 1 1-9 0 4.5 4.5 0 0 1 9 0ZM18.75 10.5h.008v.008h-.008V10.5Z" />
						</svg>
					</div>
					<h3 class="text-base font-semibold text-white">Photo Documentation</h3>
					<p class="mt-2 text-sm leading-relaxed text-slate-400">
						Capture photos directly during walkthroughs and attach them to findings. Photos are stored locally and synced to the cloud.
					</p>
				</div>

				<!-- Feature 6 -->
				<div class="rounded-2xl border border-slate-800 bg-slate-900 p-6">
					<div class="mb-4 flex size-11 items-center justify-center rounded-xl bg-rose-600/10">
						<svg class="size-6 text-rose-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 0 1-2.25 2.25h-15a2.25 2.25 0 0 1-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25m19.5 0v.243a2.25 2.25 0 0 1-1.07 1.916l-7.5 4.615a2.25 2.25 0 0 1-2.36 0L3.32 8.91a2.25 2.25 0 0 1-1.07-1.916V6.75" />
						</svg>
					</div>
					<h3 class="text-base font-semibold text-white">One-Click Delivery</h3>
					<p class="mt-2 text-sm leading-relaxed text-slate-400">
						Email finished reports to clients, agents, and other parties directly from the app. Track who's received what.
					</p>
				</div>
			</div>
		</div>
	</section>

	<!-- How it works -->
	<section class="border-t border-slate-800 bg-slate-900/30 py-20 px-6">
		<div class="mx-auto max-w-4xl">
			<div class="text-center mb-14">
				<h2 class="text-3xl font-bold tracking-tight sm:text-4xl">How it works</h2>
				<p class="mt-3 text-base text-slate-400">
					Four steps from booking to delivery.
				</p>
			</div>

			<div class="grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
				<div class="text-center">
					<div class="mx-auto mb-4 flex size-12 items-center justify-center rounded-full bg-blue-600 text-lg font-bold">
						1
					</div>
					<h3 class="text-sm font-semibold text-white">Schedule</h3>
					<p class="mt-2 text-sm text-slate-400">
						Create an appointment with property details, client info, and inspection date.
					</p>
				</div>
				<div class="text-center">
					<div class="mx-auto mb-4 flex size-12 items-center justify-center rounded-full bg-blue-600 text-lg font-bold">
						2
					</div>
					<h3 class="text-sm font-semibold text-white">Inspect</h3>
					<p class="mt-2 text-sm text-slate-400">
						Walk through all 10 systems on your phone. Mark statuses, add findings, snap photos.
					</p>
				</div>
				<div class="text-center">
					<div class="mx-auto mb-4 flex size-12 items-center justify-center rounded-full bg-blue-600 text-lg font-bold">
						3
					</div>
					<h3 class="text-sm font-semibold text-white">Generate</h3>
					<p class="mt-2 text-sm text-slate-400">
						Tap one button to create a branded PDF report — no internet required.
					</p>
				</div>
				<div class="text-center">
					<div class="mx-auto mb-4 flex size-12 items-center justify-center rounded-full bg-blue-600 text-lg font-bold">
						4
					</div>
					<h3 class="text-sm font-semibold text-white">Deliver</h3>
					<p class="mt-2 text-sm text-slate-400">
						Send the report to clients and agents with one click. Done before you leave the property.
					</p>
				</div>
			</div>
		</div>
	</section>

	<!-- Testimonials -->
	<section class="py-20 px-6">
		<div class="mx-auto max-w-5xl">
			<div class="text-center mb-14">
				<h2 class="text-3xl font-bold tracking-tight sm:text-4xl">Inspectors love Juno</h2>
			</div>

			<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
				<div class="rounded-2xl border border-slate-800 bg-slate-900 p-6">
					<div class="flex items-center gap-1 mb-3">
						{#each Array(5) as _}
							<svg class="size-4 text-amber-400" viewBox="0 0 20 20" fill="currentColor">
								<path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
							</svg>
						{/each}
					</div>
					<p class="text-sm leading-relaxed text-slate-300">
						"I used to spend 2 hours on reports after each inspection. With Juno, I generate the PDF before I leave the property. My clients are amazed at the turnaround."
					</p>
					<div class="mt-4 flex items-center gap-3">
						<div class="flex size-9 items-center justify-center rounded-full bg-blue-600/20 text-sm font-bold text-blue-400">
							M
						</div>
						<div>
							<p class="text-sm font-medium text-white">Mike Torres</p>
							<p class="text-xs text-slate-500">Licensed Inspector, Texas</p>
						</div>
					</div>
				</div>

				<div class="rounded-2xl border border-slate-800 bg-slate-900 p-6">
					<div class="flex items-center gap-1 mb-3">
						{#each Array(5) as _}
							<svg class="size-4 text-amber-400" viewBox="0 0 20 20" fill="currentColor">
								<path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
							</svg>
						{/each}
					</div>
					<p class="text-sm leading-relaxed text-slate-300">
						"The offline mode is a game-changer. Half my inspections are in rural areas with no signal. Juno just works. I sync everything when I'm back in my truck."
					</p>
					<div class="mt-4 flex items-center gap-3">
						<div class="flex size-9 items-center justify-center rounded-full bg-emerald-600/20 text-sm font-bold text-emerald-400">
							R
						</div>
						<div>
							<p class="text-sm font-medium text-white">Rachel Nguyen</p>
							<p class="text-xs text-slate-500">Home Inspector, Oregon</p>
						</div>
					</div>
				</div>

				<div class="rounded-2xl border border-slate-800 bg-slate-900 p-6">
					<div class="flex items-center gap-1 mb-3">
						{#each Array(5) as _}
							<svg class="size-4 text-amber-400" viewBox="0 0 20 20" fill="currentColor">
								<path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
							</svg>
						{/each}
					</div>
					<p class="text-sm leading-relaxed text-slate-300">
						"Finally, software built by someone who understands InterNACHI standards. Every system, every field — it's all there. No more guessing what to include."
					</p>
					<div class="mt-4 flex items-center gap-3">
						<div class="flex size-9 items-center justify-center rounded-full bg-amber-600/20 text-sm font-bold text-amber-400">
							D
						</div>
						<div>
							<p class="text-sm font-medium text-white">David Park</p>
							<p class="text-xs text-slate-500">Certified Inspector, Florida</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	</section>

	<!-- Pricing -->
	<section class="border-t border-slate-800 bg-slate-900/30 py-20 px-6">
		<div class="mx-auto max-w-4xl">
			<div class="text-center mb-14">
				<h2 class="text-3xl font-bold tracking-tight sm:text-4xl">Simple pricing</h2>
				<p class="mt-3 text-base text-slate-400">
					No per-report fees. No hidden costs. One plan, everything included.
				</p>
			</div>

			<div class="mx-auto grid max-w-lg gap-4 lg:max-w-none lg:grid-cols-3">
				<!-- Free tier -->
				<div class="rounded-2xl border border-slate-800 bg-slate-900 p-8">
					<h3 class="text-lg font-semibold text-white">Starter</h3>
					<p class="mt-1 text-sm text-slate-400">Try Juno with no commitment</p>
					<div class="mt-6">
						<span class="text-4xl font-extrabold text-white">$0</span>
						<span class="text-sm text-slate-500">/month</span>
					</div>
					<ul class="mt-8 space-y-3 text-sm text-slate-300">
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							5 inspections per month
						</li>
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							Full offline mode
						</li>
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							PDF report generation
						</li>
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							Local data storage
						</li>
					</ul>
					<a
						href="/register"
						class="mt-8 block w-full rounded-xl border border-slate-700 bg-slate-800 px-4 py-3 text-center text-sm font-semibold text-white transition-colors hover:bg-slate-700"
					>
						Get started free
					</a>
				</div>

				<!-- Pro tier (highlighted) -->
				<div class="relative rounded-2xl border-2 border-blue-500 bg-slate-900 p-8 shadow-lg shadow-blue-600/10">
					<div class="absolute -top-3 left-1/2 -translate-x-1/2">
						<span class="rounded-full bg-blue-600 px-3 py-1 text-xs font-semibold">Most popular</span>
					</div>
					<h3 class="text-lg font-semibold text-white">Professional</h3>
					<p class="mt-1 text-sm text-slate-400">For full-time inspectors</p>
					<div class="mt-6">
						<span class="text-4xl font-extrabold text-white">$49</span>
						<span class="text-sm text-slate-500">/month</span>
					</div>
					<ul class="mt-8 space-y-3 text-sm text-slate-300">
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							Unlimited inspections
						</li>
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							Cloud sync & backup
						</li>
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							Email report delivery
						</li>
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							Custom branding
						</li>
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							Priority support
						</li>
					</ul>
					<a
						href="/register"
						class="mt-8 block w-full rounded-xl bg-blue-600 px-4 py-3 text-center text-sm font-semibold text-white transition-colors hover:bg-blue-500"
					>
						Start 14-day free trial
					</a>
				</div>

				<!-- Team tier -->
				<div class="rounded-2xl border border-slate-800 bg-slate-900 p-8">
					<h3 class="text-lg font-semibold text-white">Company</h3>
					<p class="mt-1 text-sm text-slate-400">For inspection firms</p>
					<div class="mt-6">
						<span class="text-4xl font-extrabold text-white">$99</span>
						<span class="text-sm text-slate-500">/month</span>
					</div>
					<ul class="mt-8 space-y-3 text-sm text-slate-300">
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							Everything in Professional
						</li>
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							Up to 10 inspectors
						</li>
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							Team scheduling
						</li>
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							Admin dashboard
						</li>
						<li class="flex items-start gap-2">
							<svg class="mt-0.5 size-4 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							API access
						</li>
					</ul>
					<a
						href="/register"
						class="mt-8 block w-full rounded-xl border border-slate-700 bg-slate-800 px-4 py-3 text-center text-sm font-semibold text-white transition-colors hover:bg-slate-700"
					>
						Contact sales
					</a>
				</div>
			</div>
		</div>
	</section>

	<!-- Final CTA -->
	<section class="py-20 px-6">
		<div class="mx-auto max-w-2xl text-center">
			<h2 class="text-3xl font-bold tracking-tight sm:text-4xl">
				Ready to modernize your inspections?
			</h2>
			<p class="mt-4 text-base text-slate-400">
				Join hundreds of licensed inspectors who deliver faster, more professional reports with Juno.
			</p>
			<div class="mt-8 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
				<a
					href="/register"
					class="w-full rounded-xl bg-blue-600 px-8 py-3.5 text-center text-sm font-semibold transition-colors hover:bg-blue-500 sm:w-auto"
				>
					Start your free trial
				</a>
				<a
					href="/login"
					class="w-full rounded-xl border border-slate-700 bg-slate-900 px-8 py-3.5 text-center text-sm font-semibold text-slate-300 transition-colors hover:border-slate-600 hover:text-white sm:w-auto"
				>
					Sign in to your account
				</a>
			</div>
		</div>
	</section>

	<!-- Footer -->
	<footer class="border-t border-slate-800 py-10 px-6">
		<div class="mx-auto max-w-5xl">
			<div class="flex flex-col items-center gap-6 sm:flex-row sm:justify-between">
				<div>
					<p class="text-sm font-semibold text-white">Juno</p>
					<p class="mt-0.5 text-xs text-slate-500">Home inspection software for professionals</p>
				</div>
				<div class="flex gap-6 text-xs text-slate-500">
					<a href="/marketing" class="hover:text-slate-300">Home</a>
					<a href="#features" class="hover:text-slate-300">Features</a>
					<a href="/login" class="hover:text-slate-300">Sign in</a>
					<a href="/register" class="hover:text-slate-300">Get started</a>
				</div>
			</div>
			<p class="mt-8 text-center text-xs text-slate-600">
				&copy; {new Date().getFullYear()} Juno. All rights reserved.
			</p>
		</div>
	</footer>
</div>
