import type { Writable } from 'svelte/store';
import { writable } from 'svelte/store';

export type Popup = {
	messages: string[];
};

export const popup: Writable<Popup> = writable({messages: []});
