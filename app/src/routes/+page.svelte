<script lang="ts">
	import Interruttore from './Interruttore.svelte';
	import TemperaturaAttuale from './TemperaturaAttuale.svelte';
	export let data: any;
</script>

<body class="flex flex-col items-center">
	<TemperaturaAttuale />
	<Interruttore {data} />
</body>
