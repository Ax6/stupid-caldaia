<script lang="ts">
	import type { PageData, BoilerData } from '$lib/types';
	import { madonne, gql } from '$lib/porca-madonna-ql';
	import TemperaturaAttuale from './TemperaturaAttuale.svelte';
	import Interruttore from './Interruttore.svelte';
	import Grafico from './Grafico.svelte';
	import Regole from './Regole.svelte';
	import RegolaVeloce from './RegolaVeloce.svelte';
	import Popup from './Popup.svelte';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	let boilerSubscription = madonne<BoilerData>(gql`
		subscription {
			boiler {
				state
				rules {
					id
					start
					duration
					delay
					targetTemp
					repeatDays
					isActive
					stoppedTime
				}
			}
		}
	`);
	boilerSubscription.set(data);

	$effect(() => {
		console.log('Subscription update', $boilerSubscription.boiler);
	});
</script>

<div class="w-full mb-10 p-2 relative">
	<Popup />
	<div class="w-full grid grid-cols-1 lg:grid-cols-2 gap-2">
		<TemperaturaAttuale {data} />
		<Interruttore {boilerSubscription} />
	</div>
	<section class="m-2"></section>
	<RegolaVeloce {boilerSubscription} />
	<section class="m-2"></section>
	<Grafico
		data={data.temperatureSeries}
		yLabel="Temperatura"
		yUnit="°C"
		title="Grafico Temperatura"
		height={200}
	/>
	<section class="m-2"></section>
	<Grafico
		data={data.humiditySeries.map((d) => ({ timestamp: d.timestamp, value: d.value / 1000000 }))}
		yLabel="Umidità"
		yUnit="%"
		title="Grafico Umidità"
		height={200}
	/>
	<section class="m-2"></section>
	<Regole {boilerSubscription} />
</div>
