<script lang="ts">
	import type { BoilerData, Rule } from '$lib/types';
	import type { Readable } from 'svelte/store';
	import { porca, gql } from '$lib/porca-madonna-ql';
	import type { Duration } from 'date-fns';
	import { formatRelative, formatDuration, formatISODuration } from 'date-fns';
	import { it } from 'date-fns/locale';
	import { popup } from '$lib/popup';

	export let boilerSubscription: Readable<BoilerData>;

	let editing: boolean = false;
	let targetTempIndex: number = 2;
	let targetTimeIndex: number = 1;
	const possibleTargetTemps: number[] = [10, 15, 18, 20, 22, 25];
	const possibleTargetTimes: Duration[] = [30, 60, 120, 240, 360, 480].map((minutes) => {
		return { minutes: minutes % 60, hours: Math.floor(minutes / 60) };
	});

	$: rule = $boilerSubscription.boiler.rules;
	$: regolaVeloce = rule.find((interval) => interval.id === 'regola-veloce');
	$: onTime = regolaVeloce ? formatRelative(regolaVeloce.start, new Date(), { locale: it }) : null;

	function changeTargetTemp() {
		targetTempIndex = (targetTempIndex + 1) % possibleTargetTemps.length;
	}

	function changeTargetTime() {
		targetTimeIndex = (targetTimeIndex + 1) % possibleTargetTimes.length;
	}

	async function unset() {
		const result = await porca<boolean>(gql`
			mutation quickTarget {
				deleteRule(id: "regola-veloce")
			}
		`);
		if (result) {
			editing = false;
		} else {
			popup.set({ messages: ['Regola veloce non impostata'] });
		}
	}

	async function set() {
		const result = await porca<Rule>(
			gql`
				mutation quickTarget($targetTemp: Float!, $start: Time!, $duration: Duration!) {
					setRule(
						id: "regola-veloce"
						start: $start
						duration: $duration
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
				duration: formatISODuration(possibleTargetTimes[targetTimeIndex])
			}
		);
		if (result) {
			editing = false;
		} else {
			alert('Errore');
		}
	}
</script>

{#if regolaVeloce && regolaVeloce.isActive}
	<div class="w-full bg-gray-800 border border-black grid place-items-center rounded-xl text-4xl text-white p-2">
		<p class="font-thin">
			Mantieni {regolaVeloce.targetTemp}°C
		</p>
		<p class="text-base mt-2 mb-4 text-gray-300">
			Impostato {onTime}
		</p>
		<button class="bg-red-400 hover:bg-red-500 w-full p-3 rounded-xl" on:click={unset}
			>Cancella</button
		>
	</div>
{:else if editing}
	<div class="rounded-xl bg-orange-300 border border-orange-400 flex justify-around place-items-center flex-col lg:flex-row">
		<div class="flex items-center text-2xl lg:text-3xl">
			<p>Mantieni</p>
			<button class="p-2 m-2 bg-blue-400 hover:bg-blue-500 rounded-xl" on:click={changeTargetTemp}>
				{possibleTargetTemps[targetTempIndex]}°C</button
			>
			<p>per</p>
			<button class="p-2 m-2 bg-blue-400 hover:bg-blue-500 rounded-xl" on:click={changeTargetTime}>
				{formatDuration(possibleTargetTimes[targetTimeIndex], { locale: it })}</button
			>
		</div>
		<div class="w-full flex p-2">
			<button class="flex-grow bg-green-400 hover:bg-green-500 p-2 lg:w-64 rounded-lg" on:click={set}
				>Salva</button
			>
			<div class="w-2"></div>
			<button
				class="bg-gray-400 hover:bg-gray-500 p-2 lg:w-64 rounded-lg"
				on:click={() => (editing = false)}>Annulla</button
			>
		</div>
	</div>
{:else}
	<button
		class="bg-blue-400 hover:bg-blue-500 border border-blue-600 w-full py-6 rounded-xl text-4xl"
		on:click={() => (editing = true)}>Imposta veloce</button
	>
{/if}
