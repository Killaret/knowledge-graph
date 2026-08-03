<script lang="ts">
  import { GraphMode } from "$entities";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  const {
    mode,
    focusMode,
    onReset,
    onSearch,
    onToggleMode,
    onToggleFocus,
  }: {
    mode: GraphMode;
    focusMode?: boolean;
    onReset: () => void;
    onSearch: () => void;
    onToggleMode: () => void;
    onToggleFocus: () => void;
  } = $props();

  const graphMode = $derived(mode);
  const focusGraphMode = $derived(GraphMode.fromFocus(focusMode ?? graphMode.isFocus));
</script>

<div
  data-testid="graph-controls"
  class="graph-controls"
  style="display: flex; flex-direction: row; gap: 4px; align-items: center;"
>
  <button
    data-testid="graph-controls-reset"
    onclick={onReset}
    title={t("graphControls.resetView")}
    style="background: rgba(0,0,0,0.6); border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; padding: 8px; color: white; cursor: pointer; font-size: 14px; transition: all 0.2s;"
  >
    🔄
  </button>
  <button
    data-testid="graph-controls-search"
    onclick={onSearch}
    title={t("graphControls.search")}
    style="background: rgba(0,0,0,0.6); border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; padding: 8px; color: white; cursor: pointer; font-size: 14px; transition: all 0.2s;"
  >
    🎯
  </button>
  <button
    data-testid="graph-controls-mode"
    onclick={onToggleMode}
    title={graphMode.label}
    style="background: rgba(0,0,0,0.6); border: 1px solid {graphMode.borderColor}; border-radius: 8px; padding: 8px; color: {graphMode.textColor}; cursor: pointer; font-size: 14px; transition: all 0.2s;"
  >
    {graphMode.icon}
  </button>
  <button
    data-testid="graph-controls-focus"
    onclick={onToggleFocus}
    title={focusGraphMode.focusLabel}
    style="background: rgba(0,0,0,0.6); border: 1px solid {focusGraphMode.borderColor}; border-radius: 8px; padding: 8px; color: {focusGraphMode.textColor}; cursor: pointer; font-size: 14px; transition: all 0.2s;"
  >
    {focusGraphMode.focusIcon}
  </button>
</div>
