<script lang="ts">
	import type { Rule } from '$lib/types';
	import Regola from './Regola.svelte';
	export let rules: Rule[] = [];
	export let maxTemp: number;
	export let minTemp: number;

	function addRule() {
		rules = [...rules, {} as Rule];
	}

	function removeRule(event: CustomEvent<number>) {
		rules = rules.filter((_, i) => i !== event.detail);
	}
</script>

<div class="m-2 rounded-xl">
	<h1 class="text-4xl font-bold">Regole</h1>
	{#if rules.length === 0}
		<p class="text-xl font-semibold">Nessuna regola impostata</p>
	{/if}
	<button class="bg-blue-400 rounded-lg p-2 w-full my-2" on:click={addRule}>
		Aggiungi una regola
	</button>
	{#each rules as rule, ruleIndex}
		<Regola {rule} {minTemp} {maxTemp} {ruleIndex} on:remove={removeRule} />
	{/each}
</div>
