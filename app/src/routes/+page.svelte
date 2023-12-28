<script lang="ts">
	import type { PageData, BoilerData, SensorRangeData } from './+page.server';
	import TemperaturaAttuale from './TemperaturaAttuale.svelte';
	import Interruttore from './Interruttore.svelte';
	import Grafico from './Grafico.svelte';
	import Regole from './Regole.svelte';
	import Target from './Target.svelte';

	export let data: PageData;
</script>

<div class="w-full mb-10">
	<div class="w-full p-2 pt-4">
		<Target {data} />
	</div>
	<div class="w-full p-2 grid grid-cols-1 lg:grid-cols-2 gap-4">
		<TemperaturaAttuale {data} />
		<Interruttore {data} />
	</div>
	<Regole {data} />
	<Grafico {data} />
</div>
