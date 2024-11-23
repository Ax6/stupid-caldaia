<script lang="ts">
	import type { BoilerData, Rule } from '$lib/types';
	import type { Readable } from 'svelte/store';
	import type { Duration } from 'date-fns';
	import { porca, gql } from '$lib/porca-madonna-ql';
	import { parseISODuration } from '$lib/faster-than-open-source';
	import {
		formatRelative,
		formatDuration,
		formatISODuration,
		add,
		intervalToDuration,
		isBefore
	} from 'date-fns';
	import { it } from 'date-fns/locale';
	import { popup } from '$lib/popup';
	import { isZero, toSeconds, sum } from 'duration-fns';
	import { onMount } from 'svelte';

	interface Props {
		boilerSubscription: Readable<BoilerData>;
	}

	let { boilerSubscription }: Props = $props();

	let editing: boolean = $state(false);
	let delayStartDurationIndex: number = $state(0);
	let targetTempIndex: number = $state(3);
	let targetTimeIndex: number = $state(1);
	const possibleStartDelays: Duration[] = [
		{ seconds: 0 },
		{ minutes: 15 },
		{ minutes: 30 },
		{ hours: 1 },
		{ hours: 1, minutes: 30 },
		{ hours: 2 },
		{ hours: 3 },
		{ hours: 4 },
		{ hours: 6 },
		{ hours: 8 }
	];
	const possibleTargetTemps: number[] = [15, 18, 20, 22, 23, 24, 25];
	const possibleTargetTimes: Duration[] = [30, 60, 120, 240, 360, 480].map((minutes) => {
		return { minutes: minutes % 60, hours: Math.floor(minutes / 60) };
	});

	let rule = $derived($boilerSubscription.boiler.rules);

	/// Block to derive the active rule parameters
	let regolaAttiva = $derived(
		rule
			.filter((interval) => interval.isActive)
			.sort((a, b) => b.targetTemp - a.targetTemp)
			.at(0) ??
			// Otherwise we invent a fake rule, it's just to derive some values
			// After all, we're not going to use it
			({
				id: 'regola-veloce',
				delay: 'PT0S',
				duration: 'PT0S',
				start: '',
				targetTemp: 0,
				repeatDays: [],
				isActive: false,
				stoppedTime: new Date()
			} as Rule)
	);
	let ruleRealStartTime = $derived.by(() => {
		const originalSetTime = new Date(regolaAttiva.start);
		const delay = parseISODuration(regolaAttiva.delay);
		// The real start time is not just the sum of the two.
		// If the rule is active, the start time is definetely in the past
		// So what's the closes past time that has the same HH:MM:SS of the sum?
		const [setHH, setMM, setSS] = [
			originalSetTime.getHours(),
			originalSetTime.getMinutes(),
			originalSetTime.getSeconds()
		];
		const now = new Date();
		const possibleSetTime = new Date(
			now.getFullYear(),
			now.getMonth(),
			now.getDate(),
			setHH,
			setMM,
			setSS
		);
		let actualSetTime: Date;
		if (isBefore(now, possibleSetTime)) {
			// Actually, the rule started, but it was yesterday
			actualSetTime = add(possibleSetTime, { days: -1 });
		} else {
			// The rule started today
			actualSetTime = possibleSetTime;
		}
		return add(actualSetTime, delay);
	});
	let now = $state(new Date());
	let ruleRealEndTime = $derived(add(ruleRealStartTime, parseISODuration(regolaAttiva.duration)));
	let timeToStart = $derived(intervalToDuration({ start: now, end: ruleRealStartTime }));
	let hasStarted = $derived(toSeconds(timeToStart) <= 0);
	/// <-- End of the block

	function changeStartDelay() {
		delayStartDurationIndex = (delayStartDurationIndex + 1) % possibleStartDelays.length;
	}

	function changeTargetTemp() {
		targetTempIndex = (targetTempIndex + 1) % possibleTargetTemps.length;
	}

	function changeTargetTime() {
		targetTimeIndex = (targetTimeIndex + 1) % possibleTargetTimes.length;
	}

	async function unset() {
		const result = await porca<boolean>(
			gql`
				mutation stopRule($id: ID!) {
					stopRule(id: $id)
				}
			`,
			{ id: regolaAttiva?.id }
		);
		if (result) {
			editing = false;
		} else {
			popup.set({ messages: ['Regola veloce non impostata'] });
		}
	}

	async function set() {
		const result = await porca<Rule>(
			gql`
				mutation quickTarget(
					$targetTemp: Float!
					$start: Time!
					$duration: Duration!
					$delay: Duration!
				) {
					setRule(
						id: "regola-veloce"
						start: $start
						duration: $duration
						delay: $delay
						targetTemp: $targetTemp
						repeatDays: []
					) {
						id
					}
				}
			`,
			{
				targetTemp: possibleTargetTemps[targetTempIndex],
				start: new Date().toISOString(),
				duration: formatISODuration(possibleTargetTimes[targetTimeIndex]),
				delay: formatISODuration(possibleStartDelays[delayStartDurationIndex])
			}
		);
		if (result) {
			editing = false;
		} else {
			alert('Errore');
		}
	}

	onMount(() => {
		const interval = setInterval(() => {
			now = new Date();
		}, 1000);

		return () => {
			clearInterval(interval);
		};
	});
