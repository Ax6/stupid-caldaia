<script lang="ts">
	import type { BoilerData, ProgrammedInterval } from './+page.server';
	import { porca, madonne, gql } from '$lib/porca-madonna-ql';
	import moment from 'moment';

	export let data: BoilerData;

	let editing: boolean = false;
	let targetTempIndex: number = 2;
	let targetTimeIndex: number = 1;
	const possibleTargetTemps = [10, 15, 18, 20, 22, 25];
	const possibleTargetTimes = [
		{
			time: moment.duration(30, 'minutes'),
			label: '30 minuti'
		},
		{
			time: moment.duration(1, 'hours'),
			label: '1 ora'
		},
		{
			time: moment.duration(2, 'hours'),
			label: '2 ore'
		},
		{
			time: moment.duration(4, 'hours'),
			label: '4 ore'
		},
		{
			time: moment.duration(6, 'hours'),
			label: '6 ore'
		}
	];

	let subscription = madonne<BoilerData>(gql`
		subscription {
			boiler {
				programmedIntervals {
					id
					start
					duration
					targetTemp
				}
			}
		}
	`);
	subscription.set({ boiler: data.boiler });

	$: programmedIntervals = $subscription.boiler.programmedIntervals;
	$: regolaVeloce = programmedIntervals.find((interval) => interval.id === 'regola-veloce');

	function changeTargetTemp() {
		targetTempIndex = (targetTempIndex + 1) % possibleTargetTemps.length;
	}

	function changeTargetTime() {
		targetTimeIndex = (targetTimeIndex + 1) % possibleTargetTimes.length;
	}

	async function set() {
		const result = await porca<ProgrammedInterval>(
			gql`
				mutation quickTarget($targetTemp: Float!, $start: Time!, $duration: Duration!) {
					setProgrammedInterval(
						interval: {
							id: "regola-veloce"
							start: $start
							duration: $duration
							targetTemp: $targetTemp
							repeatDays: []
						}
					) {
						id
					}
				}
			`,
			{
				targetTemp: possibleTargetTemps[targetTempIndex],
				start: new Date().toISOString(),
				duration: possibleTargetTimes[targetTimeIndex].time.toISOString()
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
		<p>Mantieni {regolaVeloce.targetTemp}°C</p>
		<button>Annulla</button>
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
					{possibleTargetTimes[targetTimeIndex].label}</button
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
