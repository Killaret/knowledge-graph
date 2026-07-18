<script lang="ts">
  const {
    mode,
    focusMode,
    onReset,
    onSearch,
    onToggleMode,
    onToggleFocus,
  }: {
    mode: "normal" | "focus";
    focusMode: boolean;
    onReset: () => void;
    onSearch: () => void;
    onToggleMode: () => void;
    onToggleFocus: () => void;
  } = $props();
</script>

<div
  data-testid="graph-controls"
  style="position: absolute; top: 56px; left: 16px; display: flex; flex-direction: column; gap: 8px; z-index: 50;"
>
  <button
    data-testid="graph-controls-reset"
    onclick={onReset}
    title="Reset view"
    style="background: rgba(0,0,0,0.6); border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; padding: 8px; color: white; cursor: pointer; font-size: 14px; transition: all 0.2s;"
  >
    🔄
  </button>
  <button
    data-testid="graph-controls-search"
    onclick={onSearch}
    title="Search (F)"
    style="background: rgba(0,0,0,0.6); border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; padding: 8px; color: white; cursor: pointer; font-size: 14px; transition: all 0.2s;"
  >
    🔍
  </button>
  <button
    data-testid="graph-controls-mode"
    onclick={onToggleMode}
    title="Toggle mode"
    style="background: rgba(0,0,0,0.6); border: 1px solid {mode === 'focus'
      ? 'rgba(139, 92, 246, 0.8)'
      : 'rgba(255,255,255,0.2)'}; border-radius: 8px; padding: 8px; color: {mode ===
    'focus'
      ? '#a78bfa'
      : 'white'}; cursor: pointer; font-size: 14px; transition: all 0.2s;"
  >
    {mode === "focus" ? "👁" : "⚡"}
  </button>
  <button
    data-testid="graph-controls-focus"
    onclick={onToggleFocus}
    title="Toggle focus mode"
    style="background: rgba(0,0,0,0.6); border: 1px solid {focusMode
      ? 'rgba(139, 92, 246, 0.8)'
      : 'rgba(255,255,255,0.2)'}; border-radius: 8px; padding: 8px; color: {focusMode
      ? '#a78bfa'
      : 'white'}; cursor: pointer; font-size: 14px; transition: all 0.2s;"
  >
    {focusMode ? "🎯" : "🔘"}
  </button>
</div>
