<script lang="ts">
	import { gql, porca } from '$lib/porca-madonna-ql';
	import type { BoilerData } from '$lib/types';
	import type { Readable } from 'svelte/store';
	interface Props {
		boilerSubscription: Readable<BoilerData>;
	}

	let { boilerSubscription }: Props = $props();

	async function handleClick() {
		await porca<BoilerData>(
			gql`
				mutation setState($state: State!) {
					updateBoiler(state: $state) {
						state
					}
				}
			`,
			{
				state: $boilerSubscription.boiler.state === 'ON' ? 'OFF' : 'ON'
			}
		);
	}
</script>

<div
	class="pb-4 pt-2 grid place-items-center rounded-xl {$boilerSubscription.boiler.state.toLowerCase() ||
		'unknown'}"
>
	<p class="text-lg">Stato caldaia</p>
	<p class="text-5xl font-thin">
		{#if $boilerSubscription.boiler.state === 'ON'}
			Accesa
		{:else if $boilerSubscription.boiler.state === 'OFF'}
			Spenta
		{:else}
			Boh?
		{/if}
	</p>
</div>

<style lang="postcss">
	.on {
		@apply bg-green-300;
	}

	.off {
		@apply bg-gray-300;
	}

	.unknown {
		@apply bg-red-300;
	}
</style>
