<script lang="ts">
	import type { BoilerData, Rule } from '$lib/types';
	import type { Readable } from 'svelte/store';
	import Regola from './Regola.svelte';
	interface Props {
		boilerSubscription: Readable<BoilerData>;
	}

	let { boilerSubscription }: Props = $props();

	let rules = $derived(
		$boilerSubscription.boiler.rules.filter((rule) => rule.id !== 'regola-veloce')
	);
	let minTemp = $derived($boilerSubscription.boiler.minTemp);
	let maxTemp = $derived($boilerSubscription.boiler.maxTemp);
	let emptyRuleExists = $state(false);

	function spawnUnsavedRule() {
		emptyRuleExists = true;
	}

	function removeUnsavedRule() {
		emptyRuleExists = false;
	}

	let pseudoRules = $derived(emptyRuleExists ? [...rules, {} as Rule] : rules);

	$effect(() => {
		rules;
		removeUnsavedRule();
	});
</script>

<div class="rounded-xl bg-gray-200 border border-gray-300 p-2">
	<div class="flex">
		<h1 class="text-4xl font-thin flex-grow">Regola</h1>
		{#if !emptyRuleExists}
			<button
				class="bg-blue-400 hover:bg-blue-500 border border-blue-600 rounded-lg p-2 text-xl"
				onclick={spawnUnsavedRule}
			>
				Aggiungi
			</button>
		{/if}
	</div>

	{#if pseudoRules.length === 0}
		<i class="text-xl font-thin">Nessuna regola impostata</i>
	{/if}
	{#each { length: pseudoRules.length } as _, index}
		{@const reverseIndex = pseudoRules.length - 1 - index}
		{@const rule = pseudoRules[reverseIndex]}
		<Regola {rule} {minTemp} {maxTemp} on:cancel={removeUnsavedRule} />
	{/each}
</div>
