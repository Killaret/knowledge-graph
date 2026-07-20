<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  let currentNoteId: string | null = $state(null);

  onMount(() => {
    const id = $page.params.id as string;
    currentNoteId = id;

    // 3D functionality frozen for v1 - redirecting to 2D graph
    setTimeout(() => {
      if (id) {
        goto(`/graph/${id}`);
      } else {
        goto("/graph");
      }
    }, 500);
  });
</script>

<div class="page">
  <div class="center">
    <div class="frozen-notice">
      <h2>{t("graph.frozenTitle")}</h2>
      <p>{t("graph.frozenMessage1")}</p>
      <p>
        {t("graph.frozenRedirectWithId", {
          id: currentNoteId || t("graph.entireGraph"),
        })}
      </p>
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
    align-items: center;
    justify-content: center;
    height: 100%;
    color: white;
  }

  .frozen-notice {
    text-align: center;
    max-width: 500px;
    padding: 40px;
    background: rgba(0, 0, 0, 0.85);
    border-radius: 16px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    backdrop-filter: blur(10px);
  }

  .frozen-notice h2 {
    color: #88aaff;
    margin-bottom: 16px;
    font-size: 24px;
    font-weight: 600;
  }

  .frozen-notice p {
    margin-bottom: 16px;
    line-height: 1.6;
    color: #94a3b8;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255, 255, 255, 0.3);
    border-top-color: #88aaff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 20px auto 0;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
