<script lang="ts">
	import type { PageData, Boiler, Measure } from './+page.server';
	import TemperaturaAttuale from './TemperaturaAttuale.svelte';
	import Interruttore from './Interruttore.svelte';
	import Grafico from './Grafico.svelte';

	export let data: PageData;
</script>

<body class="w-full flex flex-col items-center">
	<TemperaturaAttuale {data} />
	<Interruttore {data} />
	<Grafico {data} />
</body>
