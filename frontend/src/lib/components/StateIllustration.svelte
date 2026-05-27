<script lang="ts">
  export type StateIllustrationType = 'empty' | 'error' | '404' | 'offline' | 'no-links' | 'no-results';
  let { type = 'empty' }: { type?: StateIllustrationType } = $props();
  const currentType = $derived(type);

  const labels: Record<StateIllustrationType, string> = {
    empty: 'Empty state illustration',
    error: 'Error state illustration',
    '404': '404 illustration',
    offline: 'Offline illustration',
    'no-links': 'No links illustration',
    'no-results': 'No results illustration'
  };

</script>

<div class="state-illustration" role="img" aria-label={labels[currentType]}>
  <svg viewBox="0 0 240 180" preserveAspectRatio="xMidYMid meet" aria-hidden="true" focusable="false">
    <rect x="0" y="0" width="240" height="180" rx="24" fill="var(--color-surface, #000)" />

    {#if currentType === 'empty'}
      <g fill="#fff" opacity="0.9">
        <circle cx="34" cy="32" r="2.2" />
        <circle cx="210" cy="40" r="2" />
        <circle cx="180" cy="120" r="1.8" />
        <circle cx="56" cy="100" r="1.8" />
        <circle cx="122" cy="152" r="2.2" />
      </g>
      <path d="M28 66L34 66M31 62L31 70" stroke="#fff" stroke-width="1.2" stroke-linecap="round" />
      <path d="M72 18L76 24M76 18L72 24" stroke="#fff" stroke-width="1.2" stroke-linecap="round" />
      <path d="M182 86L186 86M184 82L184 90" stroke="#fff" stroke-width="1.2" stroke-linecap="round" />
      <circle cx="148" cy="108" r="10" fill="var(--color-primary, #7c3aed)" opacity="0.24" />
      <path d="M40 140C90 110 150 110 200 80" fill="none" stroke="var(--color-primary, #7c3aed)" stroke-width="2" opacity="0.7" />
      <circle cx="62" cy="126" r="6" fill="var(--color-primary, #7c3aed)" opacity="0.3" />
      <circle cx="176" cy="50" r="6" fill="var(--color-info, #14b8a6)" opacity="0.35" />
    {:else if currentType === 'error'}
      <polygon points="56,148 120,32 184,148" fill="var(--color-primary, #7c3aed)" stroke="#fff" stroke-width="3" />
      <path d="M82 84L96 104L90 114L108 129" fill="none" stroke="var(--color-danger, #ef4444)" stroke-width="7" stroke-linecap="round" />
      <path d="M88 74L118 108" stroke="var(--color-danger, #ef4444)" stroke-width="4" stroke-linecap="round" />
      <circle cx="120" cy="150" r="9" fill="var(--color-info, #14b8a6)" opacity="0.6" />
      <path d="M70 40L74 50M80 30L84 40M160 46L156 54" stroke="#fff" stroke-width="1.5" stroke-linecap="round" />
      <rect x="48" y="148" width="144" height="10" rx="5" fill="rgba(255,255,255,0.08)" />
    {:else if currentType === '404'}
      <g opacity="0.95">
        <circle cx="72" cy="86" r="22" fill="var(--color-info, #14b8a6)" />
        <circle cx="72" cy="86" r="14" fill="var(--color-surface, #000)" />
      </g>
      <path d="M100 72L220 34L220 66L100 104Z" fill="var(--color-primary, #7c3aed)" opacity="0.9" />
      <path d="M100 72C132 60 170 52 220 48" fill="none" stroke="var(--color-info, #14b8a6)" stroke-width="12" stroke-linecap="round" opacity="0.85" />
      <circle cx="175" cy="110" r="8" fill="#fff" opacity="0.85" />
      <circle cx="48" cy="130" r="2" fill="#fff" opacity="0.9" />
      <circle cx="202" cy="30" r="2" fill="#fff" opacity="0.9" />
    {:else if currentType === 'offline'}
      <line x1="28" y1="50" x2="212" y2="50" stroke="var(--color-info, #14b8a6)" stroke-width="10" stroke-linecap="round" />
      <line x1="28" y1="130" x2="212" y2="130" stroke="var(--color-info, #14b8a6)" stroke-width="10" stroke-linecap="round" />
      <line x1="118" y1="42" x2="130" y2="62" stroke="var(--color-danger, #ef4444)" stroke-width="10" stroke-linecap="round" />
      <line x1="138" y1="118" x2="150" y2="138" stroke="var(--color-danger, #ef4444)" stroke-width="10" stroke-linecap="round" />
      <path d="M112 62L128 78" stroke="var(--color-danger, #ef4444)" stroke-width="5" stroke-linecap="round" />
      <circle cx="120" cy="90" r="10" fill="var(--color-primary, #7c3aed)" opacity="0.35" />
      <path d="M32 60L24 48M216 60L224 48" stroke="#fff" stroke-width="1.5" stroke-linecap="round" />
    {:else if currentType === 'no-links'}
      <circle cx="120" cy="84" r="28" fill="var(--color-primary, #7c3aed)" />
      <circle cx="120" cy="84" r="14" fill="var(--color-surface, #000)" />
      <circle cx="120" cy="84" r="6" fill="var(--color-info, #14b8a6)" />
      <circle cx="82" cy="46" r="3" fill="#fff" opacity="0.9" />
      <path d="M55 133C96 110 145 110 182 143" fill="none" stroke="var(--color-info, #14b8a6)" stroke-width="3" opacity="0.55" stroke-dasharray="6 5" />
      <path d="M106 132C150 90 158 72 192 60" fill="none" stroke="var(--color-primary, #7c3aed)" stroke-width="2" opacity="0.4" />
    {:else if currentType === 'no-results'}
      <g transform="translate(36, 36)">
        <rect x="0" y="30" width="96" height="26" rx="13" fill="var(--color-primary, #7c3aed)" />
        <rect x="74" y="-4" width="22" height="46" rx="11" fill="var(--color-surface, #000)" />
        <path d="M20 32L10 18L12 16L22 30Z" fill="var(--color-info, #14b8a6)" opacity="0.8" />
        <line x1="22" y1="55" x2="10" y2="85" stroke="var(--color-info, #14b8a6)" stroke-width="8" stroke-linecap="round" />
        <line x1="42" y1="55" x2="54" y2="85" stroke="var(--color-info, #14b8a6)" stroke-width="8" stroke-linecap="round" />
        <path d="M64 18L90 6" stroke="#fff" stroke-width="4" stroke-linecap="round" />
      </g>
      <circle cx="190" cy="40" r="2" fill="#fff" opacity="0.9" />
      <circle cx="196" cy="66" r="2" fill="#fff" opacity="0.82" />
      <circle cx="172" cy="80" r="1.8" fill="#fff" opacity="0.82" />
    {/if}
  </svg>
</div>

<style>
  .state-illustration {
    max-width: 320px;
    margin: 0 auto 1rem;
    display: grid;
    place-items: center;
  }

  svg {
    width: 100%;
    height: auto;
    display: block;
  }
</style>
