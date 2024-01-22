<script lang="ts">
	import type { PageData, BoilerData } from '$lib/types';
	import { madonne, gql } from '$lib/porca-madonna-ql';
	import TemperaturaAttuale from './TemperaturaAttuale.svelte';
	import Interruttore from './Interruttore.svelte';
	import Grafico from './Grafico.svelte';
	import Regole from './Regole.svelte';
	import RegolaVeloce from './RegolaVeloce.svelte';
	import Popup from './Popup.svelte';

	export let data: PageData;

	let boilerSubscription = madonne<BoilerData>(gql`
		subscription {
			boiler {
				state
				rules {
					id
					start
					duration
					targetTemp
					repeatDays
					isActive
					stoppedTime
				}
			}
		}
	`);
	boilerSubscription.set(data);

	$: console.log('Subscription update', $boilerSubscription.boiler);
</script>

<div class="w-full mb-10 p-2 relative">
	<Popup />
	<div class="w-full grid grid-cols-1 lg:grid-cols-2 gap-2">
		<TemperaturaAttuale {data} />
		<Interruttore {boilerSubscription} />
	</div>
	<section class="m-2"/>
	<RegolaVeloce {boilerSubscription} />
	<section class="m-2"/>
	<Grafico {data} />
	<section class="m-2"/>
	<Regole {boilerSubscription} />
</div>
