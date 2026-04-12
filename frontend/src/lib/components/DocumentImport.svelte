<script lang="ts">
  import { fade, scale } from 'svelte/transition';
  
  const { 
    isOpen = false,
    onClose,
    onImport
  }: {
    isOpen: boolean;
    onClose: () => void;
    onImport: (files: FileList | null, url: string) => void;
  } = $props();
  
  let importUrl = $state('');
  let isImporting = $state(false);
  let progress = $state(0);
  
  function handleBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget && !isImporting) {
      onClose();
    }
  }
  
  function handleFileDrop(event: DragEvent) {
    event.preventDefault();
    const files = event.dataTransfer?.files;
    if (files && files.length > 0) {
      startImport(files, '');
    }
  }
  
  function handleFileSelect(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      startImport(input.files, '');
    }
  }
  
  function handleUrlImport() {
    if (importUrl.trim()) {
      startImport(null, importUrl.trim());
    }
  }
  
  async function startImport(files: FileList | null, url: string) {
    isImporting = true;
    progress = 0;
    
    // Мок импорта с прогрессом
    const interval = setInterval(() => {
      progress += Math.random() * 15;
      if (progress >= 100) {
        progress = 100;
        clearInterval(interval);
        setTimeout(() => {
          onImport(files, url);
          isImporting = false;
          progress = 0;
          importUrl = '';
          onClose();
        }, 500);
      }
    }, 300);
  }
  
  function preventDefault(event: Event) {
    event.preventDefault();
  }
</script>

{#if isOpen}
  <div 
    class="modal-backdrop" 
    onclick={handleBackdropClick}
    onkeydown={(e) => e.key === 'Escape' && !isImporting && onClose()}
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    transition:fade={{ duration: 200 }}
  >
    <div class="modal-content" transition:scale={{ duration: 200, start: 0.95 }}>
      <button class="close-btn" onclick={onClose} disabled={isImporting} aria-label="Закрыть">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
      
      <h2 class="modal-title">Импорт документа</h2>
      
      {#if isImporting}
        <div class="progress-section">
          <div class="progress-bar">
            <div class="progress-fill" style="width: {progress}%"></div>
          </div>
          <p class="progress-text">Обработка документа... {Math.round(progress)}%</p>
          <p class="progress-subtext">Разбиение на заметки и создание связей</p>
        </div>
      {:else}
        <div class="upload-section"
          ondragover={preventDefault}
          ondrop={handleFileDrop}
          role="region"
          aria-label="Зона загрузки файлов"
        >
          <div class="upload-icon">
            <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
              <polyline points="17,8 12,3 7,8"></polyline>
              <line x1="12" y1="3" x2="12" y2="15"></line>
            </svg>
          </div>
          <p class="upload-text">Перетащите PDF или файл сюда</p>
          <p class="upload-subtext">или</p>
          <label class="file-input-label">
            <input 
              type="file" 
              accept=".pdf,.txt,.md,.doc,.docx"
              onchange={handleFileSelect}
              class="file-input"
            />
            Выбрать файл
          </label>
        </div>
        
        <div class="url-section">
          <div class="divider">
            <span>или URL</span>
          </div>
          <div class="url-input-group">
            <input 
              type="url" 
              bind:value={importUrl}
              placeholder="https://example.com/document"
              class="url-input"
              onkeydown={(e) => e.key === 'Enter' && handleUrlImport()}
            />
            <button class="url-btn" onclick={handleUrlImport} disabled={!importUrl.trim()}>
              Импорт
            </button>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2000;
  }

  .modal-content {
    position: relative;
    width: 90%;
    max-width: 500px;
    background: linear-gradient(145deg, #0a1a3a, #1a2a4a);
    border: 1px solid rgba(255, 221, 136, 0.2);
    border-radius: 16px;
    padding: 32px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  }

  .close-btn {
    position: absolute;
    top: 16px;
    right: 16px;
    background: none;
    border: none;
    color: #88aacc;
    cursor: pointer;
    padding: 8px;
    border-radius: 8px;
    transition: all 0.2s;
  }

  .close-btn:hover:not(:disabled) {
    color: #ffdd88;
    background: rgba(255, 221, 136, 0.1);
  }

  .close-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .modal-title {
    margin: 0 0 24px 0;
    color: #ffdd88;
    font-size: 1.5rem;
    font-weight: 600;
    text-align: center;
  }

  .upload-section {
    border: 2px dashed rgba(136, 170, 204, 0.3);
    border-radius: 12px;
    padding: 40px 20px;
    text-align: center;
    transition: all 0.2s;
  }

  .upload-section:hover {
    border-color: rgba(255, 221, 136, 0.5);
    background: rgba(255, 221, 136, 0.05);
  }

  .upload-icon {
    color: #88aacc;
    margin-bottom: 16px;
  }

  .upload-text {
    color: #e0e0e0;
    font-size: 1.1rem;
    margin: 0 0 8px 0;
  }

  .upload-subtext {
    color: #88aacc;
    margin: 0 0 16px 0;
  }

  .file-input {
    display: none;
  }

  .file-input-label {
    display: inline-block;
    padding: 10px 20px;
    background: linear-gradient(135deg, #ffdd88 0%, #ffaa44 100%);
    color: #0a1a3a;
    border-radius: 8px;
    cursor: pointer;
    font-weight: 600;
    transition: all 0.2s;
  }

  .file-input-label:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(255, 170, 68, 0.4);
  }

  .url-section {
    margin-top: 24px;
  }

  .divider {
    display: flex;
    align-items: center;
    margin: 20px 0;
    color: #88aacc;
    font-size: 0.9rem;
  }

  .divider::before,
  .divider::after {
    content: '';
    flex: 1;
    height: 1px;
    background: rgba(136, 170, 204, 0.2);
  }

  .divider span {
    padding: 0 16px;
  }

  .url-input-group {
    display: flex;
    gap: 8px;
  }

  .url-input {
    flex: 1;
    padding: 12px 16px;
    background: rgba(10, 26, 58, 0.6);
    border: 1px solid rgba(136, 170, 204, 0.2);
    border-radius: 8px;
    color: #e0e0e0;
    font-size: 0.95rem;
  }

  .url-input:focus {
    outline: none;
    border-color: rgba(255, 221, 136, 0.5);
  }

  .url-input::placeholder {
    color: rgba(136, 170, 204, 0.5);
  }

  .url-btn {
    padding: 12px 20px;
    background: linear-gradient(135deg, #44aaff 0%, #3388dd 100%);
    color: white;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    font-weight: 600;
    transition: all 0.2s;
  }

  .url-btn:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(68, 170, 255, 0.4);
  }

  .url-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .progress-section {
    text-align: center;
    padding: 20px;
  }

  .progress-bar {
    width: 100%;
    height: 8px;
    background: rgba(136, 170, 204, 0.2);
    border-radius: 4px;
    overflow: hidden;
    margin-bottom: 16px;
  }

  .progress-fill {
    height: 100%;
    background: linear-gradient(90deg, #ffdd88, #ffaa44);
    border-radius: 4px;
    transition: width 0.3s ease;
  }

  .progress-text {
    color: #ffdd88;
    font-size: 1.1rem;
    font-weight: 600;
    margin: 0 0 8px 0;
  }

  .progress-subtext {
    color: #88aacc;
    font-size: 0.9rem;
    margin: 0;
  }
</style>
