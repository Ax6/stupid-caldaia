<script>
  import "../app.css";
  import { Client, setContextClient, cacheExchange, fetchExchange, subscriptionExchange } from '@urql/svelte';

  const client = new Client({
    url: 'http://localhost:8080/query',
    fetchSubscriptions: true,
    exchanges: [
      cacheExchange,
      fetchExchange,
    ],
  });

  setContextClient(client);
</script>

<slot />