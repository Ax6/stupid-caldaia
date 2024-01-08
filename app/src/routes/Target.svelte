<script lang="ts">
	import type { BoilerData, Rule } from './+page.server';
	import { porca, madonne, gql } from '$lib/porca-madonna-ql';
	import type { Duration } from 'date-fns';
	import { formatRelative, formatDuration, formatISODuration } from 'date-fns';
	import { it } from 'date-fns/locale';

	export let data: BoilerData;

	let editing: boolean = false;
	let targetTempIndex: number = 2;
	let targetTimeIndex: number = 1;
	const possibleTargetTemps: number[] = [10, 15, 18, 20, 22, 25];
	const possibleTargetTimes: Duration[] = [30, 60, 120, 240, 360, 480].map((minutes) => {
		return { minutes: minutes % 60, hours: Math.floor(minutes / 60) };
	});

	let subscription = madonne<BoilerData>(gql`
		subscription {
			boiler {
				rule {
					id
					start
					duration
					targetTemp
				}
			}
		}
	`);
	subscription.set({ boiler: data.boiler });

	$: rule = $subscription.boiler.rule;
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
			alert('Errore');
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

<div class="w-full min-h-32 bg-gray-200 grid place-items-center rounded-xl text-4xl">
	{#if regolaVeloce}
		<p class="mt-4">
			Mantieni {regolaVeloce.targetTemp}°C
		</p>
		<p class="text-base my-2">
			Impostato {onTime}
		</p>
		<button class="bg-red-400 hover:bg-red-500 w-full p-4 rounded-xl" on:click={unset}
			>Cancella</button
		>
	{:else if editing}
		<div class="container flex justify-around place-items-center flex-col lg:flex-row">
			<button
				class="bg-red-400 hover:bg-red-500 w-full p-4 lg:w-64 rounded-xl"
				on:click={() => (editing = false)}>Annulla</button
			>
			<div class="flex items-center text-3xl lg:text-4xl py-4">
				<p>Mantieni</p>
				<button
					class="p-2 m-2 bg-blue-400 hover:bg-blue-500 rounded-xl"
					on:click={changeTargetTemp}
				>
					{possibleTargetTemps[targetTempIndex]}°C</button
				>
				<p>per</p>
				<button
					class="p-2 m-2 bg-blue-400 hover:bg-blue-500 rounded-xl"
					on:click={changeTargetTime}
				>
					{formatDuration(possibleTargetTimes[targetTimeIndex], { locale: it })}</button
				>
			</div>
			<button class="bg-green-400 hover:bg-green-500 p-4 w-full lg:w-64 rounded-xl" on:click={set}
				>Imposta</button
			>
		</div>
	{:else}
		<button
			class="bg-blue-400 hover:bg-blue-500 w-full h-full rounded-xl"
			on:click={() => (editing = true)}>Imposta regola veloce</button
		>
	{/if}
</div>
