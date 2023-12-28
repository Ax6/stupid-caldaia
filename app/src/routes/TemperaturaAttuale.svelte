<script lang="ts">
	import { GraphQLClient } from 'kikstart-graphql-client';
	
	const client = new GraphQLClient({
		url: 'http://localhost:8080/query',
		wsUrl: 'ws://localhost:8080/query',
	})

	const query = `
		subscription temperatureChange($position: String!) {
			onTemperatureChange(position: "centrale")
		}
	`

	client.runSubscription(query).subscribe({
		next: res => console.log(JSON.stringify(res.data.statusSubscription, null, 2)),
		error: error => console.error(error),
		complete: () => console.log('done'),
	})
</script>

<div class="w-full bg-green-300 m-2 p-2 grid place-items-center rounded-xl">
	<p class="text-xl">Temperatura attuale</p>

	<p class="text-6xl">
	</p>
</div>
