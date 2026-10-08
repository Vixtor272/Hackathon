<script lang="ts">
  import { route } from './lib/router';
  import TopNav from './lib/ui/layout/TopNav.svelte';
  import NotFound from './lib/ui/NotFound.svelte';
  import CheckoutPage from './lib/ui/checkout/CheckoutPage.svelte';
  import DeUnaPage from './lib/ui/deuna/DeUnaPage.svelte';
  import OperationsPage from './lib/ui/operations/OperationsPage.svelte';
  import WhatsAppPage from './lib/ui/whatsapp/WhatsAppPage.svelte';
</script>

<TopNav />
<main class="page">
  {#if $route.name === 'whatsapp'}
    <WhatsAppPage />
  {:else if $route.name === 'checkout'}
    {#key $route.orderId}
      <CheckoutPage orderId={$route.orderId} />
    {/key}
  {:else if $route.name === 'deuna'}
    {#key $route.paymentId}
      <DeUnaPage paymentId={$route.paymentId} />
    {/key}
  {:else if $route.name === 'operations'}
    <OperationsPage />
  {:else}
    <NotFound path={$route.path} />
  {/if}
</main>
