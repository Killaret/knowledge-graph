<script lang="ts">
  import type { Snippet } from "svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  interface Props {
    children?: Snippet;
    class?: string;
  }

  const { children, class: className = "" }: Props = $props();

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);
</script>

<div
  class="cockpit-frame {className}"
  data-testid="cockpit-frame"
  role="img"
  aria-label={t("cockpit.frame.ariaLabel")}
>
  <div class="frame-grid" aria-hidden="true"></div>
  <div class="frame-corners" aria-hidden="true">
    <span class="bolt bolt--tl"></span>
    <span class="bolt bolt--tr"></span>
    <span class="bolt bolt--bl"></span>
    <span class="bolt bolt--br"></span>
  </div>
  <div class="frame-content">
    {@render children?.()}
  </div>
</div>

<style>
  .cockpit-frame {
    position: relative;
    width: 100%;
    height: 100%;
    overflow: hidden;
    border: 6px solid transparent;
    border-radius: 8px;
    background:
      linear-gradient(180deg, rgba(10, 10, 15, 0.72) 0%, rgba(10, 10, 15, 0.88) 100%) padding-box,
      linear-gradient(
          135deg,
          rgba(45, 212, 191, 0.55) 0%,
          rgba(10, 10, 20, 0.8) 25%,
          rgba(45, 212, 191, 0.35) 50%,
          rgba(10, 10, 20, 0.8) 75%,
          rgba(192, 38, 211, 0.5) 100%
        )
        border-box;
    box-shadow:
      /* outer casing depth */
      0 0 0 1px rgba(45, 212, 191, 0.15),
      0 12px 40px rgba(0, 0, 0, 0.7),
      0 0 60px rgba(0, 0, 0, 0.5),
      /* inner screen shadow */ inset 0 0 50px rgba(0, 0, 0, 0.85),
      inset 0 0 16px rgba(45, 212, 191, 0.08);
    pointer-events: none;
  }

  .frame-grid {
    position: absolute;
    inset: 0;
    z-index: 0;
    background-color: rgba(8, 8, 14, 0.65);
    background-image:
      /* grid */
      linear-gradient(rgba(45, 212, 191, 0.035) 1px, transparent 1px),
      linear-gradient(90deg, rgba(45, 212, 191, 0.035) 1px, transparent 1px),
      /* distant stars */
      radial-gradient(circle at 20% 30%, rgba(255, 255, 255, 0.45) 1px, transparent 1.5px),
      radial-gradient(circle at 70% 20%, rgba(45, 212, 191, 0.5) 1px, transparent 1.5px),
      radial-gradient(circle at 40% 80%, rgba(192, 38, 211, 0.45) 1px, transparent 1.5px),
      radial-gradient(circle at 85% 70%, rgba(255, 255, 255, 0.35) 1px, transparent 1.5px),
      radial-gradient(circle at 15% 65%, rgba(45, 212, 191, 0.4) 1px, transparent 1.5px);
    background-size:
      48px 48px,
      48px 48px,
      100% 100%,
      100% 100%,
      100% 100%,
      100% 100%,
      100% 100%;
    pointer-events: none;
  }

  .frame-content {
    position: relative;
    z-index: 1;
    width: 100%;
    height: 100%;
    overflow: hidden;
    border-radius: 4px;
    pointer-events: auto;
  }

  .frame-corners {
    position: absolute;
    inset: 0;
    z-index: 2;
    pointer-events: none;
  }

  .bolt {
    position: absolute;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: radial-gradient(circle at 35% 35%, #5eead4, #0f172a 70%);
    box-shadow:
      inset 0 1px 2px rgba(255, 255, 255, 0.4),
      0 0 6px rgba(45, 212, 191, 0.4);
  }

  .bolt::before,
  .bolt::after {
    content: "";
    position: absolute;
    background: var(--cockpit-accent, rgba(45, 212, 191, 0.5));
    opacity: 0.5;
    border-radius: 1px;
  }

  .bolt::before {
    top: 50%;
    left: 2px;
    width: 6px;
    height: 1px;
    transform: translateY(-50%);
  }

  .bolt::after {
    left: 50%;
    top: 2px;
    width: 1px;
    height: 6px;
    transform: translateX(-50%);
  }

  .bolt--tl {
    top: 5px;
    left: 5px;
  }
  .bolt--tr {
    top: 5px;
    right: 5px;
  }
  .bolt--bl {
    bottom: 5px;
    left: 5px;
  }
  .bolt--br {
    bottom: 5px;
    right: 5px;
  }

  @media (prefers-reduced-motion: reduce) {
    .cockpit-frame {
      transition: none;
    }
  }
</style>