</script>

{#if regolaAttiva && regolaAttiva.isActive}
	<div
		class="w-full bg-gray-800 border border-black grid place-items-center rounded-xl text-4xl text-white p-2"
	>
		<p class="font-thin">
			Mantieni {regolaAttiva.targetTemp}°C
		</p>
		<p class="text-sm mt-2 mb-4 text-gray-300">
			{#if !hasStarted}
				Tra {formatDuration(timeToStart, { locale: it })}.
			{/if}
			Finisce {formatRelative(ruleRealEndTime, new Date(), { locale: it })}
		</p>
		<button class="bg-red-400 hover:bg-red-500 w-full p-3 rounded-xl" onclick={unset}
			>Cancella</button
		>
	</div>
{:else if editing}
	<div
		class="rounded-xl bg-orange-300 border border-orange-400 flex justify-around place-items-center flex-col lg:flex-row"
	>
		<div class="flex items-center text-2xl lg:text-3xl">
			{#if isZero(possibleStartDelays[delayStartDurationIndex])}
				<p>Da</p>
			{:else}
				<p>Tra</p>
			{/if}
			<button class="p-2 m-2 bg-blue-400 hover:bg-blue-500 rounded-xl" onclick={changeStartDelay}>
				{isZero(possibleStartDelays[delayStartDurationIndex])
					? 'adesso'
					: formatDuration(possibleStartDelays[delayStartDurationIndex], { locale: it })}
			</button>
			<button class="p-2 m-2 bg-blue-400 hover:bg-blue-500 rounded-xl" onclick={changeTargetTemp}>
				{possibleTargetTemps[targetTempIndex]}°C
			</button>
			<p>per</p>
			<button class="p-2 m-2 bg-blue-400 hover:bg-blue-500 rounded-xl" onclick={changeTargetTime}>
				{formatDuration(possibleTargetTimes[targetTimeIndex], { locale: it })}
			</button>
		</div>
		<div class="w-full flex p-2">
			<button class="flex-grow bg-green-400 hover:bg-green-500 p-2 lg:w-64 rounded-lg" onclick={set}
				>Salva</button
			>
			<div class="w-2"></div>
			<button
				class="bg-gray-400 hover:bg-gray-500 p-2 lg:w-64 rounded-lg"
				onclick={() => (editing = false)}>Annulla</button
			>
		</div>
	</div>
{:else}
	<button
		class="bg-blue-400 hover:bg-blue-500 border border-blue-600 w-full py-6 rounded-xl text-4xl"
		onclick={() => (editing = true)}>Imposta veloce</button
	>
{/if}
