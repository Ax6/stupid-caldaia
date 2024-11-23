<script lang="ts">
	import type { PageData, SensorData } from '$lib/types';
	import { gql, madonne } from '$lib/porca-madonna-ql';
	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	let currentOutsideTemp = $derived.by(() => {
		const now = new Date();
		const temp = data.outsideTemperatureSeries.map((d) => ({
			timestamp: new Date(d.timestamp),
			value: d.value
		}));
		const closestSample = data.outsideTemperatureSeries.reduce((prev, curr) =>
			Math.abs(new Date(curr.timestamp).getTime() - now.getTime()) <
			Math.abs(new Date(prev.timestamp).getTime() - now.getTime())
				? curr
				: prev
		);
		return closestSample.value;
	});

	let subscription = madonne<SensorData>(gql`
		subscription {
			currentTemperature: sensor(name: "temperatura", position: "centrale") {
				value
				timestamp
			}
		}
	`);
	subscription.set({ currentTemperature: data.currentTemperature });
</script>

<div class="pb-4 pt-2 bg-gray-300 grid place-items-center rounded-xl">
	<p class="text-lg">Temperatura attuale</p>
	<p class="text-5xl font-thin">
		{#if $subscription.currentTemperature?.value}
			{$subscription.currentTemperature.value.toFixed(1)} °C
		{:else}
			Boh?
		{/if}
	</p>
	<p>Fuori {currentOutsideTemp.toFixed(1)} °C</p>
</div>
