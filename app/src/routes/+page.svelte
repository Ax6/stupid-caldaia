<script lang="ts">
	import type { PageData, Boiler, Measure } from './+page.server';
	import TemperaturaAttuale from './TemperaturaAttuale.svelte';
	import { gql, madonne } from '$lib/porca-madonna-ql';
	import Interruttore from './Interruttore.svelte';

	export let data: PageData;
</script>

<body class="flex flex-col items-center">
	<TemperaturaAttuale {data} />
	<Interruttore {data} />
</body>
