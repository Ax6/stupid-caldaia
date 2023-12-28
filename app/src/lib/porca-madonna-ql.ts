// Porca Madonna Query Language
// Avoiding all the shitty graphql libraries out there
import { createClient } from 'graphql-ws';
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

async function porca<T>(query: string, variables?: any): Promise<T> {
	const queryName = query.split('{')[1].split('(')[0].trim();
	const response = await madonna<{ [key: string]: T }>(query, variables);
	return response[queryName];
}

async function madonna<T>(query: string, variables?: any): Promise<T> {
	return await client.http.request<T>(query, variables);
}

function madonne<T>(query: string, variables?: any): Writable<T> {
	const store = writable<T>();
	onMount(() => {
		(async () => {
			const subscription = client.ws!.iterate({ query, variables });
			for await (const event of subscription) {
				if (event.data) {
					store.set(event.data as T);
				} else {
                    console.error(event.errors);
                }
			}
		})();
	});
	return store;
}

export { gql, porca, madonna, madonne };
