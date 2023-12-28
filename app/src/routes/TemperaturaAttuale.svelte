<script lang="ts">
	import type { PageData, TemperatureData } from './+page.server';
	import { gql, madonne } from '$lib/porca-madonna-ql';
	export let data: PageData;

	let subscription = madonne<TemperatureData>(gql`
		subscription {
			temperature(position: "centrale") {
				value
				timestamp
			}
		}
	`);
	subscription.set({ temperature: data.temperature });
</script>

<div class="bg-gray-400 m-2 p-2 grid place-items-center rounded-xl">
	<p class="text-xl">Temperatura attuale</p>
	<p class="text-6xl">
		{$subscription.temperature.value.toFixed(1)} °C
	</p>
</div>
