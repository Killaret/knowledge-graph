<script lang="ts">
  import Button from './Button.svelte';
  
  const { variant, type, disabled, onClick, dataTestid, ariaLabel, label } = $props<{
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
    type?: 'button' | 'submit' | 'reset';
    disabled?: boolean;
    onClick?: (e: MouseEvent) => void;
    dataTestid?: string;
    ariaLabel?: string;
    label: string;
  }>();
</script>

<Button {variant} {type} {disabled} {onClick} data-testid={dataTestid} aria-label={ariaLabel}>
  {label}
</Button>
