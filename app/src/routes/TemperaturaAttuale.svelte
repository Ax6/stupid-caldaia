<script lang="ts">
	import type { PageData, TemperatureData } from './+page.server';
	import { gql, madonne } from '$lib/porca-madonna-ql';
	export let data: PageData;

	let subscription = madonne<TemperatureData>(gql`
		subscription {
			sensor(name: "temperatura", position: "centrale") {
				value
				timestamp
			}
		}
	`);
	subscription.set({ temperature: data.temperature });
</script>

<div class="pb-6 pt-4 bg-gray-400 grid place-items-center rounded-xl">
	<p class="text-xl">Temperatura attuale</p>
	<p class="text-6xl">
		{#if $subscription.temperature?.value}
			{$subscription.temperature.value.toFixed(1)} °C
		{:else}
			Boh?
		{/if}
	</p>
</div>
