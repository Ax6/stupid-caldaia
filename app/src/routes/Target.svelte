<script lang="ts">
	import type { BoilerData } from './+page.server';

	export let data: BoilerData;

	let editing: boolean = false;
	let targetTempIndex: number = 2;
	let targetTimeIndex: number = 1;
	const possibleTargetTemps = [10, 15, 18, 20, 22, 25];
	const possibleTargetTimes = [
		{
			time: 30,
			label: '30 minuti'
		},
		{
			time: 60,
			label: '1 ora'
		},
		{
			time: 120,
			label: '2 ore'
		},
		{
			time: 240,
			label: '4 ore'
		},
		{
			time: 360,
			label: '6 ore'
		}
	];

	function changeTargetTemp() {
		targetTempIndex = (targetTempIndex + 1) % possibleTargetTemps.length;
	}

	function changeTargetTime() {
		targetTimeIndex = (targetTimeIndex + 1) % possibleTargetTimes.length;
	}
</script>

<div class="container min-h-32 bg-gray-200 grid place-items-center rounded-xl text-4xl">
	{#if data.boiler.targetTemp}
		<p>Mantieni {data.boiler.targetTemp}°C</p>
		<button>Annulla</button>
	{:else if editing}
		<div class="container flex justify-around place-items-center flex-col lg:flex-row">
			<button class="bg-red-400 hover:bg-red-500 w-full p-4 lg:w-64 rounded-xl" on:click={() => (editing = false)}
				>Annulla</button
			>
			<div class="flex items-center text-3xl lg:text-4xl py-4">
				<p>Mantieni</p>
				<button class="p-2 m-2 bg-blue-400 hover:bg-blue-500 rounded-xl" on:click={changeTargetTemp}>
					{possibleTargetTemps[targetTempIndex]}°C</button
				>
				<p>per</p>
				<button class="p-2 m-2 bg-blue-400 hover:bg-blue-500 rounded-xl" on:click={changeTargetTime}>
					{possibleTargetTimes[targetTimeIndex].label}</button
				>
			</div>
			<button class="bg-green-400 hover:bg-green-500 p-4 w-full lg:w-64 rounded-xl">Imposta</button>
		</div>
	{:else}
		<button class="bg-blue-400 hover:bg-blue-500 w-full h-full rounded-xl" on:click={() => (editing = true)}
			>Imposta regola veloce</button
		>
	{/if}
</div>