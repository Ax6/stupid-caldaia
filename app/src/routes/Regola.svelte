<script lang="ts">
	import type { Rule, RuleInput } from '$lib/types';
	import { createEventDispatcher } from 'svelte';
	import { gql, porca } from '$lib/porca-madonna-ql';
	import {
		getDay,
		sub,
		add,
		setHours,
		setMinutes,
		formatISODuration,
		formatISO,
		format,
		intervalToDuration,
		compareAsc
	} from 'date-fns';
	import { parseISODuration } from '$lib/faster-than-open-source';
	import { popup } from '$lib/popup';
	interface Props {
		rule: Rule;
		minTemp: number;
		maxTemp: number;
	}

	let { rule, minTemp, maxTemp }: Props = $props();

	const weekDays: string[] = ['Lun', 'Mar', 'Mer', 'Gio', 'Ven', 'Sab', 'Dom'];
	let input: {
		temperature: number;
		startTime: string;
		endTime: string;
		repeatDays: number[];
	} = $state({
		temperature: 20,
		startTime: '18:00',
		endTime: '01:00',
		repeatDays: [1, 2, 3, 4, 5]
	});

	let isUnsavedRule = $derived(rule.id === undefined);
	let editMode = $state(false);
	let editing = $derived(isUnsavedRule || editMode);

	const dispatch = createEventDispatcher();

	/**
	 * This function is used to put the data (usually from the current rule) into the input form
	 * @param data The data to put into the input
	 */
	function putDataToInput(data: RuleInput) {
		const startDate = new Date(data.start);
		input.startTime = format(startDate, 'HH:mm');
		input.temperature = rule.targetTemp;
		// rule.duration is a string in the ISO 8601 format. We just need hours and minutes
		// waiting on date-fns to bring this parseISODuration function https://github.com/date-fns/date-fns/pull/3151
		input.endTime = format(add(startDate, parseISODuration(rule.duration)), 'HH:mm');
		input.repeatDays = rule.repeatDays;
	}

	/** This function parses the input from the form and puts it into the
	 * returned object so that it's ready to be sent to the server
	 */
	function putInputToData(): RuleInput {
		const [startHour, startMinute] = input.startTime.split(':').map((n) => parseInt(n));
		let start = new Date();
		for (let previousDay = 0; previousDay < 7; previousDay += 1) {
			const exists = input.repeatDays.filter((day) => day === getDay(start)).length > 0;
			if (exists) {
				break;
			}
			start = sub(start, { days: 1 });
		}
		start = setHours(start, startHour);
		start = setMinutes(start, startMinute);
		const isoStart = formatISO(start);

		const [endHour, endMinute] = input.endTime.split(':').map((n) => parseInt(n));
		let end = new Date();
		end = setHours(end, endHour);
		end = setMinutes(end, endMinute);
		if (compareAsc(start, end) > 0) {
			end = add(end, { days: 1 });
		}
		const duration = intervalToDuration({ start, end });
		const isoDuration = formatISODuration(duration);
		return {
			start: isoStart,
			duration: isoDuration,
			// No delay for regular rules
			delay: formatISODuration({ seconds: 0 }),
			targetTemp: input.temperature,
			repeatDays: input.repeatDays
		};
	}

	function toggleRepeatDay(day: number) {
		if (input.repeatDays.indexOf(day) > -1) {
			input.repeatDays = input.repeatDays.filter((d) => d !== day);
		} else {
			input.repeatDays = [...input.repeatDays, day];
		}
	}

	function annulla() {
		if (rule.id) {
			editMode = false;
			putDataToInput(rule);
		} else {
			// If the rule is new, we just remove it
			dispatch('cancel', null);
		}
	}

	async function salva() {
		// Start date happens on any day of repeatDays but must be in the past (or today)
		// This assures that the rule applies only with repeat days set
		const data = putInputToData();
		if (data.repeatDays.length === 0) {
			popup.set({ messages: ['Devi selezionare almeno un giorno'] });
			return;
		}
		await porca(gql`
			mutation {
				setRule(
					${rule.id ? `id: "${rule.id}"` : ''}
					start: "${data.start}"
					duration: "${data.duration}"
					delay: "${data.delay}"
					repeatDays: [${data.repeatDays.join(', ')}]
					targetTemp: ${data.targetTemp}
				) {
					id
				}
			}
		`);
		editMode = false;
	}

	async function elimina() {
		console.log('Elimina', rule.id);
		await porca(gql`
			mutation {
				deleteRule(id: "${rule.id}")
			}
		`);
	}

	let mainColours = $derived(editing ? 'border-orange-400 bg-orange-300' : 'border-black bg-white');
	let inputColours = $derived(editing ? 'bg-orange-200' : 'bg-inherit');

	$effect(() => {
		rule.id && putDataToInput(rule);
	});
</script>

<div class="p-2 rounded-xl border {mainColours} relative mt-4">
	<div class="absolute w-full h-full top-0 left-0 {editing ? 'hidden' : 'block'}"></div>
	<button
		class="absolute top-1 right-1 bg-orange-300 hover:bg-orange-400 rounded-full p-1"
		onclick={() => (editMode = true)}
	>
		{editing ? '' : '🖊️'}
	</button>
	<section class="m-2">
		<div class="text-l mb-2 flex place-items-end">
			<input
				type="number"
				disabled={!editing}
				bind:value={input.temperature}
				class="w-11 {inputColours} rounded px-0.5"
				max={maxTemp}
				min={minTemp}
			/>
			<p class="text-sm">°C dalle</p>
			<input
				type="time"
				disabled={!editing}
				bind:value={input.startTime}
				class="{inputColours} rounded mx-0.5"
			/>
			<p class="text-sm">alle</p>
			<input
				type="time"
				disabled={!editing}
				bind:value={input.endTime}
				class="{inputColours} rounded mx-0.5"
			/>
			<p class="text-sm">si ripete</p>
		</div>
		<div class="grid grid-cols-7 place-items-center max-w-md">
			{#each weekDays as day, i}
				{@const repeatsOnThisDay = input.repeatDays.indexOf((i + 8) % 7) >= 0}
				{#if repeatsOnThisDay || editing}
					<button
						onclick={() => toggleRepeatDay((i + 8) % 7)}
						class="w-11 h-11 sm:w-14 sm:h-14 rounded-full grid place-items-center color-white {repeatsOnThisDay
							? 'bg-blue-400 hover:bg-blue-500'
							: 'bg-gray-200 hover:bg-gray-300'}"
					>
						{day}
					</button>
				{/if}
			{/each}
		</div>
	</section>
	{#if editing}
		<div class="flex justify-between mt-6">
			{#if rule.id}
				<button class="bg-red-400 hover:bg-red-500 rounded-lg p-2 mr-2" onclick={elimina}>
					Elimina
				</button>
			{/if}
			<button class="bg-green-400 hover:bg-green-500 rounded-lg p-2 flex-grow mr-2" onclick={salva}>
				Salva
			</button>
			<button class="bg-gray-400 hover:bg-gray-500 rounded-lg p-2" onclick={annulla}>
				Annulla
			</button>
		</div>
	{/if}
</div>
