<script lang="ts">
  import { goto } from "$app/navigation";
  import { createNote } from "$shared/api/notes";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  let title = $state("");
  let content = $state("");
  let saving = $state(false);
  let error = $state("");

  async function handleSubmit() {
    if (!title.trim()) {
      error = t("noteEditor.titleRequired");
      return;
    }
    saving = true;
    error = "";
    try {
      const note = await createNote({ title, content, metadata: {} });
      await goto(`/notes/${note.id}`);
    } catch (e) {
      error = t("note.createError");
      console.error("Create note error:", e);
    } finally {
      saving = false;
    }
  }
</script>

<h1>{t("note.newTitle")}</h1>

{#if error}
  <p class="error">{error}</p>
{/if}

<form onsubmit={handleSubmit}>
  <input
    type="text"
    name="title"
    placeholder={t("noteEditor.titlePlaceholder")}
    bind:value={title}
    required
  />
  <textarea
    name="content"
    placeholder={t("note.contentPlaceholderWiki")}
    bind:value={content}
    rows="15"
  ></textarea>
  <button type="submit" disabled={saving}
    >{saving ? t("noteEditor.saving") : t("noteEditor.create")}</button
  >
</form>

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
