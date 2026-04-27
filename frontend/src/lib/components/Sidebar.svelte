<script lang="ts">
	// ==========================================
	// CONTEXT CONTROL CENTER (CCC) - ЗАРЕЗЕРВИРОВАНО
	// ==========================================
	// Планируемые props (будут активированы в v2.0):
	// export let collapsed = false;
	// export let activeFilters = [];
	// export let selectedGroups = [];
	// export let savedSearches = [];
	// ...
</script>

<div class="sidebar-placeholder" data-testid="sidebar-placeholder">
	<!--
    ==========================================
    ЗАРЕЗЕРВИРОВАНО ДЛЯ БУДУЩЕЙ НАВИГАЦИИ (v2.0)
    ==========================================
    Назначение: Центр управления контекстом (Context Control Center)
    Размещение: узкая колонка (250–300px) слева, видимая на всех основных страницах

    Планируемые модули (сверху вниз):

    1. ГРУППЫ ЗАМЕТОК (Projects/Folders)
       - Ручная группировка заметок в папки/проекты
       - Создание, переименование, удаление групп
       - Drag-and-drop заметок в группы

    2. ДИНАМИЧЕСКИЕ КЛАСТЕРЫ (Smart Clusters)
       - Автоматически формируемые группы на основе алгоритмов
       - Семантическая близость, общие связи
       - Возможность "закрепить" кластер как статическую группу

    3. ГИБКИЕ ФИЛЬТРЫ (Advanced Filters)
       - По типу: TypeSelector (перенос или дублирование)
       - По тегам: TagSelector
       - По ключевым словам: вкл/выкл NLP, конкретные keywords
       - По "маркам": пользовательские метки/ярлыки
       - Комбинация условий с логикой И/ИЛИ

    4. СОХРАНЁННЫЕ ПОИСКИ (Saved Searches)
       - Закреплённые комбинации фильтров
       - Быстрый доступ к частым поискам

    Интеграция:
    - Текущие элементы (строка поиска, быстрые фильтры) остаются сверху
    - Левая панель берёт на себя "аналитические" сценарии
    ==========================================
  -->
	<div class="sidebar-content-hidden">Navigation Panel (v2.0)</div>
</div>

<style>
	.sidebar-placeholder {
		width: 0; /* Сейчас скрыта - резервирование для будущего */
		overflow: hidden;
		background: var(--color-surface);
		border-right: 1px solid var(--color-border);
		transition: width 0.3s ease;
		flex-shrink: 0;
		/* Будущие стили для активации:
    width: 280px;
    padding: 1rem;
    */
	}

	.sidebar-content-hidden {
		display: none; /* Скрыт до активации */
		padding: 1rem;
		font-size: 0.875rem;
		color: var(--color-text-secondary);
	}
</style>
