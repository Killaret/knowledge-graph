<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  onMount(() => {
    // 3D functionality frozen for v1 - redirecting to 2D graph
    setTimeout(() => {
      goto("/graph");
    }, 500);
  });
</script>

<div class="page">
  <div class="center">
    <div class="frozen-notice">
      <h2>{t("graph.frozenTitle")}</h2>
      <p>{t("graph.frozenMessage1")}</p>
      <p>{t("graph.frozenMessage2")}</p>
      <div class="spinner"></div>
    </div>
  </div>
</div>

<style>
  .page {
    position: relative;
    width: 100%;
    height: 100vh;
    overflow: hidden;
    background: #050510;
  }

  .center {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: white;
    gap: 16px;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255, 255, 255, 0.3);
    border-top-color: #88aaff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
