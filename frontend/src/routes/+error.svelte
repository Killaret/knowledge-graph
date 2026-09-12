<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { getCurrentLocale, formatMessage } from "$shared/utils/i18n";
  import StateIllustration from "$components/atoms/StateIllustration.svelte";

  const locale = getCurrentLocale();

  const illustrationType =
    $page.status === 404
      ? "404"
      : $page.status >= 500 && $page.status < 600
        ? "server-error"
        : "error";

  function t(key: string, params?: Record<string, string | number>) {
    return formatMessage(key, locale, params);
  }

  function goHome() {
    goto("/");
  }

  function reload() {
    if (typeof window !== "undefined") {
      window.location.reload();
    }
  }

  const isDev = import.meta.env.DEV;
  const pageError = $page.error as Error | null | undefined;
</script>

<div class="error-page" role="alert" aria-live="assertive" data-testid="error-page">
  <div class="error-container">
    <div class="error-illustration">
      <StateIllustration type={illustrationType} />
    </div>
    <h1 class="error-title">
      {#if $page.status === 500}
        {t("error.500.title")}
      {:else}
        {t("error.unknown.title", { status: $page.status })}
      {/if}
    </h1>
    <p class="error-message">
      {#if $page.status === 500}
        {t("error.500.message")}
      {:else}
        {t("error.unknown.message")}
      {/if}
    </p>
    <div class="error-actions">
      <button class="btn btn-primary" onclick={reload} type="button">
        {t("error.500.retry")}
      </button>
      <button class="btn btn-secondary" onclick={goHome} type="button">
        {t("error.500.goHome")}
      </button>
    </div>
    {#if isDev && pageError}
      <div class="dev-panel">
        <p class="dev-hint">{t("error.devHint")}</p>
        <pre class="dev-details">{pageError.message || String(pageError)}</pre>
        {#if pageError.stack}
          <pre class="dev-stack">{pageError.stack}</pre>
        {/if}
      </div>
    {/if}
  </div>
</div>

<style>
  .error-page {
    position: fixed;
    inset: 0;
    z-index: 900;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--gradient-cosmic-bg);
    color: var(--color-text);
    padding: 2rem;
    overflow-y: auto;
  }

  .error-container {
    max-width: 640px;
    width: 100%;
    text-align: center;
  }

  .error-illustration {
    margin: 0 auto 1.5rem;
    max-width: 260px;
  }

  .error-title {
    font-size: 2.5rem;
    font-weight: 700;
    margin: 0 0 1rem;
    color: #ff6b6b;
    text-shadow: 0 0 12px rgba(255, 107, 107, 0.4);
  }

  .error-message {
    font-size: 1.125rem;
    color: var(--color-text-secondary);
    margin: 0 0 2rem;
    line-height: 1.6;
  }

  .error-actions {
    display: flex;
    gap: 1rem;
    justify-content: center;
    flex-wrap: wrap;
  }

  .btn {
    padding: 0.75rem 1.5rem;
    border-radius: 0.5rem;
    border: none;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition:
      transform 0.2s ease,
      box-shadow 0.2s ease;
  }

  .btn:hover {
    transform: translateY(-2px);
  }

  .btn-primary {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: #ffffff;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
  }

  .btn-secondary {
    background: rgba(255, 255, 255, 0.1);
    color: var(--color-text);
    border: 1px solid rgba(255, 255, 255, 0.2);
  }

  .btn-secondary:hover {
    background: rgba(255, 255, 255, 0.15);
  }

  .dev-panel {
    margin-top: 2rem;
    text-align: left;
    background: rgba(0, 0, 0, 0.4);
    border-radius: 0.5rem;
    padding: 1rem;
    border: 1px solid rgba(255, 107, 107, 0.3);
  }

  .dev-hint {
    color: #ff6b6b;
    font-weight: 600;
    margin: 0 0 0.5rem;
  }

  .dev-details,
  .dev-stack {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 0.875rem;
    white-space: pre-wrap;
    word-break: break-word;
    color: var(--color-text-secondary);
    background: rgba(0, 0, 0, 0.3);
    border-radius: 0.25rem;
    padding: 0.75rem;
    margin: 0 0 0.5rem;
    overflow-x: auto;
  }

  .dev-stack {
    color: #ff6b6b;
  }
</style>
