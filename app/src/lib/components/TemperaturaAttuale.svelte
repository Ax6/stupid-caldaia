<script lang="ts">
	import { subscriptionStore, gql, getContextClient } from '@urql/svelte';

	const temperatura = subscriptionStore({
		client: getContextClient(),
		query: gql`
			subscription Temperatura {
				onTemperatureChange(position: "centrale")
			}
		`
	});
</script>

<div class='w-full bg-green-300 m-2 p-2 grid place-items-center rounded-xl'>
    <p class='text-xl'> Temperatura attuale </p>

	<p class='text-6xl'>
		{#if !$temperatura.data}
			<p> - </p>
		{:else}
			<p> { parseFloat($temperatura.data.onTemperatureChange).toFixed(1) } °C </p>
		{/if}
	</p>
</div>