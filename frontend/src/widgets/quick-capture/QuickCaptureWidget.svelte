<script lang="ts">
  import { createNote } from "$shared/api/notes";
  import { CelestialBody } from "$entities";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  // State
  let isOpen = $state(false);
  let content = $state("");
  let isSubmitting = $state(false);
  let showSuccess = $state(false);
  const MAX_TITLE_LENGTH = 200;

  // Toggle widget
  function toggle() {
    isOpen = !isOpen;
    if (isOpen) {
      content = "";
      showSuccess = false;
    }
  }

  // Submit quick capture — always type 'dust' per documentation
  async function submitCapture() {
    if (!content.trim()) return;

    isSubmitting = true;
    try {
      const title = content.slice(0, 50) + (content.length > 50 ? "..." : "");

      // Validate title length
      if (title.length > MAX_TITLE_LENGTH) {
        const truncatedTitle = content.slice(0, MAX_TITLE_LENGTH - 3) + "...";
        await createNote({
          title: truncatedTitle,
          content,
          type: CelestialBody.DUST.type,
        });
      } else {
        await createNote({
          title,
          content,
          type: CelestialBody.DUST.type,
        });
      }

      showSuccess = true;
      content = "";
      setTimeout(() => {
        showSuccess = false;
        toggle();
      }, 1000);
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error("Error creating note:", error);
      }
    } finally {
      isSubmitting = false;
    }
  }

  // Handle keyboard shortcuts
  function handleKeydown(e: KeyboardEvent) {
    // Ctrl+Shift+N to open quick capture
    if (e.key === "n" && e.ctrlKey && e.shiftKey && !isOpen) {
      e.preventDefault();
      toggle();
    }
    if (e.key === "Escape" && isOpen) {
      toggle();
    }
    if (e.key === "Enter" && (e.ctrlKey || e.metaKey) && isOpen) {
      e.preventDefault();
      submitCapture();
    }
  }

  // Close modal when clicking outside
  function handleBackdropClick(e: MouseEvent) {
    if ((e.target as HTMLElement).classList.contains("quick-capture-modal")) {
      toggle();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- Floating button -->
<div class="quick-capture-container">
  {#if isOpen}
    <div
      class="quick-capture-backdrop"
      role="dialog"
      aria-modal="true"
      tabindex="-1"
      onmousedown={handleBackdropClick}
      onkeydown={(e) => e.key === "Escape" && toggle()}
    >
      <div class="quick-capture-modal">
        <div class="modal-header">
          <h3>{t("quickCapture.title")}</h3>
          <button class="close-btn" onclick={toggle} aria-label={t("close")}>×</button>
        </div>
        <div class="modal-body">
          <textarea
            bind:value={content}
            placeholder={t("quickCapture.placeholder")}
            disabled={isSubmitting}
          ></textarea>
          {#if showSuccess}
            <div class="success-message">{t("quickCapture.saved")}</div>
          {/if}
        </div>
        <div class="modal-footer">
          <button class="cancel-btn" onclick={toggle} disabled={isSubmitting}
            >{t("quickCapture.cancel")}</button
          >
          <button
            class="submit-btn"
            onclick={submitCapture}
            disabled={isSubmitting || !content.trim()}
          >
            {isSubmitting ? t("quickCapture.saving") : t("quickCapture.save")}
          </button>
        </div>
      </div>
    </div>
  {/if}

  <button class="quick-capture-btn" onclick={toggle} title={t("quickCapture.tooltip")}> ✨ </button>
</div>

<style>
  .quick-capture-container {
    position: fixed;
    bottom: 30px;
    left: 30px;
    z-index: 1000;
  }

  .quick-capture-btn {
    width: 64px;
    height: 64px;
    padding: 0;
    border: none;
    background:
      radial-gradient(circle at 30% 30%, var(--carbon-elevated), var(--carbon-graphite) 60%, var(--carbon-black));
    color: var(--carbon-glow-cyan);
    font-size: 26px;
    line-height: 1;
    cursor: pointer;
    clip-path: polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%);
    box-shadow:
      0 0 0 2px var(--carbon-border),
      0 0 14px rgba(139, 92, 246, 0.25);
    transition: all 0.3s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    filter: drop-shadow(0 0 10px rgba(139, 92, 246, 0.2));
  }

  .quick-capture-btn:hover {
    color: var(--carbon-glow-amber);
    transform: scale(1.08) rotate(1deg);
    box-shadow:
      0 0 0 2px var(--carbon-border-active),
      0 0 22px rgba(245, 158, 11, 0.35);
    filter: drop-shadow(0 0 16px rgba(245, 158, 11, 0.3));
  }

  .quick-capture-btn:active {
    transform: scale(0.97);
  }

  /* Backdrop overlay */
  .quick-capture-backdrop {
    position: fixed;
    inset: 0;
    z-index: 1001;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(5, 5, 8, 0.65);
    backdrop-filter: blur(8px);
    animation: fadeIn 0.2s ease;
  }

  .quick-capture-modal {
    width: min(420px, calc(100vw - 40px));
    max-height: calc(100vh - 100px);
    background:
      linear-gradient(135deg, var(--carbon-graphene) 0%, var(--carbon-c70) 100%);
    border: 1px solid var(--carbon-border);
    border-radius: 16px;
    box-shadow: var(--carbon-shadow), 0 0 40px rgba(139, 92, 246, 0.15);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    animation: slideUp 0.25s ease;
    color: var(--carbon-text);
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px) scale(0.97);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    background: var(--carbon-graphite);
    border-bottom: 1px solid var(--carbon-border);
    color: var(--carbon-text);
    flex-shrink: 0;
  }

  .modal-header h3 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }

  .close-btn {
    background: transparent;
    border: 1px solid var(--carbon-border);
    color: var(--carbon-text-muted);
    font-size: 22px;
    cursor: pointer;
    padding: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    transition: all 0.2s;
  }

  .close-btn:hover {
    background: rgba(139, 92, 246, 0.12);
    border-color: var(--carbon-border-active);
    color: var(--carbon-text);
  }

  .modal-body {
    padding: 20px;
    overflow-y: auto;
    flex: 1;
  }

  .modal-body textarea {
    width: 100%;
    min-height: 140px;
    max-height: 60vh;
    padding: 12px;
    background: var(--carbon-black);
    border: 1px solid var(--carbon-border);
    border-radius: 10px;
    color: var(--carbon-text);
    font-size: 14px;
    font-family: inherit;
    resize: vertical;
    transition: all 0.2s;
    box-sizing: border-box;
  }

  .modal-body textarea::placeholder {
    color: var(--carbon-text-dim);
  }

  .modal-body textarea:focus {
    outline: none;
    border-color: var(--carbon-glow-cyan);
    box-shadow: 0 0 0 3px rgba(34, 211, 238, 0.12);
  }

  .modal-body textarea:disabled {
    opacity: 0.6;
    background: var(--carbon-graphite);
  }

  .success-message {
    margin-top: 10px;
    color: var(--carbon-glow-cyan);
    font-weight: 600;
    text-align: center;
    animation: fadeIn 0.3s ease;
  }

  .modal-footer {
    display: flex;
    gap: 10px;
    padding: 16px 20px;
    background: var(--carbon-graphite);
    border-top: 1px solid var(--carbon-border);
    flex-shrink: 0;
  }

  .modal-footer button {
    flex: 1;
    padding: 10px 16px;
    border-radius: 10px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .cancel-btn {
    background: var(--carbon-graphene);
    border: 1px solid var(--carbon-border);
    color: var(--carbon-text-muted);
  }

  .cancel-btn:hover:not(:disabled) {
    background: var(--carbon-elevated);
    border-color: var(--carbon-border-active);
    color: var(--carbon-text);
  }

  .submit-btn {
    background: linear-gradient(135deg, var(--carbon-glow-purple) 0%, var(--carbon-glow-cyan) 100%);
    border: none;
    color: white;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
  }

  .submit-btn:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 14px rgba(139, 92, 246, 0.4);
  }

  .modal-footer button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Mobile */
  @media (max-width: 480px) {
    .quick-capture-container {
      bottom: 20px;
      left: 20px;
    }

    .quick-capture-btn {
      width: 54px;
      height: 54px;
      font-size: 22px;
    }

    .quick-capture-modal {
      width: calc(100vw - 24px);
      max-height: calc(100vh - 60px);
      border-radius: 12px;
    }

    .modal-header {
      padding: 14px 16px;
    }

    .modal-header h3 {
      font-size: 16px;
    }

    .modal-body {
      padding: 16px;
    }

    .modal-footer {
      padding: 12px 16px;
    }
  }
</style>
