<script lang="ts">
	import type { PageData, SensorData } from '$lib/types';
	import { gql, madonne } from '$lib/porca-madonna-ql';
	export let data: PageData;

	let subscription = madonne<SensorData>(gql`
		subscription {
			sensor(name: "temperatura", position: "centrale") {
				value
				timestamp
			}
		}
	`);
	subscription.set({ sensor: data.sensor });
</script>

<div class="pb-4 pt-2 bg-gray-300 grid place-items-center rounded-xl">
	<p class="text-lg">Temperatura attuale</p>
	<p class="text-5xl font-thin">
		{#if $subscription.sensor?.value}
			{$subscription.sensor.value.toFixed(1)} °C
		{:else}
			Boh?
		{/if}
	</p>
</div>
