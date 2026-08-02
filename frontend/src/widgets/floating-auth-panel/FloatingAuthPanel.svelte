<script lang="ts">
  import { browser } from "$app/environment";
  import LoginForm from "$components/organisms/LoginForm.svelte";
  import RegisterForm from "$components/organisms/RegisterForm.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  interface Props {
    open: boolean;
    initialTab?: "login" | "register";
    onClose: () => void;
    onSuccess?: () => void;
  }
  let {
    open,
    initialTab = "login",
    onClose,
    onSuccess,
  }: Props = $props();

  let activeTab = $state<"login" | "register">("login");
  let panelEl: HTMLDivElement | null = $state(null);
  let dragging = $state(false);
  let startX = $state(0);
  let startY = $state(0);
  let posX = $state(0);
  let posY = $state(0);
  let panelWidth = $state(0);
  let panelHeight = $state(0);

  $effect(() => {
    // Reset position and active tab when (re-)opened.
    if (open) {
      activeTab = initialTab;
      posX = 0;
      posY = 0;
    }
  });

  function startDrag(e: PointerEvent) {
    if (!browser || !panelEl) return;
    e.preventDefault();
    dragging = true;
    startX = e.clientX - posX;
    startY = e.clientY - posY;
    const rect = panelEl.getBoundingClientRect();
    panelWidth = rect.width;
    panelHeight = rect.height;
    window.addEventListener("pointermove", handleDrag);
    window.addEventListener("pointerup", stopDrag);
  }

  function handleDrag(e: PointerEvent) {
    if (!dragging || !browser) return;
    let nextX = e.clientX - startX;
    let nextY = e.clientY - startY;

    // Keep at least a 40 px grab-handle visible in the viewport.
    const maxX = window.innerWidth - 40;
    const maxY = window.innerHeight - 40;
    const minX = -(panelWidth - 40);
    const minY = -(panelHeight - 40);

    posX = Math.max(minX, Math.min(maxX, nextX));
    posY = Math.max(minY, Math.min(maxY, nextY));
  }

  function stopDrag() {
    dragging = false;
    if (browser) {
      window.removeEventListener("pointermove", handleDrag);
      window.removeEventListener("pointerup", stopDrag);
    }
  }

  function handleSuccess() {
    onClose();
    onSuccess?.();
  }

  function switchTab(tab: "login" | "register") {
    activeTab = tab;
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      onClose();
    }
  }
</script>

<svelte:window onkeydown={handleKeyDown} />

{#if open}
  <div
    bind:this={panelEl}
    class="floating-auth-panel"
    style:transform="translate({posX}px, {posY}px)"
    role="dialog"
    aria-modal="true"
    aria-labelledby="floating-auth-title"
    data-testid="floating-auth-panel"
  >
    <div
      class="panel-header"
      onpointerdown={startDrag}
      role="button"
      tabindex="-1"
      aria-label={t("auth.dragHandle")}
      data-testid="floating-auth-drag-handle"
    >
      <span class="drag-handle">⠿</span>
      <span id="floating-auth-title" class="panel-title">
        {t("auth.authTitle")}
      </span>
      <button
        type="button"
        class="close-btn"
        onclick={onClose}
        aria-label={t("common.close")}
        data-testid="floating-auth-close"
      >
        ×
      </button>
    </div>

    <div class="panel-tabs" role="tablist">
      <button
        type="button"
        role="tab"
        aria-selected={activeTab === "login"}
        class="tab"
        class:active={activeTab === "login"}
        onclick={() => switchTab("login")}
        data-testid="floating-auth-tab-login"
      >
        {t("auth.signInButton")}
      </button>
      <button
        type="button"
        role="tab"
        aria-selected={activeTab === "register"}
        class="tab"
        class:active={activeTab === "register"}
        onclick={() => switchTab("register")}
        data-testid="floating-auth-tab-register"
      >
        {t("auth.registerButton")}
      </button>
    </div>

    <div class="panel-body" role="tabpanel">
      {#if activeTab === "login"}
        <LoginForm
          onSuccess={handleSuccess}
          onRegister={() => switchTab("register")}
        />
      {:else}
        <RegisterForm
          onSuccess={handleSuccess}
          onLogin={() => switchTab("login")}
        />
      {/if}
    </div>
  </div>
{/if}

<style>
  .floating-auth-panel {
    position: fixed;
    top: 80px;
    right: 20px;
    width: 360px;
    max-width: 90vw;
    max-height: 90vh;
    overflow: auto;
    background: var(--color-surface, #1a1a2e);
    border: 1px solid var(--color-border, rgba(255, 255, 255, 0.1));
    border-radius: 1rem;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
    z-index: 100;
    display: flex;
    flex-direction: column;
    color: var(--color-text, #e0e0e0);
    font-family: var(--font-sans, system-ui, sans-serif);
  }

  .panel-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--color-border, rgba(255, 255, 255, 0.1));
    cursor: grab;
    user-select: none;
    background: var(--color-surface-elevated, #252540);
    border-radius: 1rem 1rem 0 0;
  }

  .panel-header:active {
    cursor: grabbing;
  }

  .drag-handle {
    font-size: 1rem;
    line-height: 1;
    color: var(--color-text-secondary, #9ca3af);
    cursor: grab;
  }

  .panel-title {
    flex: 1;
    font-size: 0.95rem;
    font-weight: 600;
    text-align: center;
  }

  .close-btn {
    width: 1.75rem;
    height: 1.75rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    border-radius: 0.375rem;
    color: var(--color-text, #e0e0e0);
    font-size: 1.25rem;
    cursor: pointer;
    transition: background 0.15s ease;
    flex-shrink: 0;
  }

  .close-btn:hover {
    background: var(--color-danger-light, rgba(239, 68, 68, 0.2));
    color: var(--color-danger, #ef4444);
  }

  .panel-tabs {
    display: flex;
    border-bottom: 1px solid var(--color-border, rgba(255, 255, 255, 0.1));
  }

  .tab {
    flex: 1;
    padding: 0.75rem 0.5rem;
    background: transparent;
    border: none;
    color: var(--color-text-secondary, #9ca3af);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition:
      background 0.15s ease,
      color 0.15s ease;
  }

  .tab:hover {
    background: var(--color-primary-light, rgba(59, 130, 246, 0.15));
  }

  .tab.active {
    color: var(--color-primary, #60a5fa);
    border-bottom: 2px solid var(--color-primary, #60a5fa);
    margin-bottom: -1px;
  }

  .panel-body {
    padding: 1rem;
    overflow-y: auto;
  }

  @media (max-width: 480px) {
    .floating-auth-panel {
      right: 10px;
      top: 60px;
      width: calc(100vw - 20px);
    }
  }
</style>
