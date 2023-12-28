import { gql, madonna } from '$lib/porca-madonna-ql';

export async function load() {
    return await madonna(gql`
        query {
            switch {
                state
            }
        }
    `);
}