<script lang="ts">
	import type { Rule } from '$lib/types';
	import {} from 'date-fns';
	import { it } from 'date-fns/locale';
	export let rule: Rule | undefined;
	export let minTemp: number;
	export let maxTemp: number;

	const weekDays: string[] = ['Lun', 'Mar', 'Mer', 'Gio', 'Ven', 'Sab', 'Dom'];
	$: temperature = rule ? rule.targetTemp : 20;
	let startTime: string = '18:00';
	let duration: string = '1 ora';
	let repeatDays: number[] = [1, 2, 3, 4, 5];

	function toggleRepeatDay(day: number) {
		if (repeatDays.indexOf(day) > -1) {
			repeatDays = repeatDays.filter((d) => d !== day);
		} else {
			repeatDays = [...repeatDays, day];
		}
	}
</script>

<div class="p-4 rounded-xl border border-black border-lg">
	<div class="text-xl mb-2">
		<input type="number" bind:value={temperature} class="w-11" min={minTemp} max={maxTemp}/>°C
		dalle <input type="time" bind:value={startTime} />
		alle <input type="time" bind:value={duration} />
		e si ripete
	</div>
	<div class="grid grid-cols-7 place-items-center">
		{#each weekDays as day, i}
			<button
				on:click={() => toggleRepeatDay((i + 8) % 7)}
				class="w-12 h-12 rounded-full grid place-items-center color-white {repeatDays.indexOf(
					(i + 8) % 7
				) >= 0
					? 'bg-blue-300'
					: 'bg-gray-200'}"
			>
				{day}
			</button>
		{/each}
	</div>
</div>
