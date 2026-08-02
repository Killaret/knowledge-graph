<script lang="ts">
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  interface Props {
    visible: boolean;
    hovered: boolean;
  }

  const { visible, hovered }: Props = $props();

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);
</script>

{#if visible}
  <div
    class="singularity-drop-zone"
    class:hovered
    role="region"
    aria-label={t("graph.singularity.ariaLabel")}
    data-testid="singularity-drop-zone"
  >
    <div class="singularity-core"></div>
    <div class="singularity-ring ring-1"></div>
    <div class="singularity-ring ring-2"></div>
    <div class="singularity-ring ring-3"></div>
    <span class="singularity-label">{t("graph.singularity.label")}</span>
  </div>
{/if}

<style>
  .singularity-drop-zone {
    position: fixed;
    right: 24px;
    bottom: 24px;
    width: 120px;
    height: 120px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: none;
    z-index: 120;
    opacity: 0.65;
    transition:
      opacity 0.25s ease,
      transform 0.25s ease;
  }

  .singularity-drop-zone.hovered {
    opacity: 1;
    transform: scale(1.1);
  }

  .singularity-core {
    position: absolute;
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: radial-gradient(circle at 30% 30%, #2a0a3a, #000000);
    box-shadow: 0 0 24px 4px rgba(168, 85, 247, 0.35);
    transition: box-shadow 0.25s ease;
  }

  .singularity-drop-zone.hovered .singularity-core {
    box-shadow: 0 0 40px 8px rgba(192, 132, 252, 0.55);
  }

  .singularity-ring {
    position: absolute;
    border-radius: 50%;
    border: 1px solid rgba(168, 85, 247, 0.35);
    animation: spin 6s linear infinite;
  }

  .ring-1 {
    width: 72px;
    height: 72px;
    animation-duration: 4s;
  }

  .ring-2 {
    width: 96px;
    height: 96px;
    animation-duration: 6s;
    border-style: dashed;
  }

  .ring-3 {
    width: 120px;
    height: 120px;
    animation-duration: 8s;
    animation-direction: reverse;
    opacity: 0.5;
  }

  .singularity-drop-zone.hovered .singularity-ring {
    border-color: rgba(192, 132, 252, 0.7);
  }

  .singularity-label {
    position: absolute;
    bottom: -22px;
    font-size: 11px;
    color: rgba(255, 255, 255, 0.65);
    white-space: nowrap;
    text-shadow: 0 1px 4px rgba(0, 0, 0, 0.7);
    transition: color 0.25s ease;
  }

  .singularity-drop-zone.hovered .singularity-label {
    color: rgba(255, 255, 255, 0.9);
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
