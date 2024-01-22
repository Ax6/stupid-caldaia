<script lang="ts">
	import { popup } from '$lib/popup';

	let show: boolean = false;
	let messages: string[] = [];

	popup.subscribe((data) => {
		messages = data.messages;
		show = messages.length > 0;
	});
</script>

{#if show}
	<div class="fixed w-full top-0 left-0 p-4 z-50">
		<div class="rounded-xl border-4 border-red-500 bg-red-400 grid grid-cols-4">
			<p class="flex-grow m-2 col-span-3 text-wrap break-words overflow-y-auto max-h-46">
				Errore:<br />
				{#each messages as message, index}
					{index + 1}: {message}<br />
				{/each}
			</p>
			<button
				class="h-20 bg-red-500 m-2 rounded-2xl text-4xl text-white border-red-800 border-2"
				on:click={() => (show = false)}>🙆🏿</button
			>
		</div>
	</div>
{/if}
