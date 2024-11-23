// Porca Madonna Query Language
// Avoiding all the shitty graphql libraries out there
import { createClient } from 'graphql-ws';
import { popup } from './popup';
import { gql, GraphQLClient } from 'graphql-request';
import { onMount } from 'svelte';
import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';
import { PUBLIC_SERVER_HOST, PUBLIC_CLIENT_HOST } from '$env/static/public';

const SERVER_HOST = PUBLIC_SERVER_HOST || 'localhost';
const CLIENT_HOST = PUBLIC_CLIENT_HOST || 'localhost';
const HOST = typeof window !== 'undefined' ? CLIENT_HOST : SERVER_HOST;

const client = {
	http: new GraphQLClient(`http://${HOST}:8080/query`),
	ws: typeof window !== 'undefined' ? createClient({ url: `ws://${HOST}:8080/query` }) : null
};

function moSonCazzi(e: any) {
	const convert = (e: any): string => {
		if (typeof e === 'string') {
			return e;
		} else if (e instanceof Error) {
			return e.message;
		} else {
			return JSON.stringify(e);
		}
	};
	if (e instanceof Array) {
		popup.set({ messages: e.map(convert) });
	} else {
		popup.set({ messages: [convert(e)] });
	}
}

async function porca<T>(query: string, variables?: any): Promise<T> {
	const queryName = query.split('{')[1].split('(')[0].trim();
	const result = await madonna<{ [key: string]: T }>(query, variables);
	return result[queryName];
}

async function madonna<T>(query: string, variables?: any): Promise<T> {
	try {
		return await client.http.request<T>(query, variables);
	} catch (e) {
		moSonCazzi(e);
		throw e;
	}
}

function madonne<T>(query: string, variables?: any): Writable<T> {
	const store = writable<T>();
	onMount(async () => {
		try {
			await (async () => {
				const subscription = client.ws!.iterate({ query, variables });
				for await (const event of subscription) {
					if (event.data) {
						store.set(event.data as T);
					} else {
						moSonCazzi(event.errors);
					}
				}
			})();
		} catch (e: any) { // Typescript, so beautiful yet so ugly
			moSonCazzi(e);
		}
	});
	return store;
}

export { gql, porca, madonna, madonne };
