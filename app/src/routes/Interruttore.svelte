<script lang="ts">
	import { gql, madonna } from '$lib/porca-madonna-ql';
	export let data: { switch: { state: State } };

	type State = 'ON' | 'OFF';
	type SetSwitch = { setSwitch: State };

	async function handleClick() {
		const result = await madonna<SetSwitch>(
			gql`
				mutation setSwitch($state: State!) {
					setSwitch(state: $state)
				}
			`,
			{
				state: data.switch.state === 'ON' ? 'OFF' : 'ON'
			}
		);
		data.switch.state = result.setSwitch;
	}
</script>

<button
	class="m-2 p-2 grid place-items-center rounded-xl {data.switch.state.toLowerCase() || 'unknown'}"
	on:click={handleClick}
>
	<p class="text-xl">Caldaia</p>
	<p class="text-6xl">
		{data.switch.state}
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
