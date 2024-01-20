<script lang="ts">
	import type { Rule } from '$lib/types';
	import { createEventDispatcher } from 'svelte';
	import {} from 'date-fns';
	import { it } from 'date-fns/locale';
	export let ruleIndex: number;
	export let rule: Rule;
	export let minTemp: number;
	export let maxTemp: number;

	const weekDays: string[] = ['Lun', 'Mar', 'Mer', 'Gio', 'Ven', 'Sab', 'Dom'];
	let temperature = rule.id ? rule.targetTemp : 20;
	let startTime: string = '18:00';
	let duration: string = '1 ora';
	let repeatDays: number[] = [1, 2, 3, 4, 5];
	let editing = rule !== undefined;

	const dispatch = createEventDispatcher();

	function toggleRepeatDay(day: number) {
		if (repeatDays.indexOf(day) > -1) {
			repeatDays = repeatDays.filter((d) => d !== day);
		} else {
			repeatDays = [...repeatDays, day];
		}
	}

	function annulla() {
		if (rule.id) {
			temperature = rule.targetTemp;
			startTime = rule.start;
			duration = rule.duration;
			repeatDays = rule.repeatDays;
			editing = false;
		} else {
			dispatch('remove', ruleIndex);
		}
	}

	function salva() {}

	function elimina() {}
	$: colours = editing ? 'border-orange-500 bg-orange-300' : 'border-black bg-white';
</script>

<div class="p-4 rounded-xl border-2 {colours} relative mt-4">
	<div class="absolute w-full h-full top-0 left-0 {editing ? 'hidden' : 'block'}" />
	<button
		class="absolute top-1 right-1 bg-orange-300 rounded-full p-1"
		on:click={() => (editing = true)}
	>
		{editing ? '' : '🖊️'}
	</button>
	<div class="text-xl mb-2 flex place-items-center">
		<input
			type="number"
			bind:value={temperature}
			class="w-11 bg-inherit rounded"
			max={maxTemp}
			min={minTemp}
		/>
		<p class="text-lg">°C dalle</p>
		<input type="time" bind:value={startTime} class="bg-inherit rounded" />
		<p class="text-lg">alle</p>
		<input type="time" bind:value={duration} class="bg-inherit rounded" />
		<p class="text-lg">si ripete</p>
	</div>
	<div class="grid grid-cols-7 place-items-center">
		{#each weekDays as day, i}
			<button
				on:click={() => toggleRepeatDay((i + 8) % 7)}
				class="w-12 h-12 rounded-full grid place-items-center color-white {repeatDays.indexOf(
					(i + 8) % 7
				) >= 0
					? 'bg-blue-400'
					: 'bg-gray-200'}"
			>
				{day}
			</button>
		{/each}
	</div>
	{#if editing}
		<div class="flex justify-between mt-10">
			{#if rule.id}
				<button class="bg-red-400 rounded-lg p-2 mr-2" on:click={elimina}> Elimina </button>
			{/if}
			<button class="bg-green-400 rounded-lg p-2 flex-grow mr-2" on:click={salva}> Salva </button>
			<button class="bg-gray-400 rounded-lg p-2" on:click={annulla}> Annulla </button>
		</div>
	{/if}
</div>
