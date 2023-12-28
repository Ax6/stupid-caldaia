<script lang="ts">
	import { gql, madonne } from '$lib/porca-madonna-ql';

	type TemperatureChange = {
		onTemperatureChange: number;
	};

	let listener = madonne<TemperatureChange>(
		gql`
			subscription {
				onTemperatureChange(position: "centrale")
			}
		`
	);
</script>

<div class="bg-gray-400 m-2 p-2 grid place-items-center rounded-xl">
	<p class="text-xl">Temperatura attuale</p>
	<p class="text-6xl">
		{#if $listener}
			{$listener.onTemperatureChange.toFixed(1)} °C
		{/if}
	</p>
</div>
