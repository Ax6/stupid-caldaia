<script lang="ts">
	import type { BoilerData, Rule } from '$lib/types';
	import type { Readable } from 'svelte/store';
	import Regola from './Regola.svelte';
	export let boilerSubscription: Readable<BoilerData>;

	$: rules = $boilerSubscription.boiler.rules;
	$: minTemp = $boilerSubscription.boiler.minTemp;
	$: maxTemp = $boilerSubscription.boiler.maxTemp;
	$: emptyRuleExists = rules.some((rule) => rule.id === undefined);

	function spawnRule() {
		if (!emptyRuleExists) {
			rules = [...rules, {} as Rule];
		}
	}

	function removeRule(event: CustomEvent<string | null>) {
		if (event.detail) {
			rules = rules.filter((rule) => rule.id !== event.detail);
		} else {
			rules = rules.filter((rule) => rule.id);
		}
	}
</script>

<div class="rounded-xl bg-gray-200 border border-gray-400 p-2">
	<div class="flex">
		<h1 class="text-4xl font-thin flex-grow">Regola</h1>
		{#if !emptyRuleExists}
			<button class="bg-blue-400 hover:bg-blue-500 border border-blue-600 rounded-lg p-2 text-xl" on:click={spawnRule}>
				Aggiungi
			</button>
		{/if}
	</div>

	{#if rules.length === 0}
		<i class="text-xl font-thin">Nessuna regola impostata</i>
	{/if}
	{#each { length: rules.length } as _, index}
		{@const reverseIndex = rules.length - 1 - index}
		{@const rule = rules[reverseIndex]}
		<Regola {rule} {minTemp} {maxTemp} on:remove={removeRule} />
	{/each}
</div>
