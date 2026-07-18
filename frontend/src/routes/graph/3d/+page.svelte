<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";

  onMount(() => {
    // 3D functionality frozen for v1 - redirecting to 2D graph
    setTimeout(() => {
      goto("/graph");
    }, 500);
  });
</script>

<div class="page">
  <div class="center">
    <div class="frozen-notice">
      <h2>3D Graph Feature Frozen</h2>
      <p>
        The 3D graph functionality has been temporarily frozen for version 1.0
        to improve stability and reduce maintenance overhead.
      </p>
      <p>You will be automatically redirected to the 2D graph view.</p>
      <div class="spinner"></div>
    </div>
  </div>
</div>

<style>
  .page {
    position: relative;
    width: 100%;
    height: 100vh;
    overflow: hidden;
    background: #050510;
  }

  .center {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: white;
    gap: 16px;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255, 255, 255, 0.3);
    border-top-color: #88aaff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
