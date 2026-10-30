<script lang="ts">
  import Modal from "$components/atoms/Modal.svelte";

  const { open, title, onClose, content } = $props<{
    open: boolean;
    title: string;
    onClose?: () => void;
    content?: string;
  }>();
</script>

<Modal {open} {title} {onClose}>
  {content}
</Modal>
