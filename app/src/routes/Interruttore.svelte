<script lang="ts">
	import { gql, porca, madonna } from '$lib/porca-madonna-ql';

	type State = 'ON' | 'OFF';
	type SetSwitch = { setSwitch: State };

	let data = madonna<{ switch: { state: State } }>(
		gql`
			query {
				switch {
					state
				}
			}
		`,
		{}
	);

	async function handleClick() {
		const result = await porca<SetSwitch>(
			gql`
				mutation setSwitch($state: State!) {
					setSwitch(state: $state)
				}
			`,
			{
				state: $data.switch.state === 'ON' ? 'OFF' : 'ON'
			}
		);
		data.update((value) => {
			value.switch.state = result.setSwitch;
			return value;
		});
	}
</script>

<button
	class="m-2 p-2 grid place-items-center rounded-xl {$data?.switch.state.toLowerCase() ||
		'unknown'}"
	on:click={handleClick}
>
	<p class="text-xl">Caldaia</p>
	<p class="text-6xl">
		{#if $data}
			{$data.switch.state}
		{/if}
	</p>
</button>

<style>
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
