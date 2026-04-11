<script lang="ts">
  import KeyboardShortcuts from '$lib/components/KeyboardShortcuts.svelte';
  
  let { children } = $props();
</script>

{@render children()}

<KeyboardShortcuts />