<script lang="ts">
	import {} from 'date-fns';
	import { it } from 'date-fns/locale';

	export let data;

	const weekDays: string[] = ['Lun', 'Mar', 'Mer', 'Gio', 'Ven', 'Sab', 'Dom'];
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
		Dalle <input type="time" bind:value={startTime} />
		alle <input type="time" bind:value={duration} />
		si ripete
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
