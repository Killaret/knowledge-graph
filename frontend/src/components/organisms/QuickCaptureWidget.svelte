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
      await createNote({
        title,
        content,
        type: CelestialBody.DUST.type,
      });

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
    right: 30px;
    z-index: 1000;
  }

  .quick-capture-btn {
    width: 60px;
    height: 60px;
    border-radius: 50%;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border: none;
    color: white;
    font-size: 28px;
    cursor: pointer;
    box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
    transition: all 0.3s ease;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .quick-capture-btn:hover {
    transform: scale(1.1);
    box-shadow: 0 6px 20px rgba(102, 126, 234, 0.6);
  }

  /* Backdrop overlay */
  .quick-capture-backdrop {
    position: fixed;
    inset: 0;
    z-index: 1001;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.4);
    backdrop-filter: blur(4px);
    animation: fadeIn 0.2s ease;
  }

  .quick-capture-modal {
    width: min(420px, calc(100vw - 40px));
    max-height: calc(100vh - 100px);
    background: white;
    border-radius: 16px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    animation: slideUp 0.25s ease;
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
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    flex-shrink: 0;
  }

  .modal-header h3 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }

  .close-btn {
    background: none;
    border: none;
    color: white;
    font-size: 24px;
    cursor: pointer;
    padding: 0;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    transition: background 0.2s;
  }

  .close-btn:hover {
    background: rgba(255, 255, 255, 0.2);
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
    border: 2px solid #e2e8f0;
    border-radius: 8px;
    font-size: 14px;
    font-family: inherit;
    resize: vertical;
    transition: border-color 0.2s;
    box-sizing: border-box;
  }

  .modal-body textarea:focus {
    outline: none;
    border-color: #667eea;
  }

  .modal-body textarea:disabled {
    background: #f7fafc;
  }

  .success-message {
    margin-top: 10px;
    color: #48bb78;
    font-weight: 600;
    text-align: center;
    animation: fadeIn 0.3s ease;
  }

  .modal-footer {
    display: flex;
    gap: 10px;
    padding: 16px 20px;
    background: #f7fafc;
    border-top: 1px solid #e2e8f0;
    flex-shrink: 0;
  }

  .modal-footer button {
    flex: 1;
    padding: 10px 16px;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .cancel-btn {
    background: white;
    border: 2px solid #e2e8f0;
    color: #4a5568;
  }

  .cancel-btn:hover:not(:disabled) {
    background: #f7fafc;
    border-color: #cbd5e0;
  }

  .submit-btn {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border: none;
    color: white;
  }

  .submit-btn:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  }

  .modal-footer button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  /* Mobile */
  @media (max-width: 480px) {
    .quick-capture-container {
      bottom: 20px;
      right: 20px;
    }

    .quick-capture-btn {
      width: 52px;
      height: 52px;
      font-size: 24px;
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
