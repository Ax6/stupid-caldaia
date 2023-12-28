<script lang="ts">
	import { gql, porca, madonne } from '$lib/porca-madonna-ql';
	import type { PageData, BoilerData, Boiler } from './+page.server';
	export let data: PageData;

	let subscription = madonne<BoilerData>(gql`
		subscription {
			boiler {
				state
			}
		}
	`);
	subscription.set({ boiler: data.boiler });

	async function handleClick() {
		const result = await porca<Boiler>(
			gql`
				mutation setState($state: State!) {
					updateBoiler(config: { state: $state }) {
						state
					}
				}
			`,
			{
				state: $subscription.boiler.state === 'ON' ? 'OFF' : 'ON'
			}
		);
		subscription.set({ boiler: result });
	}
</script>

<button
	class="pb-6 pt-4 grid place-items-center rounded-xl {$subscription.boiler.state.toLowerCase() ||
		'unknown'}"
	on:click={handleClick}
>
	<p class="text-xl">Caldaia</p>
	<p class="text-6xl">
		{#if $subscription.boiler.state === 'ON'}
			Accesa
		{:else if $subscription.boiler.state === 'OFF'}
			Spenta
		{:else}
			Boh?
		{/if}
	</p>
</button>

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
