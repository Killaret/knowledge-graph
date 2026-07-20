<script lang="ts">
  /**
   * Test index page
   */
  import GraphCanvas from "$components/organisms/GraphCanvas.svelte";
  import type { GraphLink } from "$shared/api/graph";

  const nodes = [{ id: "1", title: "Test Star", type: "star" }];
  const links: GraphLink[] = [];
</script>

<div class="test-container">
  <h1>Visual Test Page</h1>
  <div class="canvas-wrapper">
    <GraphCanvas {nodes} {links} />
  </div>
</div>

<style>
  .test-container {
    padding: 20px;
    background: radial-gradient(ellipse at 50% 100%, #1a0505 0%, #000 80%);
    min-height: 100vh;
    color: white;
  }

  .canvas-wrapper {
    width: 600px;
    height: 400px;
    border: 2px solid rgba(255, 255, 255, 0.2);
    border-radius: 8px;
  }
</style>
