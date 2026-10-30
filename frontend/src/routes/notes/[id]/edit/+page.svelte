<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { getNote, updateNote } from "$shared/api/notes";
  import type { Note } from "$shared/api/notes";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  let note: Note | null = $state(null);
  let title = $state("");
  let content = $state("");
  let saving = $state(false);
  let error = $state("");
  let loading = $state(true);

  function getRouteId(): string {
    const id = $page.params.id;
    if (!id) throw new Error("Missing route parameter: id");
    return id;
  }

  onMount(async () => {
    try {
      const id = getRouteId();
      note = await getNote(id);
      title = note.title;
      content = note.content;
    } catch {
      error = t("note.notFoundShort");
    } finally {
      loading = false;
    }
  });

  async function handleSubmit(event: Event) {
    event.preventDefault();
    const id = getRouteId();
    if (!title.trim()) {
      error = t("noteEditor.titleRequired");
      return;
    }
    saving = true;
    error = "";
    try {
      await updateNote(id, { title, content });
      goto(`/notes/${id}`);
    } catch {
      error = t("note.updateError");
    } finally {
      saving = false;
    }
  }
</script>

<h1>{t("noteEditor.titleEdit")}</h1>

{#if loading}
  <p>{t("note.loading")}</p>
{:else if error}
  <p class="error">{error}</p>
{:else}
  <form onsubmit={handleSubmit}>
    <input
      type="text"
      name="title"
      placeholder={t("noteEditor.titlePlaceholder")}
      bind:value={title}
      required
    />
    <textarea name="content" bind:value={content} rows="15"></textarea>
    <button type="submit" disabled={saving}
      >{saving ? t("noteEditor.saving") : t("noteEditor.update")}</button
    >
  </form>
{/if}

<style>
  input,
  textarea {
    width: 100%;
    padding: 0.5rem;
    margin-bottom: 1rem;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-family: inherit;
  }
  .error {
    color: red;
  }
</style>
