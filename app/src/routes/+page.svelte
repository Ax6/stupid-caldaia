<script lang="ts">
	import type { PageData, BoilerData, Measure, ColorBand } from '$lib/types';
	import { madonne, gql } from '$lib/porca-madonna-ql';
	import TemperaturaAttuale from './TemperaturaAttuale.svelte';
	import Interruttore from './Interruttore.svelte';
	import Grafico from './Grafico.svelte';
	import Regole from './Regole.svelte';
	import RegolaVeloce from './RegolaVeloce.svelte';
	import Popup from './Popup.svelte';
	import { reduce, type Color } from 'd3';

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

	let now = new Date().toISOString();
	let bands = $derived(
		([] as ColorBand[]).concat(
			data.switchHistory.reduce((acc, curr) => {
				const prevSet = acc.length > 0 && acc[acc.length - 1].to === now;
				if (curr.state === 'ON' && !prevSet) {
					acc.push({ from: curr.time, to: now, color: 'orange' });
				} else if (curr.state !== 'ON' && prevSet) {
					acc[acc.length - 1].to = curr.time;
				}
				return acc;
			}, [] as ColorBand[]),
			data.overheatingProtectionHistory.reduce((acc, curr) => {
				const prevSet = acc.length > 0 && acc[acc.length - 1].to === now;
				if (curr.isActive && !prevSet) {
					acc.push({ from: curr.time, to: now, color: 'gray' });
				} else if (!curr.isActive && prevSet) {
					acc[acc.length - 1].to = curr.time;
				}
				return acc;
			}, [] as ColorBand[])
		) as ColorBand[]
	);
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
		{bands}
		yLabel="Temperatura"
		yUnit="°C"
		title="Grafico Temperatura"
		height={200}
	/>
	<section class="m-2"></section>
	<Grafico
		data={data.humiditySeries.map((d) => ({ time: d.time, value: d.value / 1000000 }))}
		{bands}
		yLabel="Umidità"
		yUnit="%"
		title="Grafico Umidità"
		height={200}
	/>
	<section class="m-2"></section>
	<Regole {boilerSubscription} />
</div>
