<script lang="ts">
	import { gql, GraphQLClient } from 'graphql-request';
	import { onMount } from 'svelte';

	let data: any;

	const client = new GraphQLClient('http://localhost:8080/query');

	const switchState = gql`
		query SwitchState {
			switch {
				state
			}
		}
	`;

	onMount(async () => data = await client.request(switchState));
</script>

<button class="w-full sm:w-64 bg-green-300 m-2 p-2 grid place-items-center rounded-xl">
	<p class="text-xl">Caldaia</p>
	<p class="text-6xl">
        { data?.switch.state }
	</p>
</button>
