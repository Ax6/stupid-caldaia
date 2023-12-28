<script lang="ts">
	import { mutationStore, queryStore, gql, getContextClient } from '@urql/svelte';
	import { onMount } from 'svelte';

	let switchState = 'UNKNOWN';
    let toggleResult: any;
    let client = getContextClient();

	let setSwitch = (state: String) => {
		toggleResult = mutationStore({
			client,
			query: gql`
				mutation ($state: State!) {
					setSwitch(state: $state)
				}
			`,
			variables: { state },
		})
	};

    $: result?.subscribe((result: any) => {
        if (result.error) {
            alert(result.error);
        } else {
            toggle = !toggle;
        }
    });

    onMount(() => {
        result = queryStore({client, query: gql`query { switch { state } }`,});
        if (result.error) {
            alert(result.error);
        } else {
            switchState = result.data.switch.state;
        } 
    });

</script>

<button
	class="w-full sm:w-64 bg-green-300 m-2 p-2 grid place-items-center rounded-xl"
	on:click={() => setSwitch(switchState)}
>
	<p class="text-xl">Interruttore</p>
	<p class="text-6xl">
		{#if $result?.data}
			<p>{$result.data.setSwitch}</p>
		{/if}
	</p>
</button>
