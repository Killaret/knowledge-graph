<script lang="ts">
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';

  let { href = '/', text = 'Back' }: { href?: string; text?: string } = $props();

  function handleBack() {
    if (browser && window.history.length > 1) {
      window.history.back();
    } else {
      goto(href);
    }
  }
</script>

<button 
  class="back-button" 
  onclick={handleBack}
  aria-label="Go back"
>
  {text}
</button>

<style>
  .back-button {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    background: #6b7280;
    color: white;
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 6px;
    text-decoration: none;
    font-size: 0.9rem;
    cursor: pointer;
    transition: all 0.2s ease;
    margin-bottom: 1rem;
  }

  .back-button:hover {
    background: #4b5563;
    transform: translateY(-1px);
  }

  .back-button:active {
    transform: translateY(0);
  }

  .back-button::before {
    content: "«";
    font-weight: bold;
  }
</style>
