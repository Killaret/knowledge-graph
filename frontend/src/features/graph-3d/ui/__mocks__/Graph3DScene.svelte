<script lang="ts">
  import { onMount } from "svelte";
  import type { GraphNode, GraphLink } from "$shared/api/graph";

  interface Props {
    nodes?: GraphNode[];
    links?: GraphLink[];
    centerNodeId?: string | null;
    selectedNodeId?: string | null;
    stableRender?: boolean;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
    onNodeDoubleClick?: (node: { id: string; title: string; type?: string }) => void;
    onReady?: () => void;
    onLoadingChange?: (loading: boolean) => void;
    onError?: (message: string) => void;
  }

  const { onReady, onLoadingChange }: Props = $props();

  onMount(() => {
    onReady?.();
    onLoadingChange?.(false);
  });
</script>

<div
  data-testid="graph-3d-scene"
  class="graph-3d-container"
></div>
