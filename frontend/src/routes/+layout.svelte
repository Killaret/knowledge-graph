<script lang="ts">
  import '$lib/styles/global.css';
  import Sidebar from '$lib/components/Sidebar.svelte';

  const { children } = $props();
</script>

<!--
  APP SHELL STRUCTURE
  Sidebar зарезервирован для будущего Context Control Center (v2.0)
  Пока скрыт (width: 0), но готов к активации
-->
<div class="app-shell">
  <Sidebar />
  <main class="app-main">
    {@render children()}
  </main>
</div>

<style>
  .app-shell {
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .app-main {
    flex: 1;
    overflow-y: auto;
    min-width: 0; /* Prevent flexbox overflow issues */
  }
</style>