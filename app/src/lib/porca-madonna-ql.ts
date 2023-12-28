// Porca Madonna Query Language
// Avoiding all the shitty graphql libraries out there
import { createClient } from 'graphql-ws';
import { gql } from 'graphql-request';
import { onMount } from 'svelte';
import { writable } from 'svelte/store';
import type { Writable } from 'svelte/store';

const client = typeof window !== 'undefined' ? createClient({
    url: 'ws://localhost:8080/query',
}) : null;


async function porca<T>(query: string, variables?: any): Promise<T> {
    const request = client!.iterate({ query, variables });
    const response = await request.next();
    if (response.value.data) {
        return response.value.data as T;
    } else {
        throw new Error(response.value.errors?.[0].message);
    }
}


function madonna<T>(query: string, variables?: any): Writable<T> {
    const store = writable<T>();
    onMount(() => {
        (async () => {
            const request = client!.iterate({ query, variables });
            const response = await request.next();
            store.set(response.value.data as T)
        })()
    });
    return store;

}


function madonne<T>(query: string, variables?: any): Writable<T> {
    const store = writable<T>();
    onMount(() => {
        (async () => {
            const subscription = client!.iterate({ query, variables });
            for await (const event of subscription) {
                if (event.data) {
                    store.set(event.data as T)
                }
            }
        })();
    })
    return store;
}

export {
    gql,
    madonna,
    madonne,
    porca,
};