<script lang="ts">
	import type { Rule } from '$lib/types';
	import Regola from './Regola.svelte';
	export let rules: Rule[] = [];
	export let maxTemp: number;
	export let minTemp: number;

	function addRule() {
		rules = [...rules, {} as Rule];
	}

</script>

<div class="p-4">
	<h1 class="text-4xl font-bold">Regole</h1>
	{#if rules.length === 0}
		<p class="text-xl font-semibold">Nessuna regola impostata</p>
	{/if}
	<button class="bg-blue-500 text-white rounded-lg p-2" on:click={addRule}>
		Aggiungi regola
	</button>
	{#each rules as rule}
		<Regola {minTemp} {maxTemp} {rule}/>
	{/each}
</div>